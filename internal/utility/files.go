package utility

import (
	"os"
	"path/filepath"
	"piflix/internal/model"
	"strings"
)

func DeleteDownloadedFiles(torrent *model.Torrent, workPath string) error {
	for _, file := range torrent.Files {
		os.RemoveAll(file.Path)

		pathComponents := strings.Split(file.Path, "/")
		if len(pathComponents) == 0 {
			continue
		}

		firstDir := pathComponents[0]
		torrentBasePath := filepath.Join(workPath, "downloads", firstDir)
		os.RemoveAll(torrentBasePath)
	}

	return nil
}
