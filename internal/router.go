package internal

import (
	"embed"
	"fmt"
	"mime"
	"piflix/internal/utility"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

func setupRoutes(engine *Engine, webFS *embed.FS) {
	torrentHandler := TorrentHandler{torrentManager: engine.torrentManager, hlsManager: engine.hlsManager, database: engine.database, config: engine.config}
	spaFileSystem := utility.EmbedFolder(*webFS, "web/piflix-web/build")

	engine.router.Use(cors.Default())

	engine.router.Use(static.Serve("/", spaFileSystem))

	engine.router.POST("/add-torrent", torrentHandler.AddTorrent)
	engine.router.DELETE("/torrent/:id", torrentHandler.DeleteTorrent)
	engine.router.GET("/downloaded-torrents", torrentHandler.DownloadedTorrents)
	engine.router.GET("/torrent/:id", torrentHandler.TorrentByID)
	engine.router.GET("/status", torrentHandler.Status)
	engine.router.POST("/torrent/:id/subtitle/:fileid", torrentHandler.AddSubtitle)
	engine.router.DELETE("/torrent/:id/subtitle/:fileid", torrentHandler.DeleteSubtitle)

	engine.router.Static("/media", fmt.Sprintf("%s/%s", engine.config.WorkDir, "media"))

	engine.router.NoRoute(func(c *gin.Context) {
		c.FileFromFS("/", spaFileSystem)
	})

	mime.AddExtensionType(".vtt", "text/vtt")
	mime.AddExtensionType(".m3u8", "application/vnd.apple.mpegurl")
	mime.AddExtensionType(".ts", "video/mp2t")
}
