package db

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"piflix/internal/model"

	_ "github.com/mattn/go-sqlite3"
)

const CURRENT_DB_VERSION int = 1

type SQLite struct {
	db *sql.DB
}

func NewSQLiteDatabase(workDir string) *SQLite {
	sqlite := &SQLite{}

	db, err := sql.Open("sqlite3", filepath.Join(workDir, "piflix.db"))
	if err != nil {
		log.Println("Couldn't create database. Error:", err)
	}

	sqlite.db = db

	err = sqlite.migrate()
	if err != nil {
		log.Println("Migration failed. Closing database. Reason:", err)
		sqlite.db.Close()

		return nil
	}

	return sqlite
}

// Managing models

func (sqlite *SQLite) SaveTorrent(t *model.Torrent) error {
	_, err := sqlite.db.Exec("INSERT INTO torrent(id, hash, name, magnet, status, added_time) VALUES (?, ?, ?, ?, ?, ?)", t.ID, t.Hash, t.Name, t.Magnet, t.Status, t.AddedTime)

	sqlite.saveTorrentFiles(t.Files)

	return err
}

func (sqlite *SQLite) GetDownloadingTorrents() ([]model.Torrent, error) {
	return sqlite.getTorrentWithStatus(model.TorrentStatusDownloading)
}

func (sqlite *SQLite) GetRenderingTorrents() ([]model.Torrent, error) {
	return sqlite.getTorrentWithStatus(model.TorrentStatusRendering)
}

func (sqlite *SQLite) GetDownloadedTorrents() ([]model.Torrent, error) {
	return sqlite.getTorrentWithStatus(model.TorrentStatusReady)
}

func (sqlite *SQLite) TorrentWithID(ID string) (*model.Torrent, error) {
	torrent := model.Torrent{}

	row := sqlite.db.QueryRow("SELECT id, hash, name, magnet, status, added_time, poster FROM torrent WHERE id = ?", ID)

	err := row.Scan(&torrent.ID, &torrent.Hash, &torrent.Name, &torrent.Magnet, &torrent.Status, &torrent.AddedTime, &torrent.Poster)
	if err != nil {
		return nil, err
	}

	torrent.Files, err = sqlite.getFilesForTorrentID(torrent.ID)
	if err != nil {
		return nil, err
	}

	return &torrent, nil
}

func (sqlite *SQLite) TorrentWithHash(hash string) (*model.Torrent, error) {
	torrent := model.Torrent{}

	row := sqlite.db.QueryRow("SELECT id, hash, name, magnet, status, added_time, poster FROM torrent WHERE hash = ?", hash)

	err := row.Scan(&torrent.ID, &torrent.Hash, &torrent.Name, &torrent.Magnet, &torrent.Status, &torrent.AddedTime, &torrent.Poster)
	if err != nil {
		return nil, err
	}

	torrent.Files, _ = sqlite.getFilesForTorrentID(torrent.ID)

	if err != nil {
		return nil, err
	}

	return &torrent, nil
}

func (sqlite *SQLite) DeleteTorrent(torrent *model.Torrent) error {
	sqlite.deleteTorrentFiles(torrent.ID)

	_, err := sqlite.db.Exec("DELETE FROM torrent where id = ?", torrent.ID)

	return err
}

func (sqlite *SQLite) DeleteFile(file *model.File) error {
	_, err := sqlite.db.Exec("DELETE FROM file where id = ?", file.ID)

	return err
}

func (sqlite *SQLite) SetStatusForTorrent(status model.TorrentStatus, ID string) error {
	_, err := sqlite.db.Exec("UPDATE torrent SET status = ? WHERE id = ?", status, ID)

	return err
}

func (sqlite *SQLite) SetImagePathForTorrent(path string, ID string) error {
	_, err := sqlite.db.Exec("UPDATE torrent SET poster = ? WHERE id = ?", path, ID)

	return err
}

func (sqlite *SQLite) SetSubtitlePathForFile(subtitle string, ID int64) error {
	_, err := sqlite.db.Exec("UPDATE file SET subtitle = ? WHERE id = ?", subtitle, ID)

	return err
}

func (sqlite *SQLite) FileWithID(ID int64) (*model.File, error) {
	file := model.File{}

	row := sqlite.db.QueryRow("SELECT id, path, subtitle, torrent_id FROM file WHERE id = ?", ID)

	err := row.Scan(&file.ID, &file.Path, &file.Subtitle, &file.TorrentID)
	if err != nil {
		log.Println("File scan failed. Reason:", err)
		return nil, err
	}

	return &file, nil
}

func (sqlite *SQLite) getTorrentWithStatus(status model.TorrentStatus) ([]model.Torrent, error) {
	rows, err := sqlite.db.Query("SELECT id, hash, name, magnet, status, added_time, poster FROM torrent WHERE status = ?", status)
	if err != nil {
		return nil, err
	}

	torrents := []model.Torrent{}

	for rows.Next() {
		var torrent model.Torrent

		err := rows.Scan(&torrent.ID, &torrent.Hash, &torrent.Name, &torrent.Magnet, &torrent.Status, &torrent.AddedTime, &torrent.Poster)
		if err != nil {
			log.Println("Torrent scan failed. Reason:", err)
			continue
		}

		torrent.Files, err = sqlite.getFilesForTorrentID(torrent.ID)
		if err != nil {
			continue
		}

		torrents = append(torrents, torrent)
	}

	return torrents, nil
}

func (sqlite *SQLite) saveTorrentFiles(files []model.File) {
	for _, file := range files {
		sqlite.db.Exec("INSERT INTO file(path, torrent_id, subtitle) VALUES (?, ?, ?)", file.Path, file.TorrentID, file.Subtitle)
	}
}

func (sqlite *SQLite) deleteTorrentFiles(ID string) {
	sqlite.db.Exec("DELETE FROM file WHERE torrent_id = ?", ID)
}

func (sqlite *SQLite) getFilesForTorrentID(ID string) ([]model.File, error) {
	rows, err := sqlite.db.Query("SELECT id, path, subtitle FROM file WHERE torrent_id = ?", ID)
	if err != nil {
		return nil, err
	}

	files := []model.File{}

	for rows.Next() {
		var file model.File

		err := rows.Scan(&file.ID, &file.Path, &file.Subtitle)
		if err != nil {
			log.Println("File scan failed. Reason:", err)
			continue
		}

		files = append(files, file)
	}

	return files, nil
}

// Helper functions

func (sqlite *SQLite) migrate() error {
	version := sqlite.dbVersion()

	switch version {
	case 0:
		err := migrateToVersion1(sqlite.db)
		if err != nil {
			return err
		}
		fallthrough
	default:
		log.Println("Database is fully migrated.")
	}

	sqlite.setDBVersion(CURRENT_DB_VERSION)

	return nil
}

func (sqlite *SQLite) dbVersion() int {
	var userVersion int = 0
	rows, err := sqlite.db.Query("PRAGMA user_version")
	if err != nil {
		return userVersion
	}

	defer rows.Close()

	for rows.Next() {
		rows.Scan(&userVersion)
	}

	return userVersion
}

func (sqlite *SQLite) setDBVersion(version int) {
	_, err := sqlite.db.Exec(fmt.Sprintf("PRAGMA user_version = %d", version))
	if err != nil {
		log.Println("Error setting database user version. Error:", err)
	}
}

// Individual migrations

func migrateToVersion1(db *sql.DB) error {
	err := createTorrent(db)
	if err != nil {
		return err
	}

	err = createFiles(db)

	return err
}

// Table creation helpers

func createTorrent(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE torrent (id TEXT PRIMARY KEY, hash TEXT, name TEXT, magnet TEXT, status INTEGER, added_time DATETIME, poster TEXT)")
	if err != nil {
		return err
	}

	_, err = db.Exec("CREATE INDEX idx_torrent_status ON torrent(status)")

	return err
}

func createFiles(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE file (id INTEGER PRIMARY KEY, path TEXT, subtitle TEXT, torrent_id TEXT, FOREIGN KEY (torrent_id) REFERENCES torrent(id))")
	if err != nil {
		return err
	}

	return err
}
