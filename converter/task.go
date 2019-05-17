package converter

import (
	// stdlib
	"bufio"
	"log"
	"os/exec"
	"time"
)

// Task represents a single task received via NATS.
type Task struct {
	Name       string
	InputFile  string
	OutputFile string

	// Filed in conversion.
	totalFrames int

	// State information.
	gotInput    bool
	gotDuration bool

	// File info.
	duration string
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

		stderrScanner.Scan()
		t.workWithOutput(stderrScanner.Text())
	}

	log.Println("Stopped reading ffmpeg output")
}

// Printing progress for this task.
func (t *Task) workWithOutput(output string) {
	if output == "" {
		return
	}

}
