package repository

import (
	"database/sql"
	"fmt"

	"example/data-access/internal/logger"
	"example/data-access/internal/models"
)

// Album database operations

// GetAllAlbums queries for all albums in the database
func GetAllAlbums(db *sql.DB) ([]models.Album, error) {
	var albums []models.Album

	rows, err := db.Query("SELECT id, title, artist, price, stock FROM album")
	if err != nil {
		logger.Log.Errorw("Failed to query albums", "error", err)
		return nil, fmt.Errorf("getAllAlbums: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var alb models.Album
		var price float64
		if err := rows.Scan(&alb.ID, &alb.Title, &alb.Artist, &price, &alb.Stock); err != nil {
			logger.Log.Errorw("Failed to scan album", "error", err)
			return nil, fmt.Errorf("getAllAlbums: %v", err)
		}
		alb.Price = float32(price)
		albums = append(albums, alb)
	}

	if err := rows.Err(); err != nil {
		logger.Log.Errorw("Error iterating albums", "error", err)
		return nil, fmt.Errorf("getAllAlbums: %v", err)
	}

	return albums, nil
}

// GetAlbumsByArtist queries for albums that have the specified artist name
func GetAlbumsByArtist(db *sql.DB, name string) ([]models.Album, error) {
	var albums []models.Album

	rows, err := db.Query("SELECT id, title, artist, price, stock FROM album WHERE artist = ?", name)
	if err != nil {
		logger.Log.Errorw("Failed to query albums by artist", "artist", name, "error", err)
		return nil, fmt.Errorf("getAlbumsByArtist %q: %v", name, err)
	}
	defer rows.Close()

	for rows.Next() {
		var alb models.Album
		var price float64
		if err := rows.Scan(&alb.ID, &alb.Title, &alb.Artist, &price, &alb.Stock); err != nil {
			logger.Log.Errorw("Failed to scan album", "artist", name, "error", err)
			return nil, fmt.Errorf("getAlbumsByArtist %q: %v", name, err)
		}
		alb.Price = float32(price)
		albums = append(albums, alb)
	}

	if err := rows.Err(); err != nil {
		logger.Log.Errorw("Error iterating albums by artist", "artist", name, "error", err)
		return nil, fmt.Errorf("getAlbumsByArtist %q: %v", name, err)
	}

	return albums, nil
}

// GetAlbumByID queries for the album with the specified ID
func GetAlbumByID(db *sql.DB, id int64) (models.Album, error) {
	var alb models.Album

	row := db.QueryRow("SELECT id, title, artist, price, stock FROM album WHERE id = ?", id)
	var price float64
	if err := row.Scan(&alb.ID, &alb.Title, &alb.Artist, &price, &alb.Stock); err != nil {
		logger.Log.Errorw("Album not found", "album_id", id, "error", err)
		return alb, fmt.Errorf("getAlbumByID %d: %v", id, err)
	}
	alb.Price = float32(price)

	return alb, nil
}

// AddAlbum adds the specified album to the database,
// returning the album ID of the new entry
func AddAlbum(db *sql.DB, alb models.Album) (int64, error) {
	logger.Log.Infow("Adding new album", "title", alb.Title, "artist", alb.Artist, "price", alb.Price, "stock", alb.Stock)

	result, err := db.Exec("INSERT INTO album (title, artist, price, stock) VALUES (?, ?, ?, ?)", alb.Title, alb.Artist, alb.Price, alb.Stock)
	if err != nil {
		logger.Log.Errorw("Failed to insert album", "error", err, "title", alb.Title)
		return 0, fmt.Errorf("addAlbum: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		logger.Log.Errorw("Failed to get album ID", "error", err)
		return 0, fmt.Errorf("addAlbum: %v", err)
	}

	logger.Log.Infow("Album created", "album_id", id, "title", alb.Title, "artist", alb.Artist)
	return id, nil
}
