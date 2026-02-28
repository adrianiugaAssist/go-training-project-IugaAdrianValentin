package repository

import (
	"context"
	"database/sql"
	"fmt"

	"example/data-access/internal/logger"
	"example/data-access/internal/models"
)

// User database operations

// GetAllUsers calls stored procedure to get all users in the database
func GetAllUsers(db *sql.DB) ([]models.User, error) {
	var users []models.User

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	rows, err := db.QueryContext(ctx, "CALL sp_get_all_users()")
	if err != nil {
		logger.Log.Errorw("Failed to call stored procedure sp_get_all_users", "error", err)
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

// GetUserByID calls stored procedure to get a user with the specified ID
func GetUserByID(db *sql.DB, id int64) (models.User, error) {
	var user models.User

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	row := db.QueryRowContext(ctx, "CALL sp_get_user_by_id(?)", id)
	if err := row.Scan(&user.ID, &user.Username, &user.Email); err != nil {
		logger.Log.Errorw("User not found", "user_id", id, "error", err)
		return user, fmt.Errorf("getUserByID %d: %v", id, err)
	}

	return user, nil
}

// AddUser calls stored procedure to add a user to the database,
// returning the user ID of the new entry
func AddUser(db *sql.DB, user models.User) (int64, error) {
	logger.Log.Infow("Adding new user", "username", user.Username, "email", user.Email)

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var userID int64
	err := db.QueryRowContext(ctx, "CALL sp_add_user(?, ?)", user.Username, user.Email).Scan(&userID)
	if err != nil {
		logger.Log.Errorw("Failed to call stored procedure sp_add_user", "error", err, "username", user.Username)
		return 0, fmt.Errorf("addUser: %v", err)
	}

	logger.Log.Infow("User created", "user_id", userID, "username", user.Username)
	return userID, nil
}
