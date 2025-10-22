package main

import(
	"os/exec"
	"fmt"
)

func processVideoForFastStart(filePath string) (string, error){
	out := filePath + ".processing"

	cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", out)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error preproccessing video for fast start: %s", err)
	}

	return out, nil
}
