package server

import (
	"encoding/json"
	"net/http"

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
			response := models.WSResponse{Success: false, Error: "invalid message format"}
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
	logger.Log.Debugw("Processing action", "action", msg.Action, "remote_addr", clientAddr)
	var response models.WSResponse
	switch msg.Action {
	case "getAlbums":
		albums, err := repository.GetAllAlbums(db)
		if err != nil {
			response = models.WSResponse{Success: false, Error: err.Error()}
		} else {
			response = models.WSResponse{Success: true, Data: albums}
		}

	case "getAlbumByArtist":
		if artistName, ok := msg.Data.(string); ok {
			albums, err := repository.GetAlbumsByArtist(db, artistName)
			if err != nil {
				response = models.WSResponse{Success: false, Error: err.Error()}
			} else {
				response = models.WSResponse{Success: true, Data: albums}
			}
		} else {
			response = models.WSResponse{Success: false, Error: "invalid artist name"}
		}

	case "getAlbumByID":
		var id int64
		if idFloat, ok := msg.Data.(float64); ok {
			id = int64(idFloat)
			alb, err := repository.GetAlbumByID(db, id)
			if err != nil {
				response = models.WSResponse{Success: false, Error: err.Error()}
			} else {
				response = models.WSResponse{Success: true, Data: alb}
			}
		} else {
			response = models.WSResponse{Success: false, Error: "invalid album ID"}
		}

	case "addAlbum":
		var newAlbum models.Album
		if data, ok := msg.Data.(map[string]interface{}); ok {
			if title, ok := data["title"].(string); ok {
				newAlbum.Title = title
			}
			if artist, ok := data["artist"].(string); ok {
				newAlbum.Artist = artist
			}
			if price, ok := data["price"].(float64); ok {
				newAlbum.Price = float32(price)
			}
			if stock, ok := data["stock"].(float64); ok {
				newAlbum.Stock = int(stock)
			}
			id, err := repository.AddAlbum(db, newAlbum)
			if err != nil {
				response = models.WSResponse{Success: false, Error: err.Error()}
			} else {
				response = models.WSResponse{Success: true, Data: map[string]interface{}{"id": id}}
			}
		} else {
			response = models.WSResponse{Success: false, Error: "invalid album data"}
		}

	case "getUsers":
		users, err := repository.GetAllUsers(db)
		if err != nil {
			response = models.WSResponse{Success: false, Error: err.Error()}
		} else {
			response = models.WSResponse{Success: true, Data: users}
		}

	case "getUserByID":
		var id int64
		if idFloat, ok := msg.Data.(float64); ok {
			id = int64(idFloat)
			user, err := repository.GetUserByID(db, id)
			if err != nil {
				response = models.WSResponse{Success: false, Error: err.Error()}
			} else {
				response = models.WSResponse{Success: true, Data: user}
			}
		} else {
			response = models.WSResponse{Success: false, Error: "invalid user ID"}
		}

	case "addUser":
		var newUser models.User
		if data, ok := msg.Data.(map[string]interface{}); ok {
			if username, ok := data["username"].(string); ok {
				newUser.Username = username
			}
			if email, ok := data["email"].(string); ok {
				newUser.Email = email
			}
			id, err := repository.AddUser(db, newUser)
			if err != nil {
				response = models.WSResponse{Success: false, Error: err.Error()}
			} else {
				response = models.WSResponse{Success: true, Data: map[string]interface{}{"id": id}}
			}
		} else {
			response = models.WSResponse{Success: false, Error: "invalid user data"}
		}

	case "getPurchases":
		purchases, err := repository.GetAllPurchases(db)
		if err != nil {
			response = models.WSResponse{Success: false, Error: err.Error()}
		} else {
			response = models.WSResponse{Success: true, Data: purchases}
		}

	case "getPurchasesByUserID":
		var userID int64
		if userIDFloat, ok := msg.Data.(float64); ok {
			userID = int64(userIDFloat)
			purchases, err := repository.GetPurchasesByUserID(db, userID)
			if err != nil {
				response = models.WSResponse{Success: false, Error: err.Error()}
			} else {
				response = models.WSResponse{Success: true, Data: purchases}
			}
		} else {
			response = models.WSResponse{Success: false, Error: "invalid user ID"}
		}

	case "addPurchase":
		var newPurchase models.Purchase
		if data, ok := msg.Data.(map[string]interface{}); ok {
			if userID, ok := data["user_id"].(float64); ok {
				newPurchase.UserID = int64(userID)
			}
			if albumID, ok := data["album_id"].(float64); ok {
				newPurchase.AlbumID = int64(albumID)
			}
			if quantity, ok := data["quantity"].(float64); ok {
				newPurchase.Quantity = int(quantity)
			}
			logger.Log.Infow("Attempting purchase", "user_id", newPurchase.UserID, "album_id", newPurchase.AlbumID, "quantity", newPurchase.Quantity, "remote_addr", clientAddr)

			id, err := repository.AddPurchase(db, newPurchase)
			if err != nil {
				logger.Log.Warnw("Purchase failed", "user_id", newPurchase.UserID, "album_id", newPurchase.AlbumID, "quantity", newPurchase.Quantity, "error", err, "remote_addr", clientAddr)
				response = models.WSResponse{Success: false, Error: err.Error()}
			} else {
				logger.Log.Infow("Purchase successful", "purchase_id", id, "user_id", newPurchase.UserID, "album_id", newPurchase.AlbumID, "quantity", newPurchase.Quantity, "remote_addr", clientAddr)
				response = models.WSResponse{Success: true, Data: map[string]interface{}{"id": id}}
			}
		} else {
			response = models.WSResponse{Success: false, Error: "invalid purchase data"}
		}

	case "getUserPurchaseSummary":
		var userID int64
		if userIDFloat, ok := msg.Data.(float64); ok {
			userID = int64(userIDFloat)
			summary, err := repository.GetUserPurchaseSummary(db, userID)
			if err != nil {
				response = models.WSResponse{Success: false, Error: err.Error()}
			} else {
				response = models.WSResponse{Success: true, Data: summary}
			}
		} else {
			response = models.WSResponse{Success: false, Error: "invalid user ID"}
		}

	case "getAllUsersPurchaseSummary":
		summaries, err := repository.GetAllUsersPurchaseSummary(db)
		if err != nil {
			response = models.WSResponse{Success: false, Error: err.Error()}
		} else {
			response = models.WSResponse{Success: true, Data: summaries}
		}

	default:
		response = models.WSResponse{Success: false, Error: "unknown action"}
		logger.Log.Infow("unknown action", "action", msg.Action, "remote_addr", clientAddr)
	}
	return response
}
