# API Documentation

## Base URL

```
http://localhost:8080
```

## Authentication

### Google OAuth Flow

1. **GET** `/auth/google` - Redirect to Google OAuth
2. **GET** `/auth/callback` - OAuth callback (handled by backend)
3. **POST** `/auth/logout` - Logout and clear cookies

### Authentication Method

-   Uses HTTP-only cookies for JWT token storage
-   Frontend includes `credentials: "include"` in requests
-   No need for Authorization header in most cases

## Protected Endpoints

### User State 

**GET** `/user/state`

Returns current user state including profile information.

**Response:**

```json
{
	"user": {
		"id": "uuid",
		"email": "user@example.com",
		"name": "User Name",
		"avatar_url": "https://...",
		"profile": {
			"age": 25,
			"city": "Ho Chi Minh City",
			"is_male": true,
			"is_hidden": false
		}
	},
	"room": null,
	"messages": null,
	"is_new_user": false
}
```

### Update Profile

**PUT** `/profile`

Update user profile information.

**Request Body:**

```json
{
	"age": 25, // optional, number
	"city": "Hanoi", // optional, string
	"is_male": true, // optional, boolean
	"is_hidden": false // optional, boolean
}
```

**Response:**

```json
{
	"message": "Profile updated successfully",
	"profile": {
		"age": 25,
		"city": "Hanoi",
		"is_male": true,
		"is_hidden": false
	}
}
```

## Health Check

**GET** `/healthz`

Returns server health status.

**Response:**

```json
{
	"status": "ok",
	"message": "Anonymous Chat API is running",
	"database": "connected"
}
```

## Error Responses

### Unauthorized (401)

```json
{
	"error": "Authorization required"
}
```

### Bad Request (400)

```json
{
	"error": "Invalid request body"
}
```

### Internal Server Error (500)

```json
{
	"error": "Error message"
}
```
