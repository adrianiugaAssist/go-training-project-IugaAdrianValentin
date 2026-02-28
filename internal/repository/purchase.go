package repository

import (
	"context"
	"database/sql"
	"fmt"

	"example/data-access/internal/logger"
	"example/data-access/internal/models"
)

// Purchase database operations

// GetAllPurchases calls stored procedure to get all purchases in the database
func GetAllPurchases(db *sql.DB) ([]models.Purchase, error) {
	var purchases []models.Purchase

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	rows, err := db.QueryContext(ctx, "CALL sp_get_all_purchases()")
	if err != nil {
		logger.Log.Errorw("Failed to call stored procedure sp_get_all_purchases", "error", err)
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

// GetPurchasesByUserID calls stored procedure to get purchases by a specific user
func GetPurchasesByUserID(db *sql.DB, userID int64) ([]models.Purchase, error) {
	var purchases []models.Purchase

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	rows, err := db.QueryContext(ctx, "CALL sp_get_purchases_by_user_id(?)", userID)
	if err != nil {
		logger.Log.Errorw("Failed to call stored procedure sp_get_purchases_by_user_id", "user_id", userID, "error", err)
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

// AddPurchase calls stored procedure to add a purchase to the database,
// returning the purchase ID of the new entry
func AddPurchase(db *sql.DB, p models.Purchase) (int64, error) {
	logger.Log.Debugw("Starting purchase through stored procedure", "user_id", p.UserID, "album_id", p.AlbumID, "quantity", p.Quantity)

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var purchaseID int64
	err := db.QueryRowContext(ctx, "CALL sp_add_purchase(?, ?, ?)", p.UserID, p.AlbumID, p.Quantity).Scan(&purchaseID)
	if err != nil {
		logger.Log.Errorw("Failed to call stored procedure sp_add_purchase", "error", err, "user_id", p.UserID, "album_id", p.AlbumID)
		return 0, fmt.Errorf("addPurchase: %v", err)
	}

	logger.Log.Infow("Purchase added successfully through stored procedure", "purchase_id", purchaseID, "user_id", p.UserID, "album_id", p.AlbumID, "quantity", p.Quantity)

	return purchaseID, nil
}

// GetUserPurchaseSummary calls stored procedure to get a user's purchases with album details and calculates total cost
func GetUserPurchaseSummary(db *sql.DB, userID int64) (models.UserPurchaseSummary, error) {
	summary := models.UserPurchaseSummary{}

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	// Call stored procedure to get user info and purchases
	rows, err := db.QueryContext(ctx, "CALL sp_get_user_purchase_summary(?)", userID)
	if err != nil {
		logger.Log.Errorw("Failed to call stored procedure sp_get_user_purchase_summary", "user_id", userID, "error", err)
		return summary, fmt.Errorf("getUserPurchaseSummary %d: %v", userID, err)
	}
	defer rows.Close()

	// First row contains user info
	if rows.Next() {
		if err := rows.Scan(&summary.UserID, &summary.Username, &summary.Email); err != nil {
			logger.Log.Errorw("Failed to scan user info from stored procedure", "user_id", userID, "error", err)
			return summary, fmt.Errorf("getUserPurchaseSummary %d: %v", userID, err)
		}
	}

	// Move to next result set with purchase details
	if !rows.NextResultSet() {
		logger.Log.Infow("No purchase details found for user", "user_id", userID)
		return summary, nil
	}

	totalCost := float32(0)
	for rows.Next() {
		var detail models.PurchaseDetail
		var price float64
		if err := rows.Scan(&detail.ID, &detail.AlbumID, &detail.AlbumTitle, &detail.Artist, &price, &detail.Quantity); err != nil {
			logger.Log.Errorw("Failed to scan purchase detail from stored procedure", "user_id", userID, "error", err)
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

// GetAllUsersPurchaseSummary calls stored procedure to get purchase summaries for all users
func GetAllUsersPurchaseSummary(db *sql.DB) ([]models.UserPurchaseSummary, error) {
	var summaries []models.UserPurchaseSummary

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	// Call stored procedure to get all users purchase summaries
	rows, err := db.QueryContext(ctx, "CALL sp_get_all_users_purchase_summary()")
	if err != nil {
		logger.Log.Errorw("Failed to call stored procedure sp_get_all_users_purchase_summary", "error", err)
		return nil, fmt.Errorf("getAllUsersPurchaseSummary: %v", err)
	}
	defer rows.Close()

	currentUserID := int64(-1)
	var currentSummary models.UserPurchaseSummary
	totalCost := float32(0)

	for rows.Next() {
		var userID int64
		var username, email string
		var purchaseID, albumID, quantity *int64
		var albumTitle, artist *string
		var price *float64

		if err := rows.Scan(&userID, &username, &email, &purchaseID, &albumID, &albumTitle, &artist, &price, &quantity); err != nil {
			logger.Log.Errorw("Failed to scan user purchase summary from stored procedure", "error", err)
			return nil, fmt.Errorf("getAllUsersPurchaseSummary: %v", err)
		}

		// If we moved to a new user, save the previous one
		if userID != currentUserID {
			if currentUserID != -1 {
				currentSummary.TotalCost = totalCost
				summaries = append(summaries, currentSummary)
			}
			currentUserID = userID
			currentSummary = models.UserPurchaseSummary{
				UserID:    userID,
				Username:  username,
				Email:     email,
				Purchases: []models.PurchaseDetail{},
			}
			totalCost = 0
		}

		// Add purchase detail if present
		if purchaseID != nil && albumID != nil {
			detail := models.PurchaseDetail{
				ID:         *purchaseID,
				AlbumID:    *albumID,
				AlbumTitle: *albumTitle,
				Artist:     *artist,
				Quantity:   int(*quantity),
			}
			if price != nil {
				detail.Price = float32(*price)
				detail.Subtotal = detail.Price * float32(detail.Quantity)
				totalCost += detail.Subtotal
			}
			currentSummary.Purchases = append(currentSummary.Purchases, detail)
		}
	}

	// Don't forget the last user
	if currentUserID != -1 {
		currentSummary.TotalCost = totalCost
		summaries = append(summaries, currentSummary)
	}

	if err := rows.Err(); err != nil {
		logger.Log.Errorw("Error iterating all users purchase summary from stored procedure", "error", err)
		return nil, fmt.Errorf("getAllUsersPurchaseSummary: %v", err)
	}

	return summaries, nil
}
