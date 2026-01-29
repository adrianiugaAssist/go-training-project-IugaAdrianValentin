package server

import (
	"database/sql"
	"fmt"
	"os"

	"example/data-access/internal/logger"

	"github.com/go-sql-driver/mysql"
)

var db *sql.DB

// InitDatabase initializes the database connection
func InitDatabase() error {
	logger.Log.Debug("Initializing database connection")

	cfg := mysql.NewConfig()
	cfg.User = os.Getenv("DBUSER")
	cfg.Passwd = os.Getenv("DBPASS")
	cfg.Net = "tcp"
	cfg.Addr = "127.0.0.1:3306"
	cfg.DBName = "recordings"

	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		logger.Log.Errorw("Failed to open database", "error", err)
		return fmt.Errorf("failed to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		logger.Log.Errorw("Failed to ping database", "error", err)
		return fmt.Errorf("failed to ping database: %v", err)
	}

	logger.Log.Infow("Database connection established", "database", "recordings", "host", "127.0.0.1:3306")
	return nil
}

// CloseDatabase closes the database connection
func CloseDatabase() error {
	if db != nil {
		logger.Log.Debug("Closing database connection")
		err := db.Close()
		if err != nil {
			logger.Log.Errorw("Error closing database", "error", err)
		} else {
			logger.Log.Info("Database connection closed")
		}
		return err
	}
	return nil
}

// GetDB returns the database instance
func GetDB() *sql.DB {
	return db
}
