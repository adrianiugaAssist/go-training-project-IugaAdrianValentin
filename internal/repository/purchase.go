package repository

import (
	"database/sql"
	"fmt"

	"example/data-access/internal/logger"
	"example/data-access/internal/models"
)

// Purchase database operations

// GetAllPurchases queries for all purchases in the database
func GetAllPurchases(db *sql.DB) ([]models.Purchase, error) {
	var purchases []models.Purchase

	rows, err := db.Query("SELECT id, user_id, album_id, quantity FROM purchase")
	if err != nil {
		logger.Log.Errorw("Failed to query purchases", "error", err)
		return nil, fmt.Errorf("getAllPurchases: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var p models.Purchase
		if err := rows.Scan(&p.ID, &p.UserID, &p.AlbumID, &p.Quantity); err != nil {
			logger.Log.Errorw("Failed to scan purchase", "error", err)
			return nil, fmt.Errorf("getAllPurchases: %v", err)
		}
		purchases = append(purchases, p)
	}

	if err := rows.Err(); err != nil {
		logger.Log.Errorw("Error iterating purchases", "error", err)
		return nil, fmt.Errorf("getAllPurchases: %v", err)
	}

	return purchases, nil
}

// GetPurchasesByUserID queries for purchases by a specific user
func GetPurchasesByUserID(db *sql.DB, userID int64) ([]models.Purchase, error) {
	var purchases []models.Purchase

	rows, err := db.Query("SELECT id, user_id, album_id, quantity FROM purchase WHERE user_id = ?", userID)
	if err != nil {
		logger.Log.Errorw("Failed to query purchases by user", "user_id", userID, "error", err)
		return nil, fmt.Errorf("getPurchasesByUserID %d: %v", userID, err)
	}
	defer rows.Close()

	for rows.Next() {
		var p models.Purchase
		if err := rows.Scan(&p.ID, &p.UserID, &p.AlbumID, &p.Quantity); err != nil {
			logger.Log.Errorw("Failed to scan purchase", "user_id", userID, "error", err)
			return nil, fmt.Errorf("getPurchasesByUserID %d: %v", userID, err)
		}
		purchases = append(purchases, p)
	}

	if err := rows.Err(); err != nil {
		logger.Log.Errorw("Error iterating purchases by user", "user_id", userID, "error", err)
		return nil, fmt.Errorf("getPurchasesByUserID %d: %v", userID, err)
	}

	return purchases, nil
}

// AddPurchase adds a purchase to the database,
// returning the purchase ID of the new entry
func AddPurchase(db *sql.DB, p models.Purchase) (int64, error) {
	logger.Log.Debugw("Starting purchase transaction", "user_id", p.UserID, "album_id", p.AlbumID, "quantity", p.Quantity)

	// Use a transaction to safely check and decrement stock
	tx, err := db.Begin()
	if err != nil {
		logger.Log.Errorw("Failed to begin transaction", "error", err, "user_id", p.UserID, "album_id", p.AlbumID)
		return 0, fmt.Errorf("addPurchase begin tx: %v", err)
	}
	defer func() {
		if err != nil {
			logger.Log.Warnw("Rolling back transaction", "user_id", p.UserID, "album_id", p.AlbumID, "error", err)
			tx.Rollback()
		}
	}()

	// Check current stock
	var stock int
	row := tx.QueryRow("SELECT stock FROM album WHERE id = ?", p.AlbumID)
	if err := row.Scan(&stock); err != nil {
		logger.Log.Warnw("Album not found", "album_id", p.AlbumID, "error", err)
		return 0, fmt.Errorf("addPurchase: album lookup: %v", err)
	}

	logger.Log.Debugw("Checked stock", "album_id", p.AlbumID, "current_stock", stock, "requested_quantity", p.Quantity)

	if stock < p.Quantity {
		logger.Log.Warnw("Insufficient stock", "album_id", p.AlbumID, "available_stock", stock, "requested_quantity", p.Quantity)
		tx.Rollback()
		return 0, fmt.Errorf("addPurchase: album %d out of stock or insufficient stock (have=%d, want=%d)", p.AlbumID, stock, p.Quantity)
	}

	// Insert purchase
	result, err := tx.Exec("INSERT INTO purchase (user_id, album_id, quantity) VALUES (?, ?, ?)", p.UserID, p.AlbumID, p.Quantity)
	if err != nil {
		logger.Log.Errorw("Failed to insert purchase", "error", err, "user_id", p.UserID, "album_id", p.AlbumID)
		return 0, fmt.Errorf("addPurchase: insert: %v", err)
	}

	// Decrement stock
	_, err = tx.Exec("UPDATE album SET stock = stock - ? WHERE id = ?", p.Quantity, p.AlbumID)
	if err != nil {
		logger.Log.Errorw("Failed to update stock", "error", err, "album_id", p.AlbumID, "decrement", p.Quantity)
		return 0, fmt.Errorf("addPurchase: update stock: %v", err)
	}

	logger.Log.Debugw("Stock decremented", "album_id", p.AlbumID, "old_stock", stock, "new_stock", stock-p.Quantity)

	id, err := result.LastInsertId()
	if err != nil {
		logger.Log.Errorw("Failed to get purchase ID", "error", err)
		return 0, fmt.Errorf("addPurchase: get id: %v", err)
	}

	if err := tx.Commit(); err != nil {
		logger.Log.Errorw("Failed to commit transaction", "error", err, "purchase_id", id)
		return 0, fmt.Errorf("addPurchase: commit: %v", err)
	}

	logger.Log.Infow("Purchase committed successfully", "purchase_id", id, "user_id", p.UserID, "album_id", p.AlbumID, "quantity", p.Quantity)

	return id, nil
}

// GetUserPurchaseSummary gets a user's purchases with album details and calculates total cost
func GetUserPurchaseSummary(db *sql.DB, userID int64) (models.UserPurchaseSummary, error) {
	summary := models.UserPurchaseSummary{}

	// Get user info
	userRow := db.QueryRow("SELECT id, username, email FROM user WHERE id = ?", userID)
	if err := userRow.Scan(&summary.UserID, &summary.Username, &summary.Email); err != nil {
		logger.Log.Errorw("Failed to get user for purchase summary", "user_id", userID, "error", err)
		return summary, fmt.Errorf("getUserPurchaseSummary %d: %v", userID, err)
	}

	// Get purchases with album details
	rows, err := db.Query(`
		SELECT p.id, p.album_id, a.title, a.artist, a.price, p.quantity
		FROM purchase p
		JOIN album a ON p.album_id = a.id
		WHERE p.user_id = ?
		ORDER BY p.id
	`, userID)
	if err != nil {
		logger.Log.Errorw("Failed to query user purchases for summary", "user_id", userID, "error", err)
		return summary, fmt.Errorf("getUserPurchaseSummary %d: %v", userID, err)
	}
	defer rows.Close()

	totalCost := float32(0)
	for rows.Next() {
		var detail models.PurchaseDetail
		var price float64
		if err := rows.Scan(&detail.ID, &detail.AlbumID, &detail.AlbumTitle, &detail.Artist, &price, &detail.Quantity); err != nil {
			logger.Log.Errorw("Failed to scan purchase detail", "user_id", userID, "error", err)
			return summary, fmt.Errorf("getUserPurchaseSummary %d: %v", userID, err)
		}
		detail.Price = float32(price)
		detail.Subtotal = detail.Price * float32(detail.Quantity)
		totalCost += detail.Subtotal
		summary.Purchases = append(summary.Purchases, detail)
	}

	summary.TotalCost = totalCost
	return summary, nil
}

// GetAllUsersPurchaseSummary gets purchase summaries for all users
func GetAllUsersPurchaseSummary(db *sql.DB) ([]models.UserPurchaseSummary, error) {
	var summaries []models.UserPurchaseSummary

	// Get all users
	userRows, err := db.Query("SELECT id, username, email FROM user ORDER BY id")
	if err != nil {
		logger.Log.Errorw("Failed to query all users for purchase summary", "error", err)
		return nil, fmt.Errorf("getAllUsersPurchaseSummary: %v", err)
	}
	defer userRows.Close()

	for userRows.Next() {
		var userID int64
		var username, email string
		if err := userRows.Scan(&userID, &username, &email); err != nil {
			logger.Log.Errorw("Failed to scan user for purchase summary", "error", err)
			return nil, fmt.Errorf("getAllUsersPurchaseSummary: %v", err)
		}

		summary, err := GetUserPurchaseSummary(db, userID)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}

	if err := userRows.Err(); err != nil {
		logger.Log.Errorw("Error iterating all users for purchase summary", "error", err)
		return nil, fmt.Errorf("getAllUsersPurchaseSummary: %v", err)
	}

	return summaries, nil
}
