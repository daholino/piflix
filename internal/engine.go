package internal

import (
	"embed"
	"log"
	"os"
	"path/filepath"
	"piflix/internal/db"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

type Engine struct {
	config         *Config
	router         *gin.Engine
	database       *db.SQLite
	torrentManager *TorrentManager
	hlsManager     *HLSManager
	cron           *cron.Cron
}

func NewEngine(config *Config, webFS *embed.FS) *Engine {
	engine := &Engine{}
	engine.config = config

	engine.router = gin.Default()
	engine.database = db.NewSQLiteDatabase(config.WorkDir)
	engine.torrentManager = NewTorrentManager(engine.config)
	engine.torrentManager.database = engine.database
	engine.hlsManager = NewHLSManager(engine.database, engine.config)

	if !engine.checkDependencies() {
		log.Fatalln("Can't start piflix. Running requirements not satisfied. Aborting.")
	}

	setupRoutes(engine, webFS)

	return engine
}

func (e *Engine) Run() {
	e.restartDownloads()
	e.restartRenders()

	e.setupCron()

	e.router.Run(":4000")
}

func (e *Engine) LogInfo() {
	e.torrentManager.printStatus()
}

func (e *Engine) setupCron() {
	e.cron = cron.New()

	e.cron.AddFunc("@every 3s", e.checkStatus)

	e.cron.Start()
}

func (e *Engine) restartDownloads() {
	torrents, err := e.database.GetDownloadingTorrents()
	if err != nil {
		log.Println(err)
		return
	}

	e.torrentManager.resumeTorrents(torrents)
}

func (e *Engine) restartRenders() {
	torrents, err := e.database.GetRenderingTorrents()
	if err != nil {
		log.Println(err)
		return
	}

	for _, torrent := range torrents {
		e.hlsManager.RenderQueueChan <- torrent.ID
	}
}

func (e *Engine) checkStatus() {
	e.LogInfo()

	e.torrentManager.checkForCompletedTorrents(e.hlsManager.RenderQueueChan)
}

func (e *Engine) checkDependencies() bool {
	return e.checkDirectories() && e.hlsManager.checkDependencies()
}

func (e *Engine) checkDirectories() bool {
	err := os.MkdirAll(e.config.WorkDir, os.ModePerm)
	if err != nil {
		return false
	}

	err = os.MkdirAll(filepath.Join(e.config.WorkDir, "downloads"), os.ModePerm)
	if err != nil {
		return false
	}

	err = os.MkdirAll(filepath.Join(e.config.WorkDir, "media"), os.ModePerm)

	return err == nil
}
