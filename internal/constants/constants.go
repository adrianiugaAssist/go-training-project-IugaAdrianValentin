package constants

import "time"

// Database Configuration
const (
	DBTimeout = 5 * time.Second
)

// WebSocket Actions
const (
	// Album Actions
	ActionGetAlbums        = "getAlbums"
	ActionGetAlbumByID     = "getAlbumByID"
	ActionGetAlbumByArtist = "getAlbumByArtist"
	ActionAddAlbum         = "addAlbum"

	// User Actions
	ActionGetUsers    = "getUsers"
	ActionGetUserByID = "getUserByID"
	ActionAddUser     = "addUser"

	// Purchase Actions
	ActionGetPurchases               = "getPurchases"
	ActionGetPurchasesByUserID       = "getPurchasesByUserID"
	ActionAddPurchase                = "addPurchase"
	ActionGetUserPurchaseSummary     = "getUserPurchaseSummary"
	ActionGetAllUsersPurchaseSummary = "getAllUsersPurchaseSummary"
)

// Database Table Names
const (
	TableAlbum    = "album"
	TableUser     = "user"
	TablePurchase = "purchase"
)

// Database Column Names - Album
const (
	ColAlbumID     = "id"
	ColAlbumTitle  = "title"
	ColAlbumArtist = "artist"
	ColAlbumPrice  = "price"
	ColAlbumStock  = "stock"
)

// Database Column Names - User
const (
	ColUserID       = "id"
	ColUserUsername = "username"
	ColUserEmail    = "email"
)

// Database Column Names - Purchase
const (
	ColPurchaseID       = "id"
	ColPurchaseUserID   = "user_id"
	ColPurchaseAlbumID  = "album_id"
	ColPurchaseQuantity = "quantity"
)

// JSON Field Names
const (
	JSONFieldTitle    = "title"
	JSONFieldArtist   = "artist"
	JSONFieldPrice    = "price"
	JSONFieldStock    = "stock"
	JSONFieldUsername = "username"
	JSONFieldEmail    = "email"
	JSONFieldUserID   = "user_id"
	JSONFieldAlbumID  = "album_id"
	JSONFieldQuantity = "quantity"
	JSONFieldID       = "id"
)

// Error Messages
const (
	ErrInvalidMessageFormat          = "invalid message format"
	ErrUnknownAction                 = "unknown action"
	ErrArtistNameEmpty               = "artist name cannot be empty"
	ErrArtistNameNotString           = "invalid artist name: must be a string"
	ErrIDMustBePositive              = "ID must be greater than 0"
	ErrAlbumIDNotNumber              = "invalid album ID: must be a number"
	ErrUserIDNotNumber               = "invalid user ID: must be a number"
	ErrInvalidOrMissingTitle         = "invalid or missing title"
	ErrInvalidOrMissingArtist        = "invalid or missing artist"
	ErrPriceMustBePositive           = "price must be greater than 0"
	ErrStockMustBeNonNegative        = "stock must be 0 or greater"
	ErrInvalidAlbumData              = "invalid album data: must be an object"
	ErrInvalidOrMissingUsername      = "invalid or missing username"
	ErrInvalidOrMissingEmail         = "invalid or missing email"
	ErrInvalidUserData               = "invalid user data: must be an object"
	ErrInvalidUserIDMustBePositive   = "invalid or missing user_id: must be greater than 0"
	ErrInvalidAlbumIDMustBePositive  = "invalid or missing album_id: must be greater than 0"
	ErrInvalidQuantityMustBePositive = "invalid quantity: must be greater than 0"
	ErrInvalidPurchaseData           = "invalid purchase data: must be an object"
)

// Log Messages
const (
	LogActionCompletedSuccessfully        = "Action completed successfully"
	LogInvalidRequest                     = "Invalid request"
	LogFailedToGetAlbums                  = "Failed to get albums"
	LogFailedToGetAlbumsByArtist          = "Failed to get albums by artist"
	LogAlbumNotFound                      = "Album not found"
	LogFailedToAddAlbum                   = "Failed to add album"
	LogFailedToGetUsers                   = "Failed to get users"
	LogUserNotFound                       = "User not found"
	LogFailedToAddUser                    = "Failed to add user"
	LogFailedToGetPurchases               = "Failed to get purchases"
	LogFailedToGetPurchasesByUser         = "Failed to get purchases by user"
	LogAttemptingPurchase                 = "Attempting purchase"
	LogPurchaseFailed                     = "Purchase failed"
	LogPurchaseSuccessful                 = "Purchase successful"
	LogFailedToGetUserPurchaseSummary     = "Failed to get user purchase summary"
	LogFailedToGetAllUsersPurchaseSummary = "Failed to get all users purchase summary"
	LogUnknownAction                      = "Unknown action"
)
