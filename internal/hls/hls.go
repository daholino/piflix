package hls

import (
	"log"
	"os/exec"
)

// GenerateHLS will generate HLS file based on resolution presets.
// The available resolutions are: 360p, 480p, 720p and 1080p.
func GenerateHLS(ffmpegPath, srcPath, targetPath, resolution string, ffmpegOnRPI bool) (*exec.Cmd, error) {
	options, err := getOptions(srcPath, targetPath, resolution, ffmpegOnRPI)
	if err != nil {
		return nil, err
	}

	return GenerateHLSCustom(ffmpegPath, options)
}

// GenerateHLSCustom will generate HLS using the flexible options params.s
// options is array of string that accepted by ffmpeg command
func GenerateHLSCustom(ffmpegPath string, options []string) (*exec.Cmd, error) {
	cmd := exec.Command(ffmpegPath, options...)
	log.Println("Executing:", cmd.String())
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()

	return cmd, nil
}
