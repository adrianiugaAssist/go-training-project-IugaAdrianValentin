package repository

import (
	"context"
	"database/sql"
	"fmt"

	"example/data-access/internal/constants"
	"example/data-access/internal/logger"
	"example/data-access/internal/models"
)

// Album database operations

// GetAllAlbums calls stored procedure to get all albums in the database
func GetAllAlbums(db *sql.DB) ([]models.Album, error) {
	var albums []models.Album

	ctx, cancel := context.WithTimeout(context.Background(), constants.DBTimeout)
	defer cancel()

	rows, err := db.QueryContext(ctx, "CALL sp_get_all_albums()")
	if err != nil {
		logger.Log.Errorw("Failed to call stored procedure sp_get_all_albums", "error", err)
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

// GetAlbumsByArtist calls stored procedure to get albums that have the specified artist name
func GetAlbumsByArtist(db *sql.DB, name string) ([]models.Album, error) {
	var albums []models.Album

	ctx, cancel := context.WithTimeout(context.Background(), constants.DBTimeout)
	defer cancel()

	rows, err := db.QueryContext(ctx, "CALL sp_get_albums_by_artist(?)", name)
	if err != nil {
		logger.Log.Errorw("Failed to call stored procedure sp_get_albums_by_artist", "artist", name, "error", err)
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

// GetAlbumByID calls stored procedure to get the album with the specified ID
func GetAlbumByID(db *sql.DB, id int64) (models.Album, error) {
	var alb models.Album

	ctx, cancel := context.WithTimeout(context.Background(), constants.DBTimeout)
	defer cancel()

	row := db.QueryRowContext(ctx, "CALL sp_get_album_by_id(?)", id)
	var price float64
	if err := row.Scan(&alb.ID, &alb.Title, &alb.Artist, &price, &alb.Stock); err != nil {
		logger.Log.Errorw("Album not found", "album_id", id, "error", err)
		return alb, fmt.Errorf("getAlbumByID %d: %v", id, err)
	}
	alb.Price = float32(price)

	return alb, nil
}

// AddAlbum calls stored procedure to add an album to the database,
// returning the album ID of the new entry
func AddAlbum(db *sql.DB, alb models.Album) (int64, error) {
	logger.Log.Infow("Adding new album", "title", alb.Title, "artist", alb.Artist, "price", alb.Price, "stock", alb.Stock)

	ctx, cancel := context.WithTimeout(context.Background(), constants.DBTimeout)
	defer cancel()

	var albumID int64
	err := db.QueryRowContext(ctx, "CALL sp_add_album(?, ?, ?, ?)", alb.Title, alb.Artist, alb.Price, alb.Stock).Scan(&albumID)
	if err != nil {
		logger.Log.Errorw("Failed to call stored procedure sp_add_album", "error", err, "title", alb.Title)
		return 0, fmt.Errorf("addAlbum: %v", err)
	}

	logger.Log.Infow("Album created", "album_id", albumID, "title", alb.Title, "artist", alb.Artist)
	return albumID, nil
}
