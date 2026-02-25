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
		albums, err := repository.GetAllAlbums(db)
		if err != nil {
			logger.Log.Errorw(constants.LogFailedToGetAlbums, "error", err, "remote_addr", clientAddr)
			response = models.WSResponse{Success: false, Error: err.Error()}
		} else {
			duration := time.Since(startTime)
			logger.Log.Infow(constants.LogActionCompletedSuccessfully, "action", msg.Action, "duration_ms", duration.Milliseconds(), "album_count", len(albums), "remote_addr", clientAddr)
			response = models.WSResponse{Success: true, Data: albums}
		}

	case constants.ActionGetAlbumByArtist:
		if artistName, ok := msg.Data.(string); ok {
			if artistName == "" {
				response = models.WSResponse{Success: false, Error: constants.ErrArtistNameEmpty}
				logger.Log.Warnw(constants.LogInvalidRequest, "action", msg.Action, "error", "empty artist name", "remote_addr", clientAddr)
			} else {
				albums, err := repository.GetAlbumsByArtist(db, artistName)
				if err != nil {
					logger.Log.Errorw(constants.LogFailedToGetAlbumsByArtist, "artist", artistName, "error", err, "remote_addr", clientAddr)
					response = models.WSResponse{Success: false, Error: err.Error()}
				} else {
					duration := time.Since(startTime)
					logger.Log.Infow(constants.LogActionCompletedSuccessfully, "action", msg.Action, "duration_ms", duration.Milliseconds(), "artist", artistName, "album_count", len(albums), "remote_addr", clientAddr)
					response = models.WSResponse{Success: true, Data: albums}
				}
			}
		} else {
			response = models.WSResponse{Success: false, Error: constants.ErrArtistNameNotString}
			logger.Log.Warnw(constants.LogInvalidRequest, "action", msg.Action, "error", "artist name not string", "remote_addr", clientAddr)
		}

	case constants.ActionGetAlbumByID:
		var id int64
		if idFloat, ok := msg.Data.(float64); ok {
			id = int64(idFloat)
			if id <= 0 {
				response = models.WSResponse{Success: false, Error: "album " + constants.ErrIDMustBePositive}
				logger.Log.Warnw(constants.LogInvalidRequest, "action", msg.Action, "album_id", id, "error", "invalid ID", "remote_addr", clientAddr)
			} else {
				alb, err := repository.GetAlbumByID(db, id)
				if err != nil {
					logger.Log.Warnw(constants.LogAlbumNotFound, "album_id", id, "error", err, "remote_addr", clientAddr)
					response = models.WSResponse{Success: false, Error: err.Error()}
				} else {
					duration := time.Since(startTime)
					logger.Log.Infow(constants.LogActionCompletedSuccessfully, "action", msg.Action, "duration_ms", duration.Milliseconds(), "album_id", id, "remote_addr", clientAddr)
					response = models.WSResponse{Success: true, Data: alb}
				}
			}
		} else {
			response = models.WSResponse{Success: false, Error: constants.ErrAlbumIDNotNumber}
			logger.Log.Warnw(constants.LogInvalidRequest, "action", msg.Action, "error", "album ID not number", "remote_addr", clientAddr)
		}

	case constants.ActionAddAlbum:
		var newAlbum models.Album
		if data, ok := msg.Data.(map[string]interface{}); ok {
			// Validate required fields
			if title, ok := data[constants.JSONFieldTitle].(string); ok && title != "" {
				newAlbum.Title = title
			} else {
				response = models.WSResponse{Success: false, Error: constants.ErrInvalidOrMissingTitle}
				logger.Log.Warnw(constants.LogInvalidRequest, "action", msg.Action, "error", "missing or empty title", "remote_addr", clientAddr)
				return response
			}
			if artist, ok := data[constants.JSONFieldArtist].(string); ok && artist != "" {
				newAlbum.Artist = artist
			} else {
				response = models.WSResponse{Success: false, Error: constants.ErrInvalidOrMissingArtist}
				logger.Log.Warnw(constants.LogInvalidRequest, "action", msg.Action, "error", "missing or empty artist", "remote_addr", clientAddr)
				return response
			}
			if price, ok := data[constants.JSONFieldPrice].(float64); ok && price > 0 {
				newAlbum.Price = float32(price)
			} else {
				response = models.WSResponse{Success: false, Error: constants.ErrPriceMustBePositive}
				logger.Log.Warnw(constants.LogInvalidRequest, "action", msg.Action, "price", price, "error", "invalid price", "remote_addr", clientAddr)
				return response
			}
			if stock, ok := data[constants.JSONFieldStock].(float64); ok && stock >= 0 {
				newAlbum.Stock = int(stock)
			} else {
				response = models.WSResponse{Success: false, Error: constants.ErrStockMustBeNonNegative}
				logger.Log.Warnw(constants.LogInvalidRequest, "action", msg.Action, "stock", stock, "error", "invalid stock", "remote_addr", clientAddr)
				return response
			}
			id, err := repository.AddAlbum(db, newAlbum)
			if err != nil {
				logger.Log.Errorw(constants.LogFailedToAddAlbum, "title", newAlbum.Title, "artist", newAlbum.Artist, "error", err, "remote_addr", clientAddr)
				response = models.WSResponse{Success: false, Error: err.Error()}
			} else {
				duration := time.Since(startTime)
				logger.Log.Infow(constants.LogActionCompletedSuccessfully, "action", msg.Action, "duration_ms", duration.Milliseconds(), "album_id", id, "title", newAlbum.Title, "remote_addr", clientAddr)
				response = models.WSResponse{Success: true, Data: map[string]interface{}{constants.JSONFieldID: id}}
			}
		} else {
			response = models.WSResponse{Success: false, Error: constants.ErrInvalidAlbumData}
			logger.Log.Warnw(constants.LogInvalidRequest, "action", msg.Action, "error", "album data not object", "remote_addr", clientAddr)
		}

	case constants.ActionGetUsers:
		users, err := repository.GetAllUsers(db)
		if err != nil {
			logger.Log.Errorw(constants.LogFailedToGetUsers, "error", err, "remote_addr", clientAddr)
			response = models.WSResponse{Success: false, Error: err.Error()}
		} else {
			duration := time.Since(startTime)
			logger.Log.Infow(constants.LogActionCompletedSuccessfully, "action", msg.Action, "duration_ms", duration.Milliseconds(), "user_count", len(users), "remote_addr", clientAddr)
			response = models.WSResponse{Success: true, Data: users}
		}

	case constants.ActionGetUserByID:
		var id int64
		if idFloat, ok := msg.Data.(float64); ok {
			id = int64(idFloat)
			if id <= 0 {
				response = models.WSResponse{Success: false, Error: "user " + constants.ErrIDMustBePositive}
				logger.Log.Warnw(constants.LogInvalidRequest, "action", msg.Action, "user_id", id, "error", "invalid ID", "remote_addr", clientAddr)
			} else {
				user, err := repository.GetUserByID(db, id)
				if err != nil {
					logger.Log.Warnw(constants.LogUserNotFound, "user_id", id, "error", err, "remote_addr", clientAddr)
					response = models.WSResponse{Success: false, Error: err.Error()}
				} else {
					duration := time.Since(startTime)
					logger.Log.Infow(constants.LogActionCompletedSuccessfully, "action", msg.Action, "duration_ms", duration.Milliseconds(), "user_id", id, "remote_addr", clientAddr)
					response = models.WSResponse{Success: true, Data: user}
				}
			}
		} else {
			response = models.WSResponse{Success: false, Error: constants.ErrUserIDNotNumber}
			logger.Log.Warnw(constants.LogInvalidRequest, "action", msg.Action, "error", "user ID not number", "remote_addr", clientAddr)
		}

	case constants.ActionAddUser:
		var newUser models.User
		if data, ok := msg.Data.(map[string]interface{}); ok {
			// Validate username
			if username, ok := data[constants.JSONFieldUsername].(string); ok && username != "" {
				newUser.Username = username
			} else {
				response = models.WSResponse{Success: false, Error: constants.ErrInvalidOrMissingUsername}
				logger.Log.Warnw(constants.LogInvalidRequest, "action", msg.Action, "error", "missing or empty username", "remote_addr", clientAddr)
				return response
			}
			// Validate email
			if email, ok := data[constants.JSONFieldEmail].(string); ok && email != "" {
				newUser.Email = email
			} else {
				response = models.WSResponse{Success: false, Error: constants.ErrInvalidOrMissingEmail}
				logger.Log.Warnw(constants.LogInvalidRequest, "action", msg.Action, "error", "missing or empty email", "remote_addr", clientAddr)
				return response
			}
			id, err := repository.AddUser(db, newUser)
			if err != nil {
				logger.Log.Errorw(constants.LogFailedToAddUser, "username", newUser.Username, "error", err, "remote_addr", clientAddr)
				response = models.WSResponse{Success: false, Error: err.Error()}
			} else {
				duration := time.Since(startTime)
				logger.Log.Infow(constants.LogActionCompletedSuccessfully, "action", msg.Action, "duration_ms", duration.Milliseconds(), "user_id", id, "username", newUser.Username, "remote_addr", clientAddr)
				response = models.WSResponse{Success: true, Data: map[string]interface{}{constants.JSONFieldID: id}}
			}
		} else {
			response = models.WSResponse{Success: false, Error: constants.ErrInvalidUserData}
			logger.Log.Warnw(constants.LogInvalidRequest, "action", msg.Action, "error", "user data not object", "remote_addr", clientAddr)
		}

	case constants.ActionGetPurchases:
		purchases, err := repository.GetAllPurchases(db)
		if err != nil {
			logger.Log.Errorw(constants.LogFailedToGetPurchases, "error", err, "remote_addr", clientAddr)
			response = models.WSResponse{Success: false, Error: err.Error()}
		} else {
			duration := time.Since(startTime)
			logger.Log.Infow(constants.LogActionCompletedSuccessfully, "action", msg.Action, "duration_ms", duration.Milliseconds(), "purchase_count", len(purchases), "remote_addr", clientAddr)
			response = models.WSResponse{Success: true, Data: purchases}
		}

	case constants.ActionGetPurchasesByUserID:
		var userID int64
		if userIDFloat, ok := msg.Data.(float64); ok {
			userID = int64(userIDFloat)
			if userID <= 0 {
				response = models.WSResponse{Success: false, Error: "user " + constants.ErrIDMustBePositive}
				logger.Log.Warnw(constants.LogInvalidRequest, "action", msg.Action, "user_id", userID, "error", "invalid ID", "remote_addr", clientAddr)
			} else {
				purchases, err := repository.GetPurchasesByUserID(db, userID)
				if err != nil {
					logger.Log.Errorw(constants.LogFailedToGetPurchasesByUser, "user_id", userID, "error", err, "remote_addr", clientAddr)
					response = models.WSResponse{Success: false, Error: err.Error()}
				} else {
					duration := time.Since(startTime)
					logger.Log.Infow(constants.LogActionCompletedSuccessfully, "action", msg.Action, "duration_ms", duration.Milliseconds(), "user_id", userID, "purchase_count", len(purchases), "remote_addr", clientAddr)
					response = models.WSResponse{Success: true, Data: purchases}
				}
			}
		} else {
			response = models.WSResponse{Success: false, Error: constants.ErrUserIDNotNumber}
			logger.Log.Warnw(constants.LogInvalidRequest, "action", msg.Action, "error", "user ID not number", "remote_addr", clientAddr)
		}

	case constants.ActionAddPurchase:
		var newPurchase models.Purchase
		if data, ok := msg.Data.(map[string]interface{}); ok {
			// Validate user_id
			if userID, ok := data[constants.JSONFieldUserID].(float64); ok && userID > 0 {
				newPurchase.UserID = int64(userID)
			} else {
				response = models.WSResponse{Success: false, Error: constants.ErrInvalidUserIDMustBePositive}
				logger.Log.Warnw(constants.LogInvalidRequest, "action", msg.Action, "user_id", userID, "error", "invalid user_id", "remote_addr", clientAddr)
				return response
			}
			// Validate album_id
			if albumID, ok := data[constants.JSONFieldAlbumID].(float64); ok && albumID > 0 {
				newPurchase.AlbumID = int64(albumID)
			} else {
				response = models.WSResponse{Success: false, Error: constants.ErrInvalidAlbumIDMustBePositive}
				logger.Log.Warnw(constants.LogInvalidRequest, "action", msg.Action, "album_id", albumID, "error", "invalid album_id", "remote_addr", clientAddr)
				return response
			}
			// Validate quantity
			if quantity, ok := data[constants.JSONFieldQuantity].(float64); ok && quantity > 0 {
				newPurchase.Quantity = int(quantity)
			} else {
				response = models.WSResponse{Success: false, Error: constants.ErrInvalidQuantityMustBePositive}
				logger.Log.Warnw(constants.LogInvalidRequest, "action", msg.Action, "quantity", quantity, "error", "invalid quantity", "remote_addr", clientAddr)
				return response
			}
			logger.Log.Infow(constants.LogAttemptingPurchase, "user_id", newPurchase.UserID, "album_id", newPurchase.AlbumID, "quantity", newPurchase.Quantity, "remote_addr", clientAddr)

			id, err := repository.AddPurchase(db, newPurchase)
			if err != nil {
				logger.Log.Warnw(constants.LogPurchaseFailed, "user_id", newPurchase.UserID, "album_id", newPurchase.AlbumID, "quantity", newPurchase.Quantity, "error", err, "remote_addr", clientAddr)
				response = models.WSResponse{Success: false, Error: err.Error()}
			} else {
				duration := time.Since(startTime)
				logger.Log.Infow(constants.LogPurchaseSuccessful, "purchase_id", id, "user_id", newPurchase.UserID, "album_id", newPurchase.AlbumID, "quantity", newPurchase.Quantity, "duration_ms", duration.Milliseconds(), "remote_addr", clientAddr)
				response = models.WSResponse{Success: true, Data: map[string]interface{}{constants.JSONFieldID: id}}
			}
		} else {
			response = models.WSResponse{Success: false, Error: constants.ErrInvalidPurchaseData}
			logger.Log.Warnw(constants.LogInvalidRequest, "action", msg.Action, "error", "purchase data not object", "remote_addr", clientAddr)
		}

	case constants.ActionGetUserPurchaseSummary:
		var userID int64
		if userIDFloat, ok := msg.Data.(float64); ok {
			userID = int64(userIDFloat)
			if userID <= 0 {
				response = models.WSResponse{Success: false, Error: "user " + constants.ErrIDMustBePositive}
				logger.Log.Warnw(constants.LogInvalidRequest, "action", msg.Action, "user_id", userID, "error", "invalid ID", "remote_addr", clientAddr)
			} else {
				summary, err := repository.GetUserPurchaseSummary(db, userID)
				if err != nil {
					logger.Log.Errorw(constants.LogFailedToGetUserPurchaseSummary, "user_id", userID, "error", err, "remote_addr", clientAddr)
					response = models.WSResponse{Success: false, Error: err.Error()}
				} else {
					duration := time.Since(startTime)
					logger.Log.Infow(constants.LogActionCompletedSuccessfully, "action", msg.Action, "duration_ms", duration.Milliseconds(), "user_id", userID, "purchase_count", len(summary.Purchases), "total_cost", summary.TotalCost, "remote_addr", clientAddr)
					response = models.WSResponse{Success: true, Data: summary}
				}
			}
		} else {
			response = models.WSResponse{Success: false, Error: constants.ErrUserIDNotNumber}
			logger.Log.Warnw(constants.LogInvalidRequest, "action", msg.Action, "error", "user ID not number", "remote_addr", clientAddr)
		}

	case constants.ActionGetAllUsersPurchaseSummary:
		summaries, err := repository.GetAllUsersPurchaseSummary(db)
		if err != nil {
			logger.Log.Errorw(constants.LogFailedToGetAllUsersPurchaseSummary, "error", err, "remote_addr", clientAddr)
			response = models.WSResponse{Success: false, Error: err.Error()}
		} else {
			duration := time.Since(startTime)
			logger.Log.Infow(constants.LogActionCompletedSuccessfully, "action", msg.Action, "duration_ms", duration.Milliseconds(), "user_count", len(summaries), "remote_addr", clientAddr)
			response = models.WSResponse{Success: true, Data: summaries}
		}

	default:
		response = models.WSResponse{Success: false, Error: constants.ErrUnknownAction}
		duration := time.Since(startTime)
		logger.Log.Warnw(constants.LogUnknownAction, "action", msg.Action, "duration_ms", duration.Milliseconds(), "remote_addr", clientAddr)
	}
	return response
}
