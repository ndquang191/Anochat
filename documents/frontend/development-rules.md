# Frontend Development Rules - Anochat

## Tech Stack

- **Framework**: Next.js 15, React 19, TypeScript
- **UI Components**: shadcn/ui, Radix UI, Tailwind CSS
- **Real-time**: WebSocket Client
- **State Management**: React Hooks & Contexts
- **Styling**: Tailwind CSS with custom animations

## Project Structure

```
src/
├── app/                    # Next.js app directory
│   ├── (auth)/            # Authentication pages
│   ├── (main)/            # Main application pages
│   └── globals.css        # Global styles
├── components/            # React components
│   ├── ui/               # shadcn/ui components
│   ├── app-sidebar.tsx   # Main sidebar component
│   ├── chat-box.tsx      # Chat interface
│   ├── header.tsx        # Application header
│   ├── login-form.tsx    # Login form
│   └── user-settings-dialog.tsx
├── contexts/             # React contexts
├── hooks/               # Custom React hooks
├── lib/                 # Utility functions
└── middleware.ts        # Next.js middleware
```

## Development Guidelines

### 1. Component Design

- **Keep components small and focused** (max 200 lines)
- **One responsibility per component**
- **Use composition over prop drilling**
- **Prefer server components when possible** (Next.js 15)
- **Use client components** only when needed (hooks, events, state)

### 2. TypeScript Rules

- **Always define types** for props and state
- **Use interfaces** for component props
- **Use type** for unions and utilities
- **Avoid `any`** - use `unknown` if type is truly unknown
- **Define types in separate files** for complex types

### 3. State Management

- **Use React Contexts** for global state
- **Use hooks** for local component state
- **Keep state minimal** - derive when possible
- **Avoid unnecessary re-renders** with `useMemo` and `useCallback`

### 4. Context Structure

```
contexts/
├── auth.tsx          # AuthContext + AuthProvider
├── chat.tsx          # ChatContext + ChatProvider
├── connection.tsx    # ConnectionContext + ConnectionProvider
├── app.tsx           # AppProvider (wrapper)
└── index.ts          # Exports all contexts
```

**Provider Hierarchy:**
```
RootLayout
└── AppProvider (Auth + AlertDialog)
    └── (main)/layout.tsx
        └── ChatAppProvider (Connection + Chat)
            └── Chat components
```

### 5. API Integration

**Authentication:**
- Backend uses HTTP-only cookies for JWT
- Frontend uses `credentials: "include"` for all API calls
- No manual Authorization headers needed

**Example API Call:**
```typescript
const response = await fetch("http://localhost:8080/user/state", {
    credentials: "include",
});
```

### 6. Error Handling

- **Network Errors**: Graceful fallback, user-friendly messages
- **Authentication Errors**: Auto-logout on 401, redirect to login
- **Validation Errors**: Show inline validation messages
- **Log errors** to console in development

### 7. Styling Guidelines

- **Use Tailwind CSS** utility classes
- **Follow mobile-first** approach
- **Use CSS variables** for theming
- **Keep custom CSS minimal** - prefer Tailwind
- **Use shadcn/ui components** for consistency

### 8. WebSocket Management

- **Use singleton pattern** for WebSocket connection
- **Handle reconnection** automatically
- **Clean up listeners** on component unmount
- **Buffer messages** during reconnection
- **Show connection status** to users

### 9. Performance Best Practices

- **Lazy load components** when appropriate
- **Use `next/image`** for images
- **Minimize bundle size** - import only what you need
- **Use server components** for static content
- **Implement proper loading states**

### 10. Code Quality Rules

- **Format with Prettier** before commit
- **Run ESLint** and fix warnings
- **Write meaningful commit messages**
- **Keep functions pure** when possible
- **Avoid side effects** in render

### 11. Naming Conventions

- **Components**: PascalCase (e.g., `ChatBox.tsx`)
- **Hooks**: camelCase with `use` prefix (e.g., `useAuth`)
- **Utilities**: camelCase (e.g., `formatDate`)
- **Constants**: UPPER_SNAKE_CASE (e.g., `API_BASE_URL`)
- **Files**: kebab-case for non-components (e.g., `auth-context.tsx`)

### 12. Testing Guidelines

- **Test user interactions** not implementation
- **Mock API calls** in tests
- **Test error states** and edge cases
- **Keep tests simple** and focused
- **Use React Testing Library** patterns

## Environment Setup

### Development URLs
- **Frontend**: `http://localhost:3000`
- **Backend**: `http://localhost:8080`

### Environment Variables
```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

## Quick Wins for Clean Code

1. **Extract magic strings** to constants
2. **Use early returns** to reduce nesting
3. **Create custom hooks** for reusable logic
4. **Use TypeScript strict mode**
5. **Implement proper loading states**
6. **Add error boundaries** for error handling
7. **Use semantic HTML** elements
8. **Implement proper accessibility** (ARIA labels)

## Code Review Checklist

### Before Submitting
- [ ] TypeScript compiles without errors
- [ ] ESLint passes without warnings
- [ ] Components are properly typed
- [ ] No console errors in browser
- [ ] Responsive design works on mobile

### During Review
- [ ] Components are small and focused
- [ ] Proper error handling
- [ ] Good naming conventions
- [ ] No unnecessary re-renders
- [ ] Accessibility implemented

Remember: **Keep it simple, make it work, then make it better.**
