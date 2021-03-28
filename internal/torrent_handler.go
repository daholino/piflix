package internal

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"piflix/internal/db"
	"piflix/internal/model"
	"piflix/internal/utility"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type TorrentHandler struct {
	torrentManager *TorrentManager
	hlsManager     *HLSManager
	database       *db.SQLite
	config         *Config
}

func (th *TorrentHandler) AddTorrent(c *gin.Context) {
	var torrentRequest model.TorrentRequest

	if err := c.ShouldBindJSON(&torrentRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	activeTorrent := th.torrentManager.addTorrentWithMagnet(torrentRequest.Magnet)
	if activeTorrent == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid magnet"})
		return
	}

	torrent, _ := th.database.TorrentWithHash(activeTorrent.torrent.InfoHash().String())
	if torrent != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "torrent already added"})
		return
	}

	torrentModel := &model.Torrent{
		ID:        activeTorrent.ID,
		Hash:      activeTorrent.torrent.InfoHash().String(),
		Magnet:    torrentRequest.Magnet,
		Status:    model.TorrentStatusDownloading,
		AddedTime: time.Now(),
		Name:      activeTorrent.torrent.Name(),
		Files:     convertPathsToFiles(activeTorrent),
	}

	err := th.database.SaveTorrent(torrentModel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "couldn't save torrent"})
		return
	}

	th.torrentManager.DownloadActiveTorrent(activeTorrent)

	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"torrent": torrentModel,
	})
}

func (th *TorrentHandler) DeleteTorrent(c *gin.Context) {
	id := c.Param("id")
	if len(id) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no id"})
		return
	}

	torrent, err := th.database.TorrentWithID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no id"})
		return
	}

	err = th.database.DeleteTorrent(torrent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	workDir := th.config.WorkDir
	if torrent.Status == model.TorrentStatusReady {
		targetBasePath := filepath.Join(workDir, "media", torrent.ID)
		os.RemoveAll(targetBasePath)
	} else if torrent.Status == model.TorrentStatusRendering {
		th.hlsManager.stopProcessing(torrent.ID)
		targetBasePath := filepath.Join(workDir, "media", torrent.ID)
		os.RemoveAll(targetBasePath)
		utility.DeleteDownloadedFiles(torrent, workDir)
	} else {
		th.torrentManager.stopAndRemoveTorrentWithID(id)
		utility.DeleteDownloadedFiles(torrent, workDir)
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "OK",
	})
}

func (th *TorrentHandler) DownloadedTorrents(c *gin.Context) {
	torrents, err := th.database.GetDownloadedTorrents()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "OK",
		"torrents": torrents,
	})
}

func (th *TorrentHandler) TorrentByID(c *gin.Context) {
	id := c.Param("id")
	if len(id) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no id"})
		return
	}

	torrent, err := th.database.TorrentWithID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"torrent": torrent,
	})
}

func (th *TorrentHandler) Status(c *gin.Context) {
	renderingTorrents, err := th.database.GetRenderingTorrents()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	downloadingTorrents := th.torrentManager.GetDownloadingTorrentsWithProgress()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":               "OK",
		"rendering_torrents":   renderingTorrents,
		"downloading_torrents": downloadingTorrents,
	})
}

func (th *TorrentHandler) AddSubtitle(c *gin.Context) {
	id := c.Param("id")
	if len(id) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no id"})
		return
	}

	torrent, err := th.database.TorrentWithID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fileIDParam := c.Param("fileid")
	fileID, err := strconv.ParseInt(fileIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	file, err := th.database.FileWithID(fileID)
	if err != nil || file.TorrentID != torrent.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.New("invalid file id")})
		return
	}

	subtitle, err := c.FormFile("subtitle")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = processAndAddSubtitle(subtitle, file, torrent, th.database, th.config.WorkDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "OK",
	})
}

func (th *TorrentHandler) DeleteSubtitle(c *gin.Context) {
	id := c.Param("id")
	if len(id) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no id"})
		return
	}

	torrent, err := th.database.TorrentWithID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fileIDParam := c.Param("fileid")
	fileID, err := strconv.ParseInt(fileIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	file, err := th.database.FileWithID(fileID)
	if err != nil || file.TorrentID != torrent.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.New("invalid file id")})
		return
	}

	err = deleteSubtitle(file, torrent, th.database, th.config.WorkDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "OK",
	})
}

// Helper functions

func convertPathsToFiles(activeTorrent *ActiveTorrent) []model.File {
	files := []model.File{}

	for _, path := range activeTorrent.filePaths {
		file := model.File{
			Path:      path,
			TorrentID: activeTorrent.ID,
		}

		files = append(files, file)
	}

	return files
}
