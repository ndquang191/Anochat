# Go Clean Code & Architecture Tips - Chat Project

## Clean Code Principles

### 1. Function Guidelines

-   **Keep functions under 15-20 lines**
-   **Maximum 3 parameters per function**
-   **One responsibility per function**
-   **Use early returns to reduce nesting**
-   **Descriptive function names with action verbs**

### 2. Variable Naming

-   **Short names for short scope**: `i`, `u`, `msg`, `conn`
-   **Descriptive names for longer scope**: `activeUsers`, `roomConnections`
-   **Boolean variables**: `isActive`, `hasPermission`, `canJoin`
-   **Constants**: Use ALL_CAPS for package constants

### 3. Error Handling

-   **Define package-level errors**: `var ErrUserNotFound = errors.New("user not found")`
-   **Wrap errors with context**: `fmt.Errorf("failed to join room: %w", err)`
-   **Check errors immediately after function calls**
-   **Don't ignore errors** - use `_` explicitly if you must

### 4. Struct Design

-   **Group related fields together**
-   **Use struct tags**: `json:"field_name" db:"field_name" validate:"required"`
-   **Keep structs focused and cohesive**
-   **Use composition over embedding when possible**

## Architecture Tips

### 1. Layer Separation

-   **Handler**: HTTP/WebSocket handling only
-   **Service**: Business logic, validation
-   **Repository**: Database operations only
-   **Model**: Data structures and basic methods

### 2. Dependency Flow

```
Handler → Service → Repository → Database
```

-   **Never skip layers**
-   **Handler shouldn't call Repository directly**
-   **Service contains all business rules**

### 3. Interface Usage

-   **Define interfaces at the consumer level**
-   **Keep interfaces small** (1-3 methods)
-   **Use `-er` suffix**: `UserFinder`, `MessageSender`

### 4. Package Organization

-   **One main concept per package**
-   **Keep related functionality together**
-   **Avoid circular dependencies**
-   **Use internal/ for private code**

## Chat-Specific Best Practices

### 1. WebSocket Management

-   **Always handle connection cleanup**
-   **Use goroutines for message broadcasting**
-   **Implement heartbeat/ping mechanism**
-   **Graceful disconnection handling**

### 2. Concurrency Safety

-   **Protect shared data with mutex**
-   **Use channels for goroutine communication**
-   **Avoid sharing memory, share by communicating**
-   **Handle race conditions in user management**

### 3. Message Handling

-   **Validate all incoming messages**
-   **Use message types/enums for clarity**
-   **Implement rate limiting per user**
-   **Log important events for debugging**

### 4. Room Management

-   **Clean up empty rooms automatically**
-   **Limit room size to prevent resource exhaustion**
-   **Handle concurrent join/leave operations**
-   **Track active users efficiently**

## Code Quality Rules

### 1. Essential Rules

-   **Format code with `gofmt` before commit**
-   **Use `go vet` to catch common mistakes**
-   **Handle all errors explicitly**
-   **Write tests for business logic**
-   **Use meaningful commit messages**

### 2. Naming Conventions

-   **Packages**: lowercase, single word when possible
-   **Types**: PascalCase for exported, camelCase for unexported
-   **Functions**: PascalCase for exported, camelCase for unexported
-   **Constants**: ALL_CAPS for package level

### 3. Documentation

-   **Comment exported functions and types**
-   **Explain complex business logic**
-   **Keep comments up-to-date with code**
-   **Use godoc format for public APIs**

## Performance Tips

### 1. Memory Management

-   **Reuse slices and maps when possible**
-   **Use sync.Pool for frequently allocated objects**
-   **Close resources properly with defer**
-   **Avoid memory leaks in goroutines**

### 2. Database Operations

-   **Use prepared statements**
-   **Implement connection pooling**
-   **Use transactions for related operations**
-   **Add proper indexes for queries**

### 3. WebSocket Optimization

-   **Buffer messages before sending**
-   **Use binary messages when appropriate**
-   **Implement message compression**
-   **Limit concurrent connections per user**

## Testing Guidelines

### 1. Test Structure

-   **Use table-driven tests for multiple cases**
-   **Follow AAA pattern**: Arrange, Act, Assert
-   **Test business logic thoroughly**
-   **Mock external dependencies**

### 2. Test Coverage

-   **Focus on business logic over getters/setters**
-   **Test error conditions**
-   **Test concurrent operations**
-   **Keep tests simple and focused**

## Configuration Management

### 1. Environment Variables

-   **Use struct tags for env parsing**
-   **Provide sensible defaults**
-   **Validate configuration on startup**
-   **Group related configs**

### 2. Feature Flags

-   **Use for gradual rollouts**
-   **Keep feature flags simple**
-   **Clean up old flags regularly**

## Monitoring & Logging

### 1. Logging Best Practices

-   **Use structured logging**
-   **Log at appropriate levels**
-   **Include context in log messages**
-   **Don't log sensitive information**

### 2. Metrics to Track

-   **Active connections**
-   **Messages per second**
-   **Room occupancy**
-   **Error rates**

## Code Review Checklist

### Before Submitting

-   [ ] Code formatted and linted
-   [ ] All tests pass
-   [ ] Error handling implemented
-   [ ] No hardcoded values
-   [ ] Proper logging added

### During Review

-   [ ] Functions are small and focused
-   [ ] Proper error handling
-   [ ] No code duplication
-   [ ] Good naming conventions
-   [ ] Concurrent operations are safe

## Quick Wins for Clean Code

1. **Extract magic numbers to constants**
2. **Use early returns to reduce nesting**
3. **Replace long parameter lists with structs**
4. **Add validation at service layer**
5. **Use context for cancellation**
6. **Implement proper error wrapping**
7. **Add comprehensive logging**
8. **Use interfaces for testability**

Remember: **Start simple, refactor as you grow. Clean code is a journey, not a destination.**
