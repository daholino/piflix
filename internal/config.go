package internal

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	WorkDir     string `mapstructure:"work_dir"`
	FfmpegPath  string `mapstructure:"ffmpeg_path"`
	FFmpegPI    bool   `mapstructure:"ffmpeg_pi"`
	LogPath     string `mapstructure:"log_path"`
	Resolutions string `mapstructure:"resolutions"`
}

func LoadConfig(path string) *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("couldn't load config file: %s", err))
	}

	config := &Config{}

	err = viper.Unmarshal(config)
	if err != nil {
		panic(fmt.Errorf("couldn't load read file: %s", err))
	}

	return config
}
