package converter

import (
	// stdlib
	"bufio"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Task represents a single task received via NATS.
type Task struct {
	Name       string
	InputFile  string
	OutputFile string

	// Filed in conversion.
	totalFrames int

	// Initial calculation state information.
	previousOutput           string
	gotInput                 bool
	gotDuration              bool
	gotTimeOrFPSParsingError bool

	// After totalFrames will be filled we will use these variables
	// to work with output.
	gotFrame bool

	// File info.
	duration string
	fps      string
}

// Convert launches conversion procedure. Should be launched in separate
// goroutine.
func (t *Task) Convert() {
	log.Printf("Starting conversion task: %+v\n", t)
	currentlyRunningMutex.Lock()
	currentlyRunning++
	currentlyRunningMutex.Unlock()

	defer func() {
		currentlyRunningMutex.Lock()
		currentlyRunning--
		currentlyRunningMutex.Unlock()
	}()

	ffmpegCmd := exec.Command(ffmpegPath, "-i", t.InputFile, "-c:v", "libx264", "-b:v", "1000k", "-c:a", "aac", "-f", "mp4", t.OutputFile, "-y")
	stderr, err := ffmpegCmd.StderrPipe()
	if err != nil {
		log.Fatalln("Error while preparing to redirect ffmpeg's stderr:", err.Error())
	}
	stderrScanner := bufio.NewScanner(stderr)
	stderrScanner.Split(bufio.ScanWords)

	// We will check state every 500ms.
	go func() {
		checkTick := time.NewTicker(time.Millisecond * 500)
		err1 := ffmpegCmd.Start()
		if err1 != nil {
			log.Fatalln("Failed to start ffmpeg:", err1.Error())
		}

		for range checkTick.C {
			// Should we shutdown immediately?
			shouldShutdownMutex.Lock()
			shouldWeStop := shouldShutdown
			shouldShutdownMutex.Unlock()

			if shouldWeStop {
				log.Println("Killing converter goroutine...")
				err := ffmpegCmd.Process.Kill()
				if err != nil {
					log.Println("ERROR: failed to kill ffmpeg process:", err.Error())
				}
				ffmpegCmd.Process.Release()
				break
			}
		}

		log.Println("Child ffmpeg process killed")
	}()

	// Read output.
	for {
		// Should we shutdown immediately?
		shouldShutdownMutex.Lock()
		shouldWeStop := shouldShutdown
		shouldShutdownMutex.Unlock()
		if shouldWeStop {
			break
		}

		proceeding := stderrScanner.Scan()
		if proceeding == false {
			break
		}
		//log.Println(stderrScanner.Text())
		t.workWithOutput(stderrScanner.Text())
	}

	log.Println("Stopped reading ffmpeg output")
}

// Printing progress for this task.
func (t *Task) workWithOutput(output string) {
	// Do nothing if we have empty output string or if we're not ready.
	if output == "" || t.gotTimeOrFPSParsingError {
		return
	}

	// If we have totalFrames defined, which is the very final state
	// for calculations below, we should work with output in different
	// manner :)
	if t.totalFrames != 0 {
		// We should look for current frame count.
		// If we have "frame=" there - then output for next function
		// call will be current frame count.
		if strings.Contains(output, "frame=") {
			t.gotFrame = true
			// If we have only "frame" here - then actual current frame
			// will be in next output. Otherwise we should fix output
			// to contain only actual current frame.
			// We have ASCII here, not runes, len() is fine.
			if len(output) == 6 {
				// Current frame will be in next output.
				return
			} else {
				output = strings.Split(output, "frame=")[1]
			}
		}

		// ... which we should properly use.
		if t.gotFrame {
			currentFrame, err := strconv.Atoi(output)
			if err != nil {
				log.Println("Failed to convert current frame value to int ("+output+"):", err.Error())
				t.gotFrame = false
				return
			}
			percentage := currentFrame / int(t.totalFrames/100)
			// What if... we mistaken with totalFrames prediction?
			if percentage > 100 {
				percentage = 100
			}
			os.Stdout.Write([]byte("\rConverting " + t.InputFile + ": " + strconv.Itoa(percentage) + "% done (" + output + " frame of " + strconv.Itoa(t.totalFrames) + ")"))

			// ... and reset it's state so next "frame=" will be the
			// next stop.
			t.gotFrame = false
		}

		return
	}

	// We got input keyword. Next function runs will look for duration.
	if output == "Input" {
		t.gotInput = true
		return
	}

	if t.gotInput && output == "Duration:" {
		t.gotDuration = true
		return
	}

	if t.gotDuration && t.duration == "" {
		t.duration = output
		log.Println("File duration:", t.duration)
		return
	}

	if t.duration != "" && output == "fps," {
		t.fps = t.previousOutput
		log.Println("Got FPS value:", t.fps)

		// Calculate total frames approximately, because even if ffmpeg
		// writes that there is 29.97 fps, it actually might be something
		// like 29.971872638217638216.
		// BTW, this is a duration, not a time, and to avoid all
		// kind of type pr0n we will just fix gathered duration to
		// be parsable by time.ParseDuration()
		fileDuration := strings.Replace(t.duration, ":", "h", 1)
		fileDuration = strings.Replace(fileDuration, ":", "m", 1)
		fileDuration = strings.Replace(fileDuration, ".", "s", 1)
		fileDuration = strings.Replace(fileDuration, ",", "", 1)
		fileDuration += "ms"
		totalTime, err := time.ParseDuration(fileDuration)
		log.Println("Got file duration parsed:", totalTime)
		seconds := totalTime.Seconds()
		if err != nil {
			log.Println("ERROR: failed to parse video file total time value. No progress output will be produced!")
			t.gotTimeOrFPSParsingError = true
		}

		log.Println("Got file duration in seconds:", seconds)
		fps, err1 := strconv.ParseFloat(t.fps, 64)
		if err1 != nil {
			log.Println("ERROR: failed to parse frames per second value: '" + t.fps + "', no progress output will be produced!")
			t.gotTimeOrFPSParsingError = true
		}
		// We don't mind to loose 1 or 2 fps from total fps counter,
		// yea? :)
		t.totalFrames = int(float64(seconds) * fps)
		log.Println("Total frames calculated:", t.totalFrames)
	}

	// Save previous output for good unless we have everything we need
	// to print progress.
	if t.totalFrames == 0 {
		t.previousOutput = output
	}
}
