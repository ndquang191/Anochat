# Testing Guide - Anochat Backend

## Overview

This document outlines testing strategies and procedures for the Anochat backend system. While automated tests are not yet implemented, this guide provides comprehensive manual testing flows and guidelines for future test implementation.

---

## Testing Scenarios

### 1. Authentication Flow

#### Test Case: Google OAuth Login
**Objective:** Verify users can successfully log in with Google OAuth

**Steps:**
1. Navigate to `/auth/google`
2. Complete Google OAuth flow
3. Verify redirect to frontend callback
4. Check JWT token is set in HTTP-only cookie
5. Verify `/user/state` returns user data

**Expected Results:**
- ✅ User is redirected to Google OAuth page
- ✅ After authentication, JWT token is set
- ✅ User data is stored in database
- ✅ Frontend receives user information

**Edge Cases:**
- OAuth fails (user denies permission)
- Invalid OAuth state
- Expired OAuth code
- Network errors during token exchange

---

#### Test Case: JWT Token Validation
**Objective:** Verify JWT authentication works correctly

**Steps:**
1. Log in successfully
2. Make request to protected endpoint with valid token
3. Make request with invalid/expired token
4. Make request without token

**Expected Results:**
- ✅ Valid token → 200 OK
- ✅ Invalid token → 401 Unauthorized
- ✅ No token → 401 Unauthorized

---

### 2. Profile Management

#### Test Case: Update Profile
**Objective:** Verify users can update their profile information

**Steps:**
1. Login as authenticated user
2. PUT `/profile` with valid data:
   ```json
   {
     "age": 25,
     "city": "Hanoi",
     "is_male": true,
     "is_hidden": false
   }
   ```
3. Verify response contains updated profile
4. GET `/user/state` to confirm changes persisted

**Expected Results:**
- ✅ Profile updates successfully
- ✅ Changes persist in database
- ✅ Subsequent requests show updated data

**Edge Cases:**
- Invalid age (negative, too large)
- Empty city string
- Null values for optional fields

---

### 3. Queue & Matching System

#### Test Case: Join Queue
**Objective:** Verify users can join matchmaking queue

**Steps:**
1. Login as User A
2. POST `/queue/join` with `{"category": "polite"}`
3. Verify response shows in_queue: true
4. GET `/queue/status` to confirm queue position

**Expected Results:**
- ✅ User added to queue successfully
- ✅ Position in queue is returned
- ✅ User cannot join queue twice (already in queue error)
- ✅ User cannot join queue if already in active room

**Edge Cases:**
- Invalid category
- User already in queue
- User already has active room
- User profile incomplete (no gender set)

---

#### Test Case: Successful Match
**Objective:** Verify two users can be matched successfully

**Steps:**
1. Login as User A (male, polite category)
2. User A joins queue
3. Login as User B (female, polite category) in different browser
4. User B joins queue
5. Verify both users receive `match_found` WebSocket event
6. Check room is created in database
7. Verify both users removed from queue

**Expected Results:**
- ✅ Match found within reasonable time
- ✅ Both users receive match notification
- ✅ Room created with both user IDs
- ✅ Both users can join room via WebSocket
- ✅ Queue positions updated for remaining users

**Matching Priority Tests:**
- Opposite gender priority (Male + Female matched first)
- Same gender fallback (if no opposite gender available)
- Unknown gender users can match with anyone

---

#### Test Case: Leave Queue
**Objective:** Verify users can leave queue before matching

**Steps:**
1. Join queue as User A
2. POST `/queue/leave`
3. Verify success response
4. Confirm user removed from queue (check `/queue/status`)
5. Verify other users' positions updated

**Expected Results:**
- ✅ User successfully leaves queue
- ✅ User can rejoin queue immediately
- ✅ Positions recalculated for remaining users

---

### 4. Real-time Chat (WebSocket)

#### Test Case: WebSocket Connection
**Objective:** Verify WebSocket connection establishes correctly

**Steps:**
1. Login as authenticated user
2. Connect to `/ws` with JWT cookie
3. Verify `connected` event received
4. Send test message to verify bidirectional communication

**Expected Results:**
- ✅ Connection established successfully
- ✅ `connected` event received with user_id
- ✅ Connection authenticated via JWT

**Edge Cases:**
- Connection without authentication
- Multiple concurrent connections from same user
- Connection timeout/disconnect handling

---

#### Test Case: Send and Receive Messages
**Objective:** Verify message sending and receiving works correctly

**Steps:**
1. Match User A and User B
2. User A sends message: `{type: "send_message", payload: {content: "Hello"}}`
3. Verify User B receives `receive_message` event
4. Verify User A receives confirmation `receive_message`
5. Check message saved to database
6. User B sends reply
7. Verify User A receives reply

**Expected Results:**
- ✅ Messages sent successfully
- ✅ Both users receive messages in real-time
- ✅ Messages saved to database with correct metadata
- ✅ Message order preserved
- ✅ Sender receives own message as confirmation

---

#### Test Case: Typing Indicators
**Objective:** Verify typing indicators work correctly

**Steps:**
1. User A starts typing → send `{type: "typing", payload: {is_typing: true}}`
2. Verify User B receives `partner_typing` event
3. User A stops typing → send `{is_typing: false}`
4. Verify User B receives update

**Expected Results:**
- ✅ Typing indicator shows on partner's side
- ✅ Typing stops after timeout or explicit false
- ✅ Sender does not receive own typing indicator

---

#### Test Case: Leave Room
**Objective:** Verify room leaving works correctly

**Steps:**
1. User A and User B in active chat
2. User A sends `{type: "leave_room", payload: {room_id: "xxx"}}`
3. Verify User B receives `partner_left` event
4. Verify room `ended_at` is set in database
5. Verify User A can join queue again immediately

**Expected Results:**
- ✅ Room marked as ended
- ✅ Partner notified of leave
- ✅ User can rejoin queue without errors
- ✅ Post-chat cleanup triggered (sensitive keyword analysis)

---

### 5. Rate Limiting

#### Test Case: API Rate Limiting
**Objective:** Verify rate limiting prevents abuse

**Steps:**
1. Send 100 requests to any endpoint within 1 second
2. Send additional 101st request
3. Verify 429 response received

**Expected Results:**
- ✅ First 100 requests succeed
- ✅ 101st request returns 429
- ✅ After waiting 1 second, requests work again

**Monitoring:**
- Check server logs for rate limit warnings
- Verify IP-based rate limiting works
- Test burst capacity (200 requests)

---

#### Test Case: Message Rate Limiting
**Objective:** Verify message spam prevention

**Steps:**
1. Connect via WebSocket
2. Send 10 messages rapidly
3. Attempt to send 11th message
4. Verify error message received

**Expected Results:**
- ✅ First 10 messages succeed
- ✅ 11th message rejected with error
- ✅ Error message shows: "You are sending messages too quickly"
- ✅ After cooldown, messages work again

---

### 6. Post-Chat Cleanup

#### Test Case: Sensitive Content Detection
**Objective:** Verify rooms with sensitive content are retained

**Steps:**
1. Create chat room with 2 users
2. Send messages containing sensitive keywords (e.g., "hà nội", "bikini")
3. Both users leave room
4. Wait for cleanup goroutine to complete
5. Check database:
   - Room should have `is_sensitive = true`
   - Messages should still exist

**Expected Results:**
- ✅ Room marked as sensitive
- ✅ Messages NOT deleted
- ✅ Room history accessible via `/room/history/:id`

---

#### Test Case: Non-Sensitive Content Cleanup
**Objective:** Verify rooms without sensitive content are deleted

**Steps:**
1. Create chat room with 2 users
2. Send normal messages without sensitive keywords
3. Both users leave room
4. Wait for cleanup goroutine to complete
5. Check database:
   - Room should be deleted
   - Messages should be deleted

**Expected Results:**
- ✅ Room deleted from database
- ✅ All messages deleted
- ✅ No orphaned data

---

### 7. Error Handling & Edge Cases

#### Test Case: Race Conditions
**Scenarios:**

**1. Leave Room → Join Queue Immediately**
- User leaves room
- Immediately joins queue
- Should succeed (no "already has active room" error)

**2. Multiple Queue Join Attempts**
- User joins queue
- User attempts to join queue again
- Should return "already in queue" error

**3. WebSocket Disconnect During Match**
- User A in queue
- User A's WebSocket disconnects
- User B joins queue
- User A should be removed from queue
- User B should not match with disconnected User A

---

#### Test Case: Database Failures
**Scenarios:**

**1. Database Connection Lost**
- Simulate database disconnect
- Attempt operations
- Verify graceful error handling
- Verify reconnection works

**2. Transaction Failures**
- Simulate transaction rollback
- Verify data consistency
- Check error messages

---

## Integration Testing Flows

### Flow 1: Complete User Journey
```
1. User opens app
2. Clicks "Login with Google"
3. Completes OAuth flow
4. Sets profile (age, gender, city)
5. Clicks "Find Chat Partner"
6. Waits in queue
7. Gets matched with another user
8. Sends messages back and forth
9. Sees typing indicators
10. Clicks "Leave Room"
11. Can join queue again
```

**Success Criteria:**
- All steps complete without errors
- Data persists correctly
- UI updates in real-time

---

### Flow 2: Multi-User Concurrent Matching
```
1. 10 users join queue simultaneously
2. Verify all get matched correctly
3. No duplicate matches
4. All matches receive notifications
5. Queue empties completely
```

---

### Flow 3: Reconnection Handling
```
1. User in active chat
2. Close WebSocket connection (simulate disconnect)
3. Wait 5 seconds
4. Reconnect WebSocket
5. Verify chat state restored
6. Verify messages can be sent again
```

---

## Performance Testing

### Load Testing Scenarios

#### 1. Queue Performance
- **Test:** 1000 users join queue simultaneously
- **Metrics:** Match time, memory usage, CPU usage
- **Expected:** < 5 second match time, stable memory

#### 2. WebSocket Scalability
- **Test:** 500 concurrent WebSocket connections
- **Metrics:** Message latency, connection stability
- **Expected:** < 100ms message latency

#### 3. Message Throughput
- **Test:** 100 messages per second across all rooms
- **Metrics:** Database write speed, message delivery time
- **Expected:** All messages delivered within 200ms

---

## Security Testing

### Authentication Security
- ✅ JWT tokens stored in HTTP-only cookies
- ✅ No token leakage in URLs or logs
- ✅ CORS properly configured
- ✅ SQL injection protection (parameterized queries)
- ✅ XSS protection (content sanitization)

### Rate Limiting Security
- ✅ IP-based rate limiting prevents brute force
- ✅ Message rate limiting prevents spam
- ✅ Burst capacity prevents legitimate users from being blocked

---

## Monitoring & Observability

### Key Metrics to Track

#### Application Metrics
- Active WebSocket connections
- Queue wait times
- Match success rate
- Message delivery latency
- API response times
- Rate limit hits per IP

#### Database Metrics
- Query execution times
- Connection pool usage
- Transaction rollback rate
- Database size growth

#### Error Metrics
- 500 error rate
- 401 authentication failures
- 429 rate limit errors
- WebSocket connection failures

### Logging Best Practices

**What to Log:**
- ✅ User authentication events
- ✅ Queue join/leave events
- ✅ Match found events
- ✅ Room creation/end events
- ✅ Rate limit violations
- ✅ Database errors
- ✅ WebSocket connection/disconnection

**What NOT to Log:**
- ❌ JWT tokens
- ❌ User passwords (we don't have any, using OAuth)
- ❌ Message content (privacy)
- ❌ Email addresses in plain text logs

---

## Automated Testing Recommendations

### Unit Tests (To Implement)

**Priority Areas:**
1. **Service Layer**
   - Queue matching algorithm
   - Message rate limiting logic
   - Sensitive keyword detection
   - JWT token generation/validation

2. **Database Operations**
   - User CRUD operations
   - Room management
   - Message storage

3. **Utilities**
   - Token bucket algorithm
   - Message analyzer

### Integration Tests (To Implement)

**Priority Areas:**
1. **API Endpoints**
   - Auth flow end-to-end
   - Queue operations
   - Profile updates

2. **WebSocket**
   - Connection establishment
   - Message broadcasting
   - Room management

3. **Database**
   - Migrations
   - Transaction rollbacks
   - Constraint validation

### Testing Tools Recommendations

**Go Testing Frameworks:**
- `testing` (built-in) - Unit tests
- `testify/assert` - Assertions
- `testify/mock` - Mocking
- `httptest` - HTTP testing
- `gorilla/websocket/test` - WebSocket testing

**Database Testing:**
- `dockertest` - Spin up test database
- `go-sqlmock` - Mock database

**Load Testing:**
- `k6` - Load testing tool
- `vegeta` - HTTP load testing
- `Artillery` - WebSocket load testing

---

## CI/CD Recommendations

### Pre-commit Checks
```bash
# Format code
gofmt -w .

# Run linter
golangci-lint run

# Run unit tests
go test ./...

# Check for vulnerabilities
govulncheck ./...

# Build verification
go build ./cmd/server
```

### Integration Test Pipeline
```yaml
1. Start test database (PostgreSQL)
2. Run migrations
3. Start test server
4. Run integration tests
5. Generate coverage report
6. Cleanup
```

---

## Test Data Management

### Test Users
Create consistent test users:
- User A: Male, age 25, Hanoi
- User B: Female, age 23, Ho Chi Minh
- User C: Male, age 30, Da Nang
- User D: Unknown gender, age 28, Can Tho

### Test Rooms
- Room 1: User A + User B (opposite gender)
- Room 2: User A + User C (same gender fallback)
- Room 3: Sensitive content (contains keywords)
- Room 4: Normal content (clean messages)

---

## Troubleshooting Testing Issues

### Common Issues

**"User already in queue" error:**
- Check database for orphaned queue entries
- Verify queue cleanup is running
- Check for multiple concurrent join attempts

**WebSocket connection fails:**
- Verify JWT token is valid
- Check CORS configuration
- Verify WebSocket upgrader settings

**Messages not received:**
- Check both users are in same room
- Verify WebSocket connection is active
- Check message rate limiting
- Verify database write succeeded

**Match not found:**
- Check queue has users of compatible genders
- Verify queue cleanup isn't removing users
- Check matching algorithm logic

---

## Future Testing Improvements

### Short-term (1-2 months)
- [ ] Implement unit tests for service layer
- [ ] Add integration tests for queue system
- [ ] Create automated API endpoint tests
- [ ] Add WebSocket connection tests

### Medium-term (3-6 months)
- [ ] Implement load testing with k6
- [ ] Add performance benchmarks
- [ ] Create end-to-end test suite
- [ ] Add database migration tests

### Long-term (6+ months)
- [ ] Implement chaos engineering tests
- [ ] Add security penetration tests
- [ ] Create continuous performance monitoring
- [ ] Implement A/B testing framework

---

**Document Version:** 1.0
**Last Updated:** 2026-01-08
**Status:** Production Guidelines
