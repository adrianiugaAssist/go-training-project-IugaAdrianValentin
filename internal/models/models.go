package models

// Album represents an album record in the database
type Album struct {
	ID     int64
	Title  string
	Artist string
	Price  float32
	Stock  int
}

// User represents a user record in the database
type User struct {
	ID       int64
	Username string
	Email    string
}

// Purchase represents a purchase record in the database
type Purchase struct {
	ID       int64
	UserID   int64
	AlbumID  int64
	Quantity int
}

// PurchaseDetail represents purchase information with album details
type PurchaseDetail struct {
	ID         int64   `json:"id"`
	AlbumID    int64   `json:"album_id"`
	AlbumTitle string  `json:"album_title"`
	Artist     string  `json:"artist"`
	Price      float32 `json:"price"`
	Quantity   int     `json:"quantity"`
	Subtotal   float32 `json:"subtotal"`
}

// UserPurchaseSummary represents a user with their purchase history and total cost
type UserPurchaseSummary struct {
	UserID    int64            `json:"user_id"`
	Username  string           `json:"username"`
	Email     string           `json:"email"`
	Purchases []PurchaseDetail `json:"purchases"`
	TotalCost float32          `json:"total_cost"`
}

// WSMessage represents a WebSocket message from the client
type WSMessage struct {
	Action string      `json:"action"`
	Data   interface{} `json:"data"`
}

// WSResponse represents a WebSocket response to the client
type WSResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   string      `json:"error,omitempty"`
}
