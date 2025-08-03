# Frontend Backend Integration

## Authentication Flow

### Overview

-   Backend uses HTTP-only cookies for JWT token storage
-   Frontend uses `credentials: "include"` for all API calls
-   No need to manually handle Authorization headers

### Login Process

1. User clicks "Login with Google"
2. Redirected to `/auth/google` (backend)
3. Google OAuth callback handled by backend
4. Backend sets HTTP-only JWT cookie
5. Redirected to frontend `/callback` page
6. Frontend reads user data from temporary cookie
7. User is authenticated

### API Calls

All authenticated API calls should include:

```javascript
{
	credentials: "include"; // Automatically includes HTTP-only JWT cookie
}
```

## API Integration

### User State

```javascript
const response = await fetch("http://localhost:8080/user/state", {
	credentials: "include",
});

if (response.ok) {
	const userState = await response.json();
	// userState.user contains user info
	// userState.user.profile contains profile data
}
```

### Update Profile

```javascript
const response = await fetch("http://localhost:8080/profile", {
	method: "PUT",
	headers: {
		"Content-Type": "application/json",
	},
	credentials: "include",
	body: JSON.stringify({
		age: 25,
		city: "Hanoi",
		is_male: true,
		is_hidden: false,
	}),
});
```

### Logout

```javascript
const response = await fetch("http://localhost:8080/auth/logout", {
	method: "POST",
	credentials: "include",
});
```

## Components Integration

### Auth Context

-   Manages authentication state
-   Handles login/logout
-   Stores user information in client-side cookie

### App Sidebar

-   Loads user data from `/user/state`
-   Handles profile visibility toggle
-   Opens settings dialog

### User Settings Dialog

-   Updates age, gender, city
-   Calls `/profile` endpoint
-   Shows success/error messages

## Error Handling

### Network Errors

-   Graceful fallback to mock data
-   Console logging for debugging
-   User-friendly error messages

### Authentication Errors

-   Automatic logout on 401 responses
-   Redirect to login page
-   Clear local state

## Development Notes

### CORS

-   Backend configured for `http://localhost:3000`
-   Includes credentials in CORS headers

### Cookies

-   JWT token: HTTP-only cookie (secure)
-   User info: Client-side cookie (for UI state)

### Environment

-   Backend: `http://localhost:8080`
-   Frontend: `http://localhost:3000`
