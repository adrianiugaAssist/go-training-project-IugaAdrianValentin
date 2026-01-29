# Data Access WebSocket API

A Go WebSocket server that provides API endpoints to interact with a MySQL database containing album records.

## Getting Started

### Prerequisites
- Go 1.25.6 or higher
- MySQL database with `recordings` database
- `.env` file with database credentials

### Setup

1. Create a `.env` file in the project root:
```
DBUSER=your_database_user
DBPASS=your_database_password
```

2. Install dependencies:
```bash
go mod download
```

3. Start the server:
```bash
go run .
```

You should see:
```
Connected!
WebSocket server starting on :8080
```
## Run Tests

```bash
go test -v ./internal/tests
```

launch.json config for running tests

```json
        {
            "name": "Debug Tests",
            "type": "go",
            "request": "launch",
            "mode": "test",
            "program": "${workspaceFolder}/internal/tests",
            "args": [
                "-test.v",
                "-test.run", "TestAddPurchaseConcurrentOutOfStock|TestAddPurchaseOutOfStock|TestAddPurchaseSequential"
            ],
            "showLog": true
        }
```

## API Usage

Connect to the WebSocket server at: `ws://localhost:8080/ws`

### Quick WebSocket Messages Reference

**ALBUM OPERATIONS:**
```json
{"action":"getAlbums"}
```

```json
{"action":"getAlbumByArtist","data":"Adele"}
```

```json
{"action":"getAlbumByID","data":1}
```

```json
{"action":"addAlbum","data":{"title":"New Album","artist":"Artist Name","price":29.99,"stock":10}}
```

**USER OPERATIONS:**
```json
{"action":"getUsers"}
```

```json
{"action":"getUserByID","data":1}
```

```json
{"action":"addUser","data":{"username":"john_doe","email":"john@example.com"}}
```

**PURCHASE OPERATIONS:**
```json
{"action":"getPurchases"}
```

```json
{"action":"getPurchasesByUserID","data":1}
```

```json
{"action":"addPurchase","data":{"user_id":1,"album_id":2,"quantity":3}}
```

**PURCHASE SUMMARY OPERATIONS:**
```json
{"action":"getUserPurchaseSummary","data":1}
```

```json
{"action":"getAllUsersPurchaseSummary"}
```

**BATCH OPERATIONS**
```json
[
  {"action":"getAlbums"},
  {"action":"getUsers"},
  {"action":"getUserPurchaseSummary","data":1},
  {"action":"getAllUsersPurchaseSummary"}
]
```
---

### Detailed Message Explanations

#### 1. Get All Albums

**Message:**
```json
{"action":"getAlbums"}
```

**Description:** Retrieves all albums from the database.

**Response Example:**
```json
{
  "success": true,
  "data": [
    {
      "ID": 1,
      "Title": "Album Title",
      "Artist": "Artist Name",
      "Price": 19.99,
      "Stock": 15
    }
  ]
}
```

---

#### 2. Get Albums by Artist

**Message:**
```json
{"action":"getAlbumByArtist","data":"Adele"}
```

**Description:** Retrieves all albums by a specific artist. Replace `"Adele"` with the artist name you want to search for.

**Response Example:**
```json
{
  "success": true,
  "data": [
    {
      "ID": 2,
      "Title": "Hello",
      "Artist": "Adele",
      "Price": 24.99,
      "Stock": 8
    }
  ]
}
```

---

#### 3. Get Album by ID

**Message:**
```json
{"action":"getAlbumByID","data":1}
```

**Description:** Retrieves a specific album by its ID. Replace `1` with the album ID you want to fetch.

**Response Example:**
```json
{
  "success": true,
  "data": {
    "ID": 1,
    "Title": "Album Title",
    "Artist": "Artist Name",
    "Price": 19.99,
    "Stock": 15
  }
}
```

---

#### 4. Add New Album

**Message:**
```json
{"action":"addAlbum","data":{"title":"New Album","artist":"Artist Name","price":29.99,"stock":10}}
```

**Description:** Adds a new album to the database. Provide the album title, artist name, price, and initial stock quantity.

**Response Example:**
```json
{
  "success": true,
  "data": {
    "id": 5
  }
}
```

Returns the ID of the newly created album.

---

#### 5. Get All Users

**Message:**
```json
{"action":"getUsers"}
```

**Description:** Retrieves all users from the database.

**Response Example:**
```json
{
  "success": true,
  "data": [
    {
      "ID": 1,
      "Username": "john_doe",
      "Email": "john@example.com"
    },
    {
      "ID": 2,
      "Username": "jane_smith",
      "Email": "jane@example.com"
    }
  ]
}
```

---

#### 6. Get User by ID

**Message:**
```json
{"action":"getUserByID","data":1}
```

**Description:** Retrieves a specific user by ID. Replace `1` with the user ID you want to fetch.

**Response Example:**
```json
{
  "success": true,
  "data": {
    "ID": 1,
    "Username": "john_doe",
    "Email": "john@example.com"
  }
}
```

---

#### 7. Add New User

**Message:**
```json
{"action":"addUser","data":{"username":"john_doe","email":"john@example.com"}}
```

**Description:** Creates a new user account. Username and email must be unique. Replace the values with actual username and email.

**Response Example:**
```json
{
  "success": true,
  "data": {
    "id": 3
  }
}
```

Returns the ID of the newly created user.

---

#### 8. Get All Purchases

**Message:**
```json
{"action":"getPurchases"}
```

**Description:** Retrieves all purchases from the database.

**Response Example:**
```json
{
  "success": true,
  "data": [
    {
      "ID": 1,
      "UserID": 1,
      "AlbumID": 2,
      "Quantity": 3
    },
    {
      "ID": 2,
      "UserID": 2,
      "AlbumID": 1,
      "Quantity": 1
    }
  ]
}
```

---

#### 9. Get Purchases by User ID

**Message:**
```json
{"action":"getPurchasesByUserID","data":1}
```

**Description:** Retrieves all purchases made by a specific user. Replace `1` with the user ID.

**Response Example:**
```json
{
  "success": true,
  "data": [
    {
      "ID": 1,
      "UserID": 1,
      "AlbumID": 2,
      "Quantity": 3
    },
    {
      "ID": 3,
      "UserID": 1,
      "AlbumID": 5,
      "Quantity": 2
    }
  ]
}
```

---

#### 10. Add New Purchase

**Message:**
```json
{"action":"addPurchase","data":{"user_id":1,"album_id":2,"quantity":3}}
```

**Description:** Records a new purchase. Provide the user ID, album ID, and quantity purchased. The user and album must exist in the database. This operation is transaction-protected and will:
1. Check the current stock level of the album
2. Validate that sufficient stock is available
3. Create the purchase record if stock is sufficient
4. Automatically decrement the album's stock by the purchased quantity

The purchase will fail if the album does not have sufficient stock available.

**Response Example:**
```json
{
  "success": true,
  "data": {
    "id": 5
  }
}
```

Returns the ID of the newly created purchase record.

---

#### 11. Get User Purchase Summary

**Message:**
```json
{"action":"getUserPurchaseSummary","data":1}
```

**Description:** Retrieves a comprehensive summary of a specific user's purchase history, including all purchased albums with their details and a calculated total cost. This function combines user information with their purchases and album details in a single response.

**Response Example:**
```json
{
  "success": true,
  "data": {
    "user_id": 1,
    "username": "john_doe",
    "email": "john@example.com",
    "purchases": [
      {
        "id": 1,
        "album_id": 2,
        "album_title": "Hello",
        "artist": "Adele",
        "price": 24.99,
        "quantity": 2,
        "subtotal": 49.98
      },
      {
        "id": 3,
        "album_id": 5,
        "album_title": "1989",
        "artist": "Taylor Swift",
        "price": 19.99,
        "quantity": 1,
        "subtotal": 19.99
      }
    ],
    "total_cost": 69.97
  }
}
```

**Fields Explanation:**
- `user_id` - The ID of the user
- `username` - The username of the user
- `email` - The email address of the user
- `purchases` - Array of purchase records with album details
  - `id` - Purchase record ID
  - `album_id` - Album ID for this purchase
  - `album_title` - Title of the purchased album
  - `artist` - Artist name of the album
  - `price` - Price per unit of the album
  - `quantity` - Number of units purchased
  - `subtotal` - Price × Quantity
- `total_cost` - Sum of all subtotals for this user

---

#### 12. Get All Users Purchase Summary

**Message:**
```json
{"action":"getAllUsersPurchaseSummary"}
```

**Description:** Retrieves comprehensive purchase summaries for all users in the database. This is useful for reporting, analytics, or generating an overview of all customer purchase activity. The response includes purchase history and total spending for each user.

**Response Example:**
```json
{
  "success": true,
  "data": [
    {
      "user_id": 1,
      "username": "john_doe",
      "email": "john@example.com",
      "purchases": [
        {
          "id": 1,
          "album_id": 2,
          "album_title": "Hello",
          "artist": "Adele",
          "price": 24.99,
          "quantity": 2,
          "subtotal": 49.98
        }
      ],
      "total_cost": 49.98
    },
    {
      "user_id": 2,
      "username": "jane_smith",
      "email": "jane@example.com",
      "purchases": [
        {
          "id": 2,
          "album_id": 1,
          "album_title": "Album Title",
          "artist": "Artist Name",
          "price": 19.99,
          "quantity": 3,
          "subtotal": 59.97
        },
        {
          "id": 4,
          "album_id": 3,
          "album_title": "Rumours",
          "artist": "Fleetwood Mac",
          "price": 17.99,
          "quantity": 1,
          "subtotal": 17.99
        }
      ],
      "total_cost": 77.96
    }
  ]
}
```

**Response Structure:**
- Returns an array of user purchase summaries
- Each element contains the same structure as the single user summary (see #11)
- Sorted by user ID in ascending order
- Includes all users in the database, even those with no purchases (empty `purchases` array)



1. Create a new WebSocket request
2. Enter URL: `ws://localhost:8080/ws`
3. Click **Connect**
4. In the message field at the bottom, paste one of the JSON messages above
5. Click **Send**
6. View the response

## Project Structure

The project follows Go best practices with a clean layered architecture:

```
data-access/
├── main.go                          # Application entry point
├── internal/
│   ├── models/
│   │   └── models.go               # All domain models & WebSocket message types
│   ├── server/
│   │   ├── database.go             # Database connection & management
│   │   └── websocket.go            # WebSocket handlers & request routing
│   └── repository/
│       ├── album.go                # Album database operations
│       ├── user.go                 # User database operations
│       └── purchase.go             # Purchase database operations
├── go.mod                          # Go module definition
├── go.sum                          # Go module checksums
├── .env                            # Environment variables (not in repo)
└── README.md                       # This file
```

### Architecture

- **`main.go`** - Entry point that initializes the database and starts the WebSocket server
- **`internal/models/`** - Data structures for albums, users, purchases, and WebSocket messages
- **`internal/server/`** - Server logic including database management and WebSocket request handlers
- **`internal/repository/`** - Data access layer with functions to query and manipulate database records


## Error Handling

If an action fails, you'll receive:
```json
{
  "success": false,
  "error": "error description"
}
```

## Server Endpoints

- `GET /` - Returns server information
- `GET /ws` - WebSocket connection handler

## Database Schema

The project uses three main tables in the `recordings` MySQL database:

### 1. Album Table
```sql
CREATE TABLE album (
  id INT AUTO_INCREMENT PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  artist VARCHAR(255) NOT NULL,
  price DECIMAL(10, 2) NOT NULL,
  stock INT NOT NULL DEFAULT 0
);
```

**Fields:**
- `id` - Auto-incrementing primary key
- `title` - Album title (required)
- `artist` - Artist name (required)
- `price` - Album price (decimal format)
- `stock` - Quantity available (used for purchase validation)

### 2. User Table
```sql
CREATE TABLE user (
  id INT AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(255) UNIQUE NOT NULL,
  email VARCHAR(255) UNIQUE NOT NULL
);
```

**Fields:**
- `id` - Auto-incrementing primary key
- `username` - Unique username (required)
- `email` - Unique email address (required)

### 3. Purchase Table
```sql
CREATE TABLE purchase (
  id INT AUTO_INCREMENT PRIMARY KEY,
  user_id INT NOT NULL,
  album_id INT NOT NULL,
  quantity INT NOT NULL DEFAULT 1,
  FOREIGN KEY (user_id) REFERENCES user(id),
  FOREIGN KEY (album_id) REFERENCES album(id)
);
```

**Fields:**
- `id` - Auto-incrementing primary key
- `user_id` - References `user` table (required)
- `album_id` - References `album` table (required)
- `quantity` - Number of units purchased (default: 1)

**Important Features:**
- The `purchase` table uses foreign keys to maintain data integrity
- When a purchase is created, the album's stock is automatically decremented
- Purchases are protected by transactions to ensure data consistency
- Stock validation occurs before purchase completion to prevent overselling
