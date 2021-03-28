package hls

import "path/filepath"

func getOptions(srcPath, targetPath, res string, ffmpegOnRPI bool) ([]string, error) {
	config, err := getConfig(res)
	if err != nil {
		return nil, err
	}

	filenameTS := filepath.Join(targetPath, res+"_%03d.ts")
	filenameM3U8 := filepath.Join(targetPath, res+".m3u8")

	var options []string
	if ffmpegOnRPI {
		options = []string{
			"-hide_banner",
			"-y",
			"-i", srcPath,
			"-map", "0",
			"-map", "-0:s",
			"-c:a", "aac",
			"-b:a", config.AudioBitrate,
			"-ac", "2",
			"-ar", "48000",
			"-c:v", "h264_omx",
			"-profile:v", "main",
			"-crf", "20",
			"-pix_fmt", "yuv420p",
			"-sc_threshold", "0",
			"-g", "48",
			"-keyint_min", "48",
			"-hls_time", "10",
			"-hls_playlist_type", "vod",
			"-b:v", config.VideoBitrate,
			"-maxrate", config.Maxrate,
			"-bufsize", config.BufSize,
			"-preset", "ultrafast",
			"-hls_segment_filename", filenameTS,
			filenameM3U8,
		}
	} else {
		options = []string{
			"-hide_banner",
			"-y",
			"-i", srcPath,
			"-map", "0",
			"-map", "-0:s",
			"-vf", "scale=trunc(oh*a/2)*2:1080",
			"-c:a", "aac",
			"-b:a", config.AudioBitrate,
			"-ac", "2",
			"-c:v", "h264",
			"-profile:v", "main",
			"-crf", "20",
			"-pix_fmt", "yuv420p",
			"-sc_threshold", "0",
			"-g", "48",
			"-keyint_min", "48",
			"-hls_time", "10",
			"-hls_playlist_type", "vod",
			"-b:v", config.VideoBitrate,
			"-maxrate", config.Maxrate,
			"-bufsize", config.BufSize,
			"-preset", "ultrafast",
			"-hls_segment_filename", filenameTS,
			filenameM3U8,
		}
	}

	return options, nil
}
