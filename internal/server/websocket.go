package server

import (
	"encoding/json"
	"net/http"
	"time"

	"example/data-access/internal/constants"
	"example/data-access/internal/logger"
	"example/data-access/internal/models"
	"example/data-access/internal/repository"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// HandleWebSocket handles incoming WebSocket connections
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Log.Errorw("WebSocket upgrade error", "error", err, "remote_addr", r.RemoteAddr)
		return
	}
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	logger.Log.Infow("Client connected", "remote_addr", clientAddr)

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Log.Warnw("WebSocket error", "error", err, "remote_addr", clientAddr)
			}
			break
		}

		// Try to unmarshal as an array (batch) of messages first
		var batch []models.WSMessage
		if err := json.Unmarshal(p, &batch); err == nil && len(batch) > 0 {
			var responses []models.WSResponse
			for _, m := range batch {
				responses = append(responses, handleMessage(m, clientAddr))
			}
			if err := conn.WriteJSON(responses); err != nil {
				logger.Log.Errorw("Write error", "error", err, "remote_addr", clientAddr)
				break
			}
			continue
		}

		// Otherwise, try single message
		var msg models.WSMessage
		if err := json.Unmarshal(p, &msg); err != nil {
			logger.Log.Warnw("Invalid message format", "remote_addr", clientAddr, "error", err)
			response := models.WSResponse{Success: false, Error: constants.ErrInvalidMessageFormat}
			if err := conn.WriteJSON(response); err != nil {
				logger.Log.Errorw("Write error", "error", err, "remote_addr", clientAddr)
				break
			}
			continue
		}

		response := handleMessage(msg, clientAddr)
		if err := conn.WriteJSON(response); err != nil {
			logger.Log.Errorw("Write error", "error", err, "remote_addr", clientAddr)
			break
		}
	}

	logger.Log.Infow("Client disconnected", "remote_addr", clientAddr)
}

// handleMessage processes a single WSMessage and returns a WSResponse
func handleMessage(msg models.WSMessage, clientAddr string) models.WSResponse {
	startTime := time.Now()
	logger.Log.Debugw("Processing action", "action", msg.Action, "remote_addr", clientAddr)

	var response models.WSResponse
	switch msg.Action {
	case constants.ActionGetAlbums:
		response = handleGetAlbums(startTime, clientAddr)
	case constants.ActionGetAlbumByArtist:
		response = handleGetAlbumByArtist(msg.Data, startTime, clientAddr)
	case constants.ActionGetAlbumByID:
		response = handleGetAlbumByID(msg.Data, startTime, clientAddr)
	case constants.ActionAddAlbum:
		response = handleAddAlbum(msg.Data, startTime, clientAddr)
	case constants.ActionGetUsers:
		response = handleGetUsers(startTime, clientAddr)
	case constants.ActionGetUserByID:
		response = handleGetUserByID(msg.Data, startTime, clientAddr)
	case constants.ActionAddUser:
		response = handleAddUser(msg.Data, startTime, clientAddr)
	case constants.ActionGetPurchases:
		response = handleGetPurchases(startTime, clientAddr)
	case constants.ActionGetPurchasesByUserID:
		response = handleGetPurchasesByUserID(msg.Data, startTime, clientAddr)
	case constants.ActionAddPurchase:
		response = handleAddPurchase(msg.Data, startTime, clientAddr)
	case constants.ActionGetUserPurchaseSummary:
		response = handleGetUserPurchaseSummary(msg.Data, startTime, clientAddr)
	case constants.ActionGetAllUsersPurchaseSummary:
		response = handleGetAllUsersPurchaseSummary(startTime, clientAddr)
	default:
		response = models.WSResponse{Success: false, Error: constants.ErrUnknownAction}
		duration := time.Since(startTime)
		logger.Log.Warnw(constants.LogUnknownAction, "action", msg.Action, "duration_ms", duration.Milliseconds(), "remote_addr", clientAddr)
	}
	return response
}

// handleGetAlbums retrieves all albums from the database
func handleGetAlbums(startTime time.Time, clientAddr string) models.WSResponse {
	albums, err := repository.GetAllAlbums(db)
	if err != nil {
		logger.Log.Errorw(constants.LogFailedToGetAlbums, "error", err, "remote_addr", clientAddr)
		return models.WSResponse{Success: false, Error: err.Error()}
	}

	duration := time.Since(startTime)
	logger.Log.Infow(constants.LogActionCompletedSuccessfully, "action", constants.ActionGetAlbums, "duration_ms", duration.Milliseconds(), "album_count", len(albums), "remote_addr", clientAddr)
	return models.WSResponse{Success: true, Data: albums}
}

// handleGetAlbumByArtist retrieves albums by a specific artist
func handleGetAlbumByArtist(data interface{}, startTime time.Time, clientAddr string) models.WSResponse {
	artistName, ok := data.(string)
	if !ok {
		logger.Log.Warnw(constants.LogInvalidRequest, "action", constants.ActionGetAlbumByArtist, "error", "artist name not string", "remote_addr", clientAddr)
		return models.WSResponse{Success: false, Error: constants.ErrArtistNameNotString}
	}

	if artistName == "" {
		logger.Log.Warnw(constants.LogInvalidRequest, "action", constants.ActionGetAlbumByArtist, "error", "empty artist name", "remote_addr", clientAddr)
		return models.WSResponse{Success: false, Error: constants.ErrArtistNameEmpty}
	}

	albums, err := repository.GetAlbumsByArtist(db, artistName)
	if err != nil {
		logger.Log.Errorw(constants.LogFailedToGetAlbumsByArtist, "artist", artistName, "error", err, "remote_addr", clientAddr)
		return models.WSResponse{Success: false, Error: err.Error()}
	}

	duration := time.Since(startTime)
	logger.Log.Infow(constants.LogActionCompletedSuccessfully, "action", constants.ActionGetAlbumByArtist, "duration_ms", duration.Milliseconds(), "artist", artistName, "album_count", len(albums), "remote_addr", clientAddr)
	return models.WSResponse{Success: true, Data: albums}
}

// handleGetAlbumByID retrieves a specific album by ID
func handleGetAlbumByID(data interface{}, startTime time.Time, clientAddr string) models.WSResponse {
	idFloat, ok := data.(float64)
	if !ok {
		logger.Log.Warnw(constants.LogInvalidRequest, "action", constants.ActionGetAlbumByID, "error", "album ID not number", "remote_addr", clientAddr)
		return models.WSResponse{Success: false, Error: constants.ErrAlbumIDNotNumber}
	}

	id := int64(idFloat)
	if id <= 0 {
		logger.Log.Warnw(constants.LogInvalidRequest, "action", constants.ActionGetAlbumByID, "album_id", id, "error", "invalid ID", "remote_addr", clientAddr)
		return models.WSResponse{Success: false, Error: "album " + constants.ErrIDMustBePositive}
	}

	alb, err := repository.GetAlbumByID(db, id)
	if err != nil {
		logger.Log.Warnw(constants.LogAlbumNotFound, "album_id", id, "error", err, "remote_addr", clientAddr)
		return models.WSResponse{Success: false, Error: err.Error()}
	}

	duration := time.Since(startTime)
	logger.Log.Infow(constants.LogActionCompletedSuccessfully, "action", constants.ActionGetAlbumByID, "duration_ms", duration.Milliseconds(), "album_id", id, "remote_addr", clientAddr)
	return models.WSResponse{Success: true, Data: alb}
}

// handleAddAlbum adds a new album to the database
func handleAddAlbum(data interface{}, startTime time.Time, clientAddr string) models.WSResponse {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		logger.Log.Warnw(constants.LogInvalidRequest, "action", constants.ActionAddAlbum, "error", "album data not object", "remote_addr", clientAddr)
		return models.WSResponse{Success: false, Error: constants.ErrInvalidAlbumData}
	}

	var newAlbum models.Album

	// Validate title
	if title, ok := dataMap[constants.JSONFieldTitle].(string); ok && title != "" {
		newAlbum.Title = title
	} else {
		logger.Log.Warnw(constants.LogInvalidRequest, "action", constants.ActionAddAlbum, "error", "missing or empty title", "remote_addr", clientAddr)
		return models.WSResponse{Success: false, Error: constants.ErrInvalidOrMissingTitle}
	}

	// Validate artist
	if artist, ok := dataMap[constants.JSONFieldArtist].(string); ok && artist != "" {
		newAlbum.Artist = artist
	} else {
		logger.Log.Warnw(constants.LogInvalidRequest, "action", constants.ActionAddAlbum, "error", "missing or empty artist", "remote_addr", clientAddr)
		return models.WSResponse{Success: false, Error: constants.ErrInvalidOrMissingArtist}
	}

	// Validate price
	if price, ok := dataMap[constants.JSONFieldPrice].(float64); ok && price > 0 {
		newAlbum.Price = float32(price)
	} else {
		logger.Log.Warnw(constants.LogInvalidRequest, "action", constants.ActionAddAlbum, "price", dataMap[constants.JSONFieldPrice], "error", "invalid price", "remote_addr", clientAddr)
		return models.WSResponse{Success: false, Error: constants.ErrPriceMustBePositive}
	}

	// Validate stock
	if stock, ok := dataMap[constants.JSONFieldStock].(float64); ok && stock >= 0 {
		newAlbum.Stock = int(stock)
	} else {
		logger.Log.Warnw(constants.LogInvalidRequest, "action", constants.ActionAddAlbum, "stock", dataMap[constants.JSONFieldStock], "error", "invalid stock", "remote_addr", clientAddr)
		return models.WSResponse{Success: false, Error: constants.ErrStockMustBeNonNegative}
	}

	id, err := repository.AddAlbum(db, newAlbum)
	if err != nil {
		logger.Log.Errorw(constants.LogFailedToAddAlbum, "title", newAlbum.Title, "artist", newAlbum.Artist, "error", err, "remote_addr", clientAddr)
		return models.WSResponse{Success: false, Error: err.Error()}
	}

	duration := time.Since(startTime)
	logger.Log.Infow(constants.LogActionCompletedSuccessfully, "action", constants.ActionAddAlbum, "duration_ms", duration.Milliseconds(), "album_id", id, "title", newAlbum.Title, "remote_addr", clientAddr)
	return models.WSResponse{Success: true, Data: map[string]interface{}{constants.JSONFieldID: id}}
}

// handleGetUsers retrieves all users from the database
func handleGetUsers(startTime time.Time, clientAddr string) models.WSResponse {
	users, err := repository.GetAllUsers(db)
	if err != nil {
		logger.Log.Errorw(constants.LogFailedToGetUsers, "error", err, "remote_addr", clientAddr)
		return models.WSResponse{Success: false, Error: err.Error()}
	}

	duration := time.Since(startTime)
	logger.Log.Infow(constants.LogActionCompletedSuccessfully, "action", constants.ActionGetUsers, "duration_ms", duration.Milliseconds(), "user_count", len(users), "remote_addr", clientAddr)
	return models.WSResponse{Success: true, Data: users}
}

// handleGetUserByID retrieves a specific user by ID
func handleGetUserByID(data interface{}, startTime time.Time, clientAddr string) models.WSResponse {
	idFloat, ok := data.(float64)
	if !ok {
		logger.Log.Warnw(constants.LogInvalidRequest, "action", constants.ActionGetUserByID, "error", "user ID not number", "remote_addr", clientAddr)
		return models.WSResponse{Success: false, Error: constants.ErrUserIDNotNumber}
	}

	id := int64(idFloat)
	if id <= 0 {
		logger.Log.Warnw(constants.LogInvalidRequest, "action", constants.ActionGetUserByID, "user_id", id, "error", "invalid ID", "remote_addr", clientAddr)
		return models.WSResponse{Success: false, Error: "user " + constants.ErrIDMustBePositive}
	}

	user, err := repository.GetUserByID(db, id)
	if err != nil {
		logger.Log.Warnw(constants.LogUserNotFound, "user_id", id, "error", err, "remote_addr", clientAddr)
		return models.WSResponse{Success: false, Error: err.Error()}
	}

	duration := time.Since(startTime)
	logger.Log.Infow(constants.LogActionCompletedSuccessfully, "action", constants.ActionGetUserByID, "duration_ms", duration.Milliseconds(), "user_id", id, "remote_addr", clientAddr)
	return models.WSResponse{Success: true, Data: user}
}

// handleAddUser adds a new user to the database
func handleAddUser(data interface{}, startTime time.Time, clientAddr string) models.WSResponse {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		logger.Log.Warnw(constants.LogInvalidRequest, "action", constants.ActionAddUser, "error", "user data not object", "remote_addr", clientAddr)
		return models.WSResponse{Success: false, Error: constants.ErrInvalidUserData}
	}

	var newUser models.User

	// Validate username
	if username, ok := dataMap[constants.JSONFieldUsername].(string); ok && username != "" {
		newUser.Username = username
	} else {
		logger.Log.Warnw(constants.LogInvalidRequest, "action", constants.ActionAddUser, "error", "missing or empty username", "remote_addr", clientAddr)
		return models.WSResponse{Success: false, Error: constants.ErrInvalidOrMissingUsername}
	}

	// Validate email
	if email, ok := dataMap[constants.JSONFieldEmail].(string); ok && email != "" {
		newUser.Email = email
	} else {
		logger.Log.Warnw(constants.LogInvalidRequest, "action", constants.ActionAddUser, "error", "missing or empty email", "remote_addr", clientAddr)
		return models.WSResponse{Success: false, Error: constants.ErrInvalidOrMissingEmail}
	}

	id, err := repository.AddUser(db, newUser)
	if err != nil {
		logger.Log.Errorw(constants.LogFailedToAddUser, "username", newUser.Username, "error", err, "remote_addr", clientAddr)
		return models.WSResponse{Success: false, Error: err.Error()}
	}

	duration := time.Since(startTime)
	logger.Log.Infow(constants.LogActionCompletedSuccessfully, "action", constants.ActionAddUser, "duration_ms", duration.Milliseconds(), "user_id", id, "username", newUser.Username, "remote_addr", clientAddr)
	return models.WSResponse{Success: true, Data: map[string]interface{}{constants.JSONFieldID: id}}
}

// handleGetPurchases retrieves all purchases from the database
func handleGetPurchases(startTime time.Time, clientAddr string) models.WSResponse {
	purchases, err := repository.GetAllPurchases(db)
	if err != nil {
		logger.Log.Errorw(constants.LogFailedToGetPurchases, "error", err, "remote_addr", clientAddr)
		return models.WSResponse{Success: false, Error: err.Error()}
	}

	duration := time.Since(startTime)
	logger.Log.Infow(constants.LogActionCompletedSuccessfully, "action", constants.ActionGetPurchases, "duration_ms", duration.Milliseconds(), "purchase_count", len(purchases), "remote_addr", clientAddr)
	return models.WSResponse{Success: true, Data: purchases}
}

// handleGetPurchasesByUserID retrieves purchases for a specific user
func handleGetPurchasesByUserID(data interface{}, startTime time.Time, clientAddr string) models.WSResponse {
	userIDFloat, ok := data.(float64)
	if !ok {
		logger.Log.Warnw(constants.LogInvalidRequest, "action", constants.ActionGetPurchasesByUserID, "error", "user ID not number", "remote_addr", clientAddr)
		return models.WSResponse{Success: false, Error: constants.ErrUserIDNotNumber}
	}

	userID := int64(userIDFloat)
	if userID <= 0 {
		logger.Log.Warnw(constants.LogInvalidRequest, "action", constants.ActionGetPurchasesByUserID, "user_id", userID, "error", "invalid ID", "remote_addr", clientAddr)
		return models.WSResponse{Success: false, Error: "user " + constants.ErrIDMustBePositive}
	}

	purchases, err := repository.GetPurchasesByUserID(db, userID)
	if err != nil {
		logger.Log.Errorw(constants.LogFailedToGetPurchasesByUser, "user_id", userID, "error", err, "remote_addr", clientAddr)
		return models.WSResponse{Success: false, Error: err.Error()}
	}

	duration := time.Since(startTime)
	logger.Log.Infow(constants.LogActionCompletedSuccessfully, "action", constants.ActionGetPurchasesByUserID, "duration_ms", duration.Milliseconds(), "user_id", userID, "purchase_count", len(purchases), "remote_addr", clientAddr)
	return models.WSResponse{Success: true, Data: purchases}
}

// handleAddPurchase adds a new purchase to the database
func handleAddPurchase(data interface{}, startTime time.Time, clientAddr string) models.WSResponse {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		logger.Log.Warnw(constants.LogInvalidRequest, "action", constants.ActionAddPurchase, "error", "purchase data not object", "remote_addr", clientAddr)
		return models.WSResponse{Success: false, Error: constants.ErrInvalidPurchaseData}
	}

	var newPurchase models.Purchase

	// Validate user_id
	if userID, ok := dataMap[constants.JSONFieldUserID].(float64); ok && userID > 0 {
		newPurchase.UserID = int64(userID)
	} else {
		logger.Log.Warnw(constants.LogInvalidRequest, "action", constants.ActionAddPurchase, "user_id", dataMap[constants.JSONFieldUserID], "error", "invalid user_id", "remote_addr", clientAddr)
		return models.WSResponse{Success: false, Error: constants.ErrInvalidUserIDMustBePositive}
	}

	// Validate album_id
	if albumID, ok := dataMap[constants.JSONFieldAlbumID].(float64); ok && albumID > 0 {
		newPurchase.AlbumID = int64(albumID)
	} else {
		logger.Log.Warnw(constants.LogInvalidRequest, "action", constants.ActionAddPurchase, "album_id", dataMap[constants.JSONFieldAlbumID], "error", "invalid album_id", "remote_addr", clientAddr)
		return models.WSResponse{Success: false, Error: constants.ErrInvalidAlbumIDMustBePositive}
	}

	// Validate quantity
	if quantity, ok := dataMap[constants.JSONFieldQuantity].(float64); ok && quantity > 0 {
		newPurchase.Quantity = int(quantity)
	} else {
		logger.Log.Warnw(constants.LogInvalidRequest, "action", constants.ActionAddPurchase, "quantity", dataMap[constants.JSONFieldQuantity], "error", "invalid quantity", "remote_addr", clientAddr)
		return models.WSResponse{Success: false, Error: constants.ErrInvalidQuantityMustBePositive}
	}

	logger.Log.Infow(constants.LogAttemptingPurchase, "user_id", newPurchase.UserID, "album_id", newPurchase.AlbumID, "quantity", newPurchase.Quantity, "remote_addr", clientAddr)

	id, err := repository.AddPurchase(db, newPurchase)
	if err != nil {
		logger.Log.Warnw(constants.LogPurchaseFailed, "user_id", newPurchase.UserID, "album_id", newPurchase.AlbumID, "quantity", newPurchase.Quantity, "error", err, "remote_addr", clientAddr)
		return models.WSResponse{Success: false, Error: err.Error()}
	}

	duration := time.Since(startTime)
	logger.Log.Infow(constants.LogPurchaseSuccessful, "purchase_id", id, "user_id", newPurchase.UserID, "album_id", newPurchase.AlbumID, "quantity", newPurchase.Quantity, "duration_ms", duration.Milliseconds(), "remote_addr", clientAddr)
	return models.WSResponse{Success: true, Data: map[string]interface{}{constants.JSONFieldID: id}}
}

// handleGetUserPurchaseSummary retrieves purchase summary for a specific user
func handleGetUserPurchaseSummary(data interface{}, startTime time.Time, clientAddr string) models.WSResponse {
	userIDFloat, ok := data.(float64)
	if !ok {
		logger.Log.Warnw(constants.LogInvalidRequest, "action", constants.ActionGetUserPurchaseSummary, "error", "user ID not number", "remote_addr", clientAddr)
		return models.WSResponse{Success: false, Error: constants.ErrUserIDNotNumber}
	}

	userID := int64(userIDFloat)
	if userID <= 0 {
		logger.Log.Warnw(constants.LogInvalidRequest, "action", constants.ActionGetUserPurchaseSummary, "user_id", userID, "error", "invalid ID", "remote_addr", clientAddr)
		return models.WSResponse{Success: false, Error: "user " + constants.ErrIDMustBePositive}
	}

	summary, err := repository.GetUserPurchaseSummary(db, userID)
	if err != nil {
		logger.Log.Errorw(constants.LogFailedToGetUserPurchaseSummary, "user_id", userID, "error", err, "remote_addr", clientAddr)
		return models.WSResponse{Success: false, Error: err.Error()}
	}

	duration := time.Since(startTime)
	logger.Log.Infow(constants.LogActionCompletedSuccessfully, "action", constants.ActionGetUserPurchaseSummary, "duration_ms", duration.Milliseconds(), "user_id", userID, "purchase_count", len(summary.Purchases), "total_cost", summary.TotalCost, "remote_addr", clientAddr)
	return models.WSResponse{Success: true, Data: summary}
}

// handleGetAllUsersPurchaseSummary retrieves purchase summaries for all users
func handleGetAllUsersPurchaseSummary(startTime time.Time, clientAddr string) models.WSResponse {
	summaries, err := repository.GetAllUsersPurchaseSummary(db)
	if err != nil {
		logger.Log.Errorw(constants.LogFailedToGetAllUsersPurchaseSummary, "error", err, "remote_addr", clientAddr)
		return models.WSResponse{Success: false, Error: err.Error()}
	}

	duration := time.Since(startTime)
	logger.Log.Infow(constants.LogActionCompletedSuccessfully, "action", constants.ActionGetAllUsersPurchaseSummary, "duration_ms", duration.Milliseconds(), "user_count", len(summaries), "remote_addr", clientAddr)
	return models.WSResponse{Success: true, Data: summaries}
}
