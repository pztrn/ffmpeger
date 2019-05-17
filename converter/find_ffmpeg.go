package converter

import (
	// stdlib
	"bytes"
	"log"
	"os/exec"
	"strings"
)

func findffmpeg() {
	// Search for ffmpeg.
	var err error
	ffmpegPath, err = exec.LookPath("ffmpeg")
	if err != nil {
		log.Fatalln("Failed to find ffmpeg in path:", err.Error())
	}

	// Get ffmpeg version.
	stdout := bytes.NewBuffer(nil)
	ffmpegVersionCmd := exec.Command(ffmpegPath, "-version")
	ffmpegVersionCmd.Stdout = stdout
	err1 := ffmpegVersionCmd.Run()
	if err1 != nil {
		log.Fatalln("Failed to get ffmpeg version:", err1.Error())
	}

	stdoutString := stdout.String()
	if len(stdoutString) == 0 {
		log.Fatalln("Something weird happened and '" + ffmpegPath + " -version' returns nothing! Check your ffmpeg installation!")
	}
	// ffmpeg prints it's version on line 1.
	ffmpegVersion := strings.Split(stdoutString, " ")[2]

	log.Println("ffmpeg found at", ffmpegPath, "with version", ffmpegVersion)
}
