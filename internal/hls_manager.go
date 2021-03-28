package internal

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"piflix/internal/db"
	"piflix/internal/hls"
	"piflix/internal/model"
	"piflix/internal/utility"
	"strings"

	"github.com/h2non/filetype"
)

type HLSManager struct {
	RenderQueueChan chan string
	database        *db.SQLite
	activeCommands  map[string]*exec.Cmd
	config          *Config
}

func NewHLSManager(database *db.SQLite, config *Config) *HLSManager {
	hlsManager := &HLSManager{
		RenderQueueChan: make(chan string),
		database:        database,
		activeCommands:  map[string]*exec.Cmd{},
		config:          config,
	}

	go hlsManager.start()

	return hlsManager
}

func (hlsm *HLSManager) checkDependencies() bool {
	return utility.CommandExists(hlsm.config.FfmpegPath)
}

func (hlsm *HLSManager) start() {
	for {
		id := <-hlsm.RenderQueueChan

		torrent, err := hlsm.database.TorrentWithID(id)
		if err != nil {
			log.Println("HLSManager start() error:", err)
			continue
		}

		err = createDirectory(torrent, hlsm.config.WorkDir)
		if err != nil {
			log.Println("HLSManager start() error:", err)
			continue
		}

		go hlsm.startRender(torrent)
	}
}

func (hlsm *HLSManager) startRender(torrent *model.Torrent) {
	targetBasePath := filepath.Join(hlsm.config.WorkDir, "media", torrent.ID)
	os.RemoveAll(targetBasePath)

	validFiles := 0

	for index, file := range torrent.Files {
		if !isFileVideo(file.Path, hlsm.config.WorkDir) {
			log.Println("File", file.Path, "is not a video. Deleting it from database.")
			hlsm.database.DeleteFile(&file)
			continue
		}

		validFiles += 1

		err := hlsm.startFileRender(torrent, &file, index)
		if err != nil {
			os.RemoveAll(targetBasePath)
			return
		}
	}

	if validFiles == 0 {
		hlsm.database.DeleteTorrent(torrent)
	} else {
		hlsm.database.SetStatusForTorrent(model.TorrentStatusReady, torrent.ID)
	}

	utility.DeleteDownloadedFiles(torrent, hlsm.config.WorkDir)
}

func (hlsm *HLSManager) startFileRender(torrent *model.Torrent, file *model.File, fileIndex int) error {
	srcPath := filepath.Join(hlsm.config.WorkDir, "downloads", file.Path)
	targetPath := filepath.Join(hlsm.config.WorkDir, "media", torrent.ID, fmt.Sprint(fileIndex))
	resOptions := strings.Split(utility.StripSpaces(hlsm.config.Resolutions), ",")

	err := os.MkdirAll(targetPath, os.ModePerm)
	if err != nil {
		log.Println("Couldn't create directories for torrent file.")
		return err
	}

	variants, _ := hls.GenerateHLSVariant(resOptions, "")
	hls.GeneratePlaylist(variants, targetPath, "")

	for _, res := range resOptions {
		cmd, err := hls.GenerateHLS(hlsm.config.FfmpegPath, srcPath, targetPath, res, hlsm.config.FFmpegPI)
		if err != nil {
			log.Println("HLS generation returned error:", err)
			return err
		}

		err = cmd.Start()
		if err != nil {
			return err
		}

		hlsm.activeCommands[torrent.ID] = cmd

		err = cmd.Wait()
		delete(hlsm.activeCommands, torrent.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (hlsm *HLSManager) stopProcessing(ID string) error {
	cmd := hlsm.activeCommands[ID]
	if cmd == nil {
		return nil
	}

	err := cmd.Process.Kill()

	delete(hlsm.activeCommands, ID)

	return err
}

func createDirectory(torrent *model.Torrent, workDir string) error {
	path := filepath.Join(workDir, "media", torrent.ID)
	err := os.MkdirAll(path, os.ModePerm)

	return err
}

func isFileVideo(path string, workDir string) bool {
	srcPath := filepath.Join(workDir, "downloads", path)

	file, err := os.Open(srcPath)
	if err != nil {
		log.Println("Couldn't open file to check if it is video. Error:", err)
		return false
	}

	head := make([]byte, 261)
	_, err = file.Read(head)
	if err != nil {
		log.Println("Couldn't read file to check if it is video. Error:", err)
		return false
	}

	return filetype.IsVideo(head)
}
