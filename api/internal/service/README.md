# Queue Service

## Overview

Queue Service là một hệ thống matchmaking được thiết kế để ghép đôi người dùng trong ứng dụng chat anonymous. Service này sử dụng thiết kế tối ưu với 2 queues riêng biệt cho nam và nữ để đảm bảo performance và fairness.

## Features

### 🚀 Performance Optimized

-   **2 Queues per Category**: Tách riêng queue nam/nữ để matching O(1)
-   **Position Tracking**: O(1) lookup cho user position
-   **Connection-based**: Chỉ match với user còn online
-   **No Expired Time**: User chỉ rời queue khi tắt app

### 📊 Enhanced Analytics

-   **Match Statistics**: Theo dõi số lượng match và thời gian chờ
-   **Queue Statistics**: Thống kê số người trong queue theo giới tính
-   **Estimated Wait Time**: Ước tính thời gian chờ dựa trên balance

### 🔄 Real-time Updates

-   **Position Updates**: Tự động cập nhật position khi có người rời
-   **Connection Tracking**: Theo dõi trạng thái kết nối của user
-   **FIFO Matching**: Đảm bảo fairness cho người vào trước

## Architecture

### Data Structures

```go
// Queue Service
type QueueService struct {
    queueMale   map[string][]*QueueEntry // category -> male entries
    queueFemale map[string][]*QueueEntry // category -> female entries
    userConnections map[uuid.UUID]bool   // userID -> isConnected
    userPositions map[uuid.UUID]*QueuePosition // userID -> position info
    matchStats struct { ... }            // Match statistics
}

// Queue Entry
type QueueEntry struct {
    UserID    uuid.UUID
    Profile   *model.Profile
    Category  string
    JoinedAt  time.Time
    ExpiresAt time.Time // No expiration
    IsMatched bool
    MatchChan chan *MatchResult
}
```

### Flow

1. **User Join Queue**

    - Validate category và user profile
    - Check user không đang trong queue hoặc có room active
    - Add vào queue tương ứng (nam/nữ)
    - Track position và connection status
    - Try match ngay lập tức

2. **Matching Process**

    - Tìm user đầu tiên trong queue đối diện (FIFO)
    - Chỉ match với user còn connected
    - Mark cả 2 là matched
    - Remove khỏi queue và update positions
    - Create room và notify cả 2 users

3. **User Disconnect**
    - Mark user as disconnected
    - Remove khỏi queue
    - Update positions cho các user còn lại
    - Close match channel

## API Endpoints

### Queue Management

-   `POST /queue/join` - Tham gia queue
-   `POST /queue/leave` - Rời khỏi queue
-   `GET /queue/status` - Lấy trạng thái trong queue

### Statistics

-   `GET /queue/stats` - Thống kê queue theo category
-   `GET /queue/match-stats` - Thống kê matching

## Configuration

### Categories

Hiện tại chỉ hỗ trợ category "polite":

```go
const CategoryPolite ChatCategory = "polite"
```

### Wait Time Estimation

-   **Balanced Queue** (|male - female| ≤ 2): ~90 giây
-   **Unbalanced Queue**: 30 giây × số người nhiều hơn
-   **No Matches Possible**: 5 phút

## Usage Example

```go
// Initialize services
db := database.GetDB()
userService := service.NewUserService(db)
roomService := service.NewRoomService(db)
queueService := service.NewQueueService(db, roomService, userService, config)

// Join queue
entry, err := queueService.JoinQueue(ctx, userID, "polite")
if err != nil {
    // Handle error
}

// Wait for match
select {
case result := <-entry.MatchChan:
    if result.Error != nil {
        // Handle match error
    }
    // Handle successful match
    roomID := result.RoomID
    // Start chat...
}

// Get queue status
status, err := queueService.GetQueueStatus(ctx, userID)
if err != nil {
    // Handle error
}
fmt.Printf("Position: %d, Category: %s\n", status.Position, status.Category)

// User disconnect (called when user closes app)
queueService.UserDisconnected(userID)
```

## Benefits

### Performance

-   **O(1) Matching**: Không cần scan toàn bộ queue
-   **O(1) Position Lookup**: Sử dụng map để track position
-   **Memory Efficient**: Chỉ lưu thông tin cần thiết

### User Experience

-   **Fair Matching**: FIFO đảm bảo công bằng
-   **Real-time Position**: User biết chính xác vị trí
-   **No Timeout**: Không bị kick do timeout
-   **Estimated Wait Time**: Biết thời gian chờ dự kiến

### Monitoring

-   **Match Statistics**: Theo dõi performance
-   **Queue Balance**: Phát hiện imbalance
-   **Connection Tracking**: Biết user online/offline

## Future Enhancements

1. **Multiple Categories**: Hỗ trợ nhiều category khác
2. **Advanced Matching**: Matching theo age, location, interests
3. **Priority Queue**: VIP users được ưu tiên
4. **Queue Balancing**: Tự động balance queue
5. **WebSocket Integration**: Real-time updates
6. **Persistence**: Lưu queue state vào database
