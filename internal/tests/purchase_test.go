package tests

import (
	"database/sql"
	"sync"
	"testing"

	"example/data-access/internal/logger"
	"example/data-access/internal/models"
	"example/data-access/internal/repository"

	_ "github.com/mattn/go-sqlite3"
)

func init() {
	// Initialize logger for tests
	logger.InitLoggerDev()
}

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Enable WAL mode for better concurrency support
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		t.Fatalf("Failed to enable WAL: %v", err)
	}

	// Create tables
	schema := `
	CREATE TABLE user (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL,
		email TEXT NOT NULL
	);
	CREATE TABLE album (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		artist TEXT NOT NULL,
		price REAL NOT NULL,
		stock INTEGER NOT NULL
	);
	CREATE TABLE purchase (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		album_id INTEGER NOT NULL,
		quantity INTEGER NOT NULL,
		FOREIGN KEY(user_id) REFERENCES user(id),
		FOREIGN KEY(album_id) REFERENCES album(id)
	);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	return db
}

// TestAddPurchaseSequential tests sequential purchases with stock decrement
func TestAddPurchaseSequential(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test data
	userResult, _ := db.Exec("INSERT INTO user (username, email) VALUES (?, ?)", "testuser", "test@example.com")
	userID, _ := userResult.LastInsertId()

	albumResult, _ := db.Exec("INSERT INTO album (title, artist, price, stock) VALUES (?, ?, ?, ?)", "Album", "Artist", 29.99, 5)
	albumID, _ := albumResult.LastInsertId()

	// First purchase
	purchase1 := models.Purchase{UserID: userID, AlbumID: albumID, Quantity: 2}
	id1, err := repository.AddPurchase(db, purchase1)
	if err != nil {
		t.Fatalf("First purchase failed: %v", err)
	}
	if id1 == 0 {
		t.Error("First purchase should return valid ID")
	}

	// Verify stock was decremented
	var stock int
	db.QueryRow("SELECT stock FROM album WHERE id = ?", albumID).Scan(&stock)
	if stock != 3 {
		t.Errorf("Expected stock 3 after first purchase, got %d", stock)
	}

	// Second purchase
	purchase2 := models.Purchase{UserID: userID, AlbumID: albumID, Quantity: 2}
	id2, err := repository.AddPurchase(db, purchase2)
	if err != nil {
		t.Fatalf("Second purchase failed: %v", err)
	}

	if id2 == 0 {
		t.Error("Second purchase should return valid ID")
	}

	// Verify stock was decremented again
	db.QueryRow("SELECT stock FROM album WHERE id = ?", albumID).Scan(&stock)
	if stock != 1 {
		t.Errorf("Expected stock 1 after second purchase, got %d", stock)
	}
}

// TestAddPurchaseConcurrentOutOfStock tests two concurrent purchases where second fails
func TestAddPurchaseConcurrentOutOfStock(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test data
	userResult, _ := db.Exec("INSERT INTO user (username, email) VALUES (?, ?)", "testuser", "test@example.com")
	userID, _ := userResult.LastInsertId()

	albumResult, _ := db.Exec("INSERT INTO album (title, artist, price, stock) VALUES (?, ?, ?, ?)", "Limited Album", "Artist", 29.99, 3)
	albumID, _ := albumResult.LastInsertId()

	var wg sync.WaitGroup
	var purchase1Err, purchase2Err error

	// First concurrent purchase (should succeed)
	wg.Add(1)
	go func() {
		defer wg.Done()
		purchase := models.Purchase{UserID: userID, AlbumID: albumID, Quantity: 2}
		_, purchase1Err = repository.AddPurchase(db, purchase)
	}()

	// Second concurrent purchase (should fail due to insufficient stock)
	wg.Add(1)
	go func() {
		defer wg.Done()
		purchase := models.Purchase{UserID: userID, AlbumID: albumID, Quantity: 2}
		_, purchase2Err = repository.AddPurchase(db, purchase)
	}()

	wg.Wait()

	// One should succeed, one should fail
	successCount := 0
	if purchase1Err == nil {
		successCount++
	}
	if purchase2Err == nil {
		successCount++
	}

	if successCount != 1 {
		t.Errorf("Expected exactly 1 successful purchase, got %d (err1: %v, err2: %v)", successCount, purchase1Err, purchase2Err)
	}

	// Verify final stock is 1 (only one purchase succeeded)
	var stock int
	db.QueryRow("SELECT stock FROM album WHERE id = ?", albumID).Scan(&stock)
	if stock != 1 {
		t.Errorf("Expected final stock 1, got %d", stock)
	}

	// Verify only one purchase was inserted
	var count int
	db.QueryRow("SELECT COUNT(*) FROM purchase WHERE album_id = ?", albumID).Scan(&count)
	if count != 1 {
		t.Errorf("Expected 1 purchase in database, got %d", count)
	}
}

// TestAddPurchaseOutOfStock tests purchase when stock is insufficient
func TestAddPurchaseOutOfStock(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test data
	userResult, _ := db.Exec("INSERT INTO user (username, email) VALUES (?, ?)", "testuser", "test@example.com")
	userID, _ := userResult.LastInsertId()

	albumResult, _ := db.Exec("INSERT INTO album (title, artist, price, stock) VALUES (?, ?, ?, ?)", "Album", "Artist", 29.99, 1)
	albumID, _ := albumResult.LastInsertId()

	// Try to purchase more than available stock
	purchase := models.Purchase{UserID: userID, AlbumID: albumID, Quantity: 5}
	_, err := repository.AddPurchase(db, purchase)

	if err == nil {
		t.Error("Expected error for out of stock purchase")
	}

	// Verify no purchase was inserted
	var count int
	db.QueryRow("SELECT COUNT(*) FROM purchase").Scan(&count)
	if count != 0 {
		t.Errorf("Expected no purchases, got %d", count)
	}

	// Verify stock unchanged
	var stock int
	db.QueryRow("SELECT stock FROM album WHERE id = ?", albumID).Scan(&stock)
	if stock != 1 {
		t.Errorf("Expected stock 1 (unchanged), got %d", stock)
	}
}
