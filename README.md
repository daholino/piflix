<img src="https://github.com/daholino/piflix/blob/main/docs/piflix-logo.png?raw=true" width="150">

# Piflix - Your mini streaming helper 🎥 🍿

Piflix is an app that provides simple torrent downloading and streaming across the devices in your local network. To access it you just need to be in the local network and to have any Internet browser.

It consists of two parts:
- backend that is written using Go
- frontend that is written using React

The frontend part ([piflix-web](https://github.com/daholino/piflix-web)) is linked to this repo as a git submodule. Both parts are linked in the same binary using go:embed directive.

> I use iPad for consuming the Internet multimedia and I wanted to have a solution to watch what I want. 🏴‍☠️ The Piflix is built with that in mind and it works on any mobile device that has a browser.

![Piflix Demo](https://github.com/daholino/piflix/blob/main/docs/piflix-demo.gif?raw=true)

## Usage

In the releases section you can find the binaries for Mac Intel and Raspberry Pi ARM architecture built with the latest stable version.

To build the project by yourself, clone the repo and use the provided Makefile. Available recipes are:
- prepare: prepares the dependencies and updates them to the latest version
- run-mac: runs the backend
- build-mac or build-linux: builds the backend
- package-mac or package-linux: builds the web and then proceeds to building the backend
- build-web: builds the frontend to create static files

If you want to build the web version you will need to have `npm` installed.

### Configuration

There is a `config-template.yaml` in the repo that explains all the fields that can be configured. When running the binary file you can provide the `-c` flag that will contain the path to the config file. If this flag is not set it will default to the directory of where the binary is stored.

## How it works

Currently the torrents can be added only via magnet link. Once the torrent is downloaded it goes into processing status where `ffmpeg` is used to create segments for streaming by using the HLS protocol. I chose HLS because I primarily use Apple devices and their native players all support HLS.

In the configuration file you can optimise the ffmpeg options for Raspberry Pi so that the flags for ffmpeg use a hardware accelerated encoder. The encoder that is used on Raspberry Pi devices is `h264_omx`. Although the `h264_v4l2m2m` is faster its results are not consisent and I struggled to make it work with some input files.

When the processing is finished the torrent is ready to be streamed and then you can add the subtitle to it, watch it or copy the playlist url and use any other software to stream it.

For persisting the data a single SQLite database is used. The database and all downloaded and processed files are stored in the working directory that is specified in the config file.

The processing phase might last longer on devices with poor performance. For example I'll get ~250 fps while processing the same file on my Mac but ~30 fps on the rpi. If you have any tips how to improve this I will be very grateful! To optimise processing on Raspberry Pis use only one resolution (e.g. 720p) for output and try to use input files with resolutions of 1080p and lower.

If you experience issues on your raspberry devices while encoding videos I suggest that you download and build the latest version of ffmpeg. You can use [this](https://gist.github.com/wildrun0/86a890585857a36c90110cee275c45fd) script to do it.

## Libraries

Libraries that are used in this project. Thanks to all the people that made them free and available!

- [anacrolix/torrent](https://github.com/anacrolix/torrent)
- [asticode/go-astisub](https://github.com/asticode/go-astisub)
- [robfig/cron](https://github.com/robfig/cron)
- [gin-gonic/gin](https://github.com/gin-gonic/gin)
- [spf13/viper](https://github.com/spf13/viper)
- [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)
- [h2non/filetype](https://github.com/h2non/filetype)
- [google/uuid](https://github.com/google/uuid)

## Contributing

If you experience any issues please report them in the issues section so that they can be addressed. Also, if you are interested in improving Piflix please submit pull requests!