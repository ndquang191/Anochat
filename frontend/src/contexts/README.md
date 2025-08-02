# Contexts

This folder contains all React contexts for the application state management.

## Structure

```
contexts/
├── auth.tsx                  # AuthContext + AuthProvider
├── alert-dialog.tsx          # AlertDialogContext + AlertDialogProvider
├── chat.tsx                  # ChatContext + ChatProvider
├── connection.tsx            # ConnectionContext + ConnectionProvider
├── app.tsx                   # AppProvider (wrapper)
├── chat-app.tsx              # ChatAppProvider (wrapper)
├── index.ts                  # Exports all contexts
└── README.md                 # This file
```

## Usage

### Individual Contexts

```tsx
import { useAuth } from "@/contexts/auth";
import { useAlertDialogContext } from "@/contexts/alert-dialog";
import { useChat } from "@/contexts/chat";
import { useConnection } from "@/contexts/connection";

function MyComponent() {
	const { user, login, logout } = useAuth();
	const { showAlert } = useAlertDialogContext();
	const { messages, sendMessage } = useChat();
	const { isConnected, connect, disconnect } = useConnection();

	// Use context values...
}
```

### Provider Structure

#### Root Provider (`AppProvider`)

Used in the root layout for global contexts:

```tsx
// In layout.tsx
import { AppProvider } from "@/contexts/app";

export default function RootLayout({ children }) {
	return (
		<html>
			<body>
				<AppProvider>{children}</AppProvider>
			</body>
		</html>
	);
}
```

#### Chat App Provider (`ChatAppProvider`)

Used in the main layout for chat-specific contexts:

```tsx
// In (main)/layout.tsx
import { ChatAppProvider } from "@/contexts/chat-app";

export default function MainLayout({ children }) {
	return (
		<ChatAppProvider>
			{/* Chat app content */}
			{children}
		</ChatAppProvider>
	);
}
```

## Context Details

### AuthContext

-   **State**: User authentication status, user info, JWT token
-   **Methods**: `login`, `logout`, `checkAuth`
-   **Hook**: `useAuth()`

### AlertDialogContext

-   **State**: Alert dialog visibility and content
-   **Methods**: `showAlert`, `hideAlert`
-   **Hook**: `useAlertDialogContext()`

### ChatContext

-   **State**: Messages, connection status, current room, partner
-   **Methods**: `sendMessage`, `joinRoom`, `leaveRoom`, `clearMessages`, `setPartner`
-   **Hook**: `useChat()`

### ConnectionContext

-   **State**: WebSocket connection status
-   **Methods**: `connect`, `disconnect`
-   **Hook**: `useConnection()`

## Provider Hierarchy

```
RootLayout
├── AppProvider (Auth + AlertDialog)
    └── (auth)/layout.tsx
        └── (main)/layout.tsx
            └── ChatAppProvider (Connection + Chat)
                └── Chat components
```

## Migration from Hooks

The contexts have been refactored from individual hooks to provide better state management and avoid prop drilling. All existing functionality is preserved with the same API.

## Types

All types are now centralized in `@/types` and imported from there. This ensures consistency across the entire application.
