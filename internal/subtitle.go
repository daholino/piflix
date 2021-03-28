package internal

import (
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"piflix/internal/db"
	"piflix/internal/model"

	"github.com/asticode/go-astisub"
)

func processAndAddSubtitle(subtitle *multipart.FileHeader, file *model.File, torrent *model.Torrent, database *db.SQLite, workDir string) error {
	srtFile, _ := subtitle.Open()
	defer srtFile.Close()
	sub, err := astisub.ReadFromSRT(srtFile)
	if err != nil {
		return err
	}

	fileIndex, _ := calculateFileIndex(file, torrent)
	subtitleRelativePath := filepath.Join("media", torrent.ID, fmt.Sprint(fileIndex), "sub.vtt")
	subtitleDestPath := filepath.Join(workDir, subtitleRelativePath)
	_ = os.Remove(subtitleDestPath)

	diskSubtitleFile, err := os.Create(subtitleDestPath)
	if err != nil {
		return err
	}

	err = sub.WriteToWebVTT(diskSubtitleFile)
	if err != nil {
		return err
	}

	err = database.SetSubtitlePathForFile(subtitleRelativePath, file.ID)
	if err != nil {
		return err
	}

	return nil
}

func deleteSubtitle(file *model.File, torrent *model.Torrent, database *db.SQLite, workDir string) error {
	fileIndex, _ := calculateFileIndex(file, torrent)
	subtitleRelativePath := filepath.Join("media", torrent.ID, fmt.Sprint(fileIndex), "sub.vtt")
	subtitleDestPath := filepath.Join(workDir, subtitleRelativePath)
	_ = os.Remove(subtitleDestPath)

	err := database.SetSubtitlePathForFile("", file.ID)
	if err != nil {
		return err
	}

	return nil
}

func calculateFileIndex(file *model.File, torrent *model.Torrent) (int64, error) {
	for index, tFile := range torrent.Files {
		if tFile.ID == file.ID {
			return int64(index), nil
		}
	}

	return 0, errors.New("file not found")
}
