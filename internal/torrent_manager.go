package internal

import (
	"errors"
	"log"
	"math"
	"os"
	"path/filepath"
	"piflix/internal/db"
	"piflix/internal/model"
	"strings"
	"time"

	logger "github.com/anacrolix/log"
	"github.com/anacrolix/torrent"
	"github.com/google/uuid"
)

const fileSizeLimit int64 = 67_108_864

type ActiveTorrent struct {
	ID        string
	torrent   *torrent.Torrent
	totalSize int64
	filePaths []string
}

// TorrentManager manages all torrent related functionalities like downloads, progress updates, etc..
type TorrentManager struct {
	Client         *torrent.Client
	activeTorrents map[string]*ActiveTorrent
	database       *db.SQLite
	config         *Config
}

func NewTorrentManager(config *Config) *TorrentManager {
	cfg := torrent.NewDefaultClientConfig()
	cfg.NoUpload = true
	cfg.DataDir = filepath.Join(config.WorkDir, "downloads")

	fileLogger := logger.StreamLogger{
		W:   log.Writer(),
		Fmt: logger.LineFormatter,
	}
	cfg.Logger = logger.Logger{LoggerImpl: fileLogger}

	client, err := torrent.NewClient(cfg)

	if err != nil {
		log.Println(err)
		return nil
	}

	return &TorrentManager{
		Client:         client,
		activeTorrents: map[string]*ActiveTorrent{},
		config:         config,
	}
}

func (tm *TorrentManager) addTorrentWithMagnet(magnet string) *ActiveTorrent {
	t, err := tm.Client.AddMagnet(magnet)

	if err != nil {
		log.Println(err)
		return nil
	}

	// Wait for torrent info to download.
	<-t.GotInfo()

	activeTorrent := &ActiveTorrent{
		ID:        uuid.NewString(),
		torrent:   t,
		totalSize: calculateSize(t),
		filePaths: getValidFilepaths(t),
	}

	return activeTorrent
}

func (tm *TorrentManager) DownloadActiveTorrent(activeTorrent *ActiveTorrent) {
	tm.activeTorrents[activeTorrent.ID] = activeTorrent

	startDownloading(activeTorrent.torrent)
}

func (tm *TorrentManager) stopAndRemoveTorrentWithID(id string) {
	activeTorrent, ok := tm.activeTorrents[id]
	if !ok {
		return
	}

	activeTorrent.torrent.Drop()
	delete(tm.activeTorrents, id)
}

func (tm *TorrentManager) resumeTorrents(torrents []model.Torrent) {
	for _, torrent := range torrents {
		at := tm.addTorrentWithMagnet(torrent.Magnet)

		// Override generated random ID because we are resuming the torrent.
		at.ID = torrent.ID

		removeAllFilesForTorrent(at, tm.config.WorkDir)
		at.torrent.VerifyData()

		tm.DownloadActiveTorrent(at)
	}
}

func (tm *TorrentManager) checkForCompletedTorrents(renderChan chan string) {
	for key, t := range tm.activeTorrents {
		downloadedBytes := t.torrent.Stats().BytesReadData
		if t.totalSize <= downloadedBytes.Int64() {
			delete(tm.activeTorrents, key)

			time.AfterFunc(5*time.Second, func() {
				t.torrent.Drop()

				checkAndRemoveInvalidFilesFromDisk(t, tm.config.WorkDir)
				tm.database.SetStatusForTorrent(model.TorrentStatusRendering, t.ID)

				go tm.downloadImageForActiveTorrent(t)

				renderChan <- t.ID
			})
		}
	}
}

func (tm *TorrentManager) downloadImageForActiveTorrent(t *ActiveTorrent) {
	omdbManager := OMDBManager{}
	imageURL, err := omdbManager.DownloadImageForMovie(t.torrent.Name())

	if err != nil {
		log.Println(err)
	} else if len(imageURL) > 0 {
		tm.database.SetImagePathForTorrent(imageURL, t.ID)
	}
}

func (tm *TorrentManager) printStatus() {
	torrentsWithProgress := tm.GetDownloadingTorrentsWithProgress()
	for _, torrentProgress := range torrentsWithProgress {
		log.Printf("Torrent %s: %d%% torrent downloaded. Bytes read %d. Total size %d.", torrentProgress.ID, torrentProgress.Progress, torrentProgress.BytesRead, torrentProgress.TotalSize)
	}
}

func (tm *TorrentManager) GetDownloadingTorrentsWithProgress() []model.TorrentProgress {
	torrentsInProgress := []model.TorrentProgress{}

	for _, activeTorrent := range tm.activeTorrents {
		bytesRead := activeTorrent.torrent.Stats().BytesReadData
		progress := float64(bytesRead.Int64()) / float64(activeTorrent.totalSize) * 100

		torrentProgress := model.TorrentProgress{
			ID:        activeTorrent.ID,
			Hash:      activeTorrent.torrent.InfoHash().String(),
			Name:      activeTorrent.torrent.Name(),
			Progress:  int32(math.Min(100, progress)),
			TotalSize: activeTorrent.totalSize,
			BytesRead: bytesRead.Int64(),
		}

		torrentsInProgress = append(torrentsInProgress, torrentProgress)
	}

	return torrentsInProgress
}

// Helper functions

func startDownloading(t *torrent.Torrent) {
	files := t.Files()
	for _, file := range files {
		if file.Length() < fileSizeLimit {
			log.Println("Skipping because of small size:", file.Path())
			file.SetPriority(torrent.PiecePriorityNone)
			continue
		}

		log.Println("Downloading:", file.Path())

		file.SetPriority(torrent.PiecePriorityNormal)
	}
}

func calculateSize(t *torrent.Torrent) int64 {
	totalSize := int64(0)
	files := t.Files()

	for _, file := range files {
		if file.Length() < fileSizeLimit {
			continue
		}

		totalSize += file.Length()
	}

	return totalSize
}

func getValidFilepaths(t *torrent.Torrent) []string {
	filepaths := []string{}
	files := t.Files()

	for _, file := range files {
		if file.Length() < fileSizeLimit {
			continue
		}

		filepaths = append(filepaths, file.Path())
	}

	return filepaths
}

func checkAndRemoveInvalidFilesFromDisk(at *ActiveTorrent, workDir string) {
	files := at.torrent.Files()

	for _, file := range files {
		if file.Length() >= fileSizeLimit {
			continue
		}

		path := filepath.Join(workDir, "downloads", file.Path())

		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			continue
		}

		err := os.Remove(path)
		if err != nil {
			log.Println("Error when deleting file:", err)
		}
	}
}

func removeAllFilesForTorrent(at *ActiveTorrent, workDir string) {
	files := at.torrent.Files()

	for _, file := range files {
		path := filepath.Join(workDir, "downloads", file.Path())

		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			continue
		}

		err := os.Remove(path)
		if err != nil {
			log.Println("Error when deleting file:", err)
		}

		pathComponents := strings.Split(file.Path(), "/")
		if len(pathComponents) == 0 {
			continue
		}

		firstDir := pathComponents[0]
		torrentBasePath := filepath.Join(workDir, "downloads", firstDir)
		os.RemoveAll(torrentBasePath)
	}
}
