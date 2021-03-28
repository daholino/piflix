package model

type File struct {
	ID        int64      `json:"id"`
	TorrentID string     `json:"-"`
	Path      string     `json:"path"`
	Subtitle  NullString `json:"subtitle"`
}
