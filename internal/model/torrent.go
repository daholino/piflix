package model

import (
	"time"
)

type TorrentStatus int

const (
	TorrentStatusDownloading TorrentStatus = iota
	TorrentStatusRendering
	TorrentStatusReady
)

type Torrent struct {
	ID        string        `json:"id"`
	Hash      string        `json:"hash"`
	Name      string        `json:"name"`
	Status    TorrentStatus `json:"status"`
	Magnet    string        `json:"magnet"`
	AddedTime time.Time     `json:"added_time"`
	Files     []File        `json:"files"`
	Poster    NullString    `json:"poster"`
}

type TorrentProgress struct {
	ID        string `json:"id"`
	Hash      string `json:"hash"`
	Name      string `json:"name"`
	Progress  int32  `json:"progress"`
	TotalSize int64  `json:"total_size"`
	BytesRead int64  `json:"bytes_read"`
}
