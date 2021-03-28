package main

import (
	"embed"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"piflix/internal"
	"syscall"
)

var engine *internal.Engine

//go:embed web/piflix-web/build/*
var webFS embed.FS

var version string = "undefined"

func main() {
	configPath := flag.String("c", ".", "Path for config file.")
	versionFlag := flag.Bool("v", false, "if set, prints version and exits")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}

	config := internal.LoadConfig(*configPath)

	signalChannel := make(chan os.Signal, 2)
	setupSignalHandling(signalChannel)

	internal.SetupLogger(config.LogPath)

	engine = internal.NewEngine(config, &webFS)
	engine.Run()
}

func setupSignalHandling(sigc chan os.Signal) {
	signal.Notify(sigc, syscall.SIGUSR1)
	go func() {
		for {
			sig := <-sigc
			switch sig {
			case syscall.SIGUSR1:
				engine.LogInfo()
			}
		}
	}()
}
