package repository

import (
	"database/sql"
	"fmt"

	"example/data-access/internal/logger"
	"example/data-access/internal/models"
)

// User database operations

// GetAllUsers queries for all users in the database
func GetAllUsers(db *sql.DB) ([]models.User, error) {
	var users []models.User

	rows, err := db.Query("SELECT id, username, email FROM user")
	if err != nil {
		logger.Log.Errorw("Failed to query users", "error", err)
		return nil, fmt.Errorf("getAllUsers: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email); err != nil {
			logger.Log.Errorw("Failed to scan user", "error", err)
			return nil, fmt.Errorf("getAllUsers: %v", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		logger.Log.Errorw("Error iterating users", "error", err)
		return nil, fmt.Errorf("getAllUsers: %v", err)
	}

	return users, nil
}

// GetUserByID queries for the user with the specified ID
func GetUserByID(db *sql.DB, id int64) (models.User, error) {
	var user models.User

	row := db.QueryRow("SELECT id, username, email FROM user WHERE id = ?", id)
	if err := row.Scan(&user.ID, &user.Username, &user.Email); err != nil {
		logger.Log.Errorw("User not found", "user_id", id, "error", err)
		return user, fmt.Errorf("getUserByID %d: %v", id, err)
	}

	return user, nil
}

// AddUser adds the specified user to the database,
// returning the user ID of the new entry
func AddUser(db *sql.DB, user models.User) (int64, error) {
	logger.Log.Infow("Adding new user", "username", user.Username, "email", user.Email)

	result, err := db.Exec("INSERT INTO user (username, email) VALUES (?, ?)", user.Username, user.Email)
	if err != nil {
		logger.Log.Errorw("Failed to insert user", "error", err, "username", user.Username)
		return 0, fmt.Errorf("addUser: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		logger.Log.Errorw("Failed to get user ID", "error", err)
		return 0, fmt.Errorf("addUser: %v", err)
	}

	logger.Log.Infow("User created", "user_id", id, "username", user.Username)
	return id, nil
}
