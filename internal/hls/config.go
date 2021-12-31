package hls

import "errors"

type config struct {
	Name         string
	VideoBitrate string
	Maxrate      string
	BufSize      string
	AudioBitrate string
	Resolution   string
	Bandwidth    string
}

var preset = map[string]*config{
	"360p": {
		Name:         "360p",
		VideoBitrate: "800k",
		Maxrate:      "856k",
		BufSize:      "1200k",
		AudioBitrate: "96k",
		Resolution:   "640x360",
		Bandwidth:    "800000",
	},
	"480p": {
		Name:         "480p",
		VideoBitrate: "1400k",
		Maxrate:      "1498k",
		BufSize:      "2100k",
		AudioBitrate: "128k",
		Resolution:   "842x480",
		Bandwidth:    "1400000",
	},
	"720p": {
		Name:         "720p",
		VideoBitrate: "5000k",
		Maxrate:      "5350k",
		BufSize:      "10600k",
		AudioBitrate: "128k",
		Resolution:   "1280x720",
		Bandwidth:    "5000000",
	},
	"1080p": {
		Name:         "1080p",
		VideoBitrate: "5000k",
		Maxrate:      "5350k",
		BufSize:      "10600k",
		AudioBitrate: "192k",
		Resolution:   "1920x1080",
		Bandwidth:    "5000000",
	},
}

func getConfig(res string) (*config, error) {
	cfg, ok := preset[res]
	if !ok {
		return nil, errors.New("preset not found")
	}

	return cfg, nil
}
