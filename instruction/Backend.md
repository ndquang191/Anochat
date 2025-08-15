# Backend Specification: Anonymous Chat App

## Overview

This document defines the backend requirements for an anonymous chat application. It outlines all core functionalities that need to be implemented in the NestJS backend. Database schema and implementation are not included here.

---

## ✅ **Backend Specification (Golang version)**

### 1. 🔐 Authentication & Authorization

-   Người dùng đăng nhập bằng Google OAuth2.
-   Sử dụng thư viện `golang.org/x/oauth2` và `google.Endpoint`.
-   Backend sẽ:
    -   Nhận `code` từ client
    -   Gọi Google lấy `access_token`
    -   Lấy thông tin người dùng: `id`, `email`, `name`, `avatar`
-   Nếu người dùng đăng nhập lần đầu:
    -   Lưu thông tin vào bảng `users`
-   Backend tạo và trả lại một **JWT token** (thư viện gợi ý: `github.com/golang-jwt/jwt/v5`)

#### 1.1 JWT Token Validation & User State Check

Sau khi nhận JWT, backend sẽ:

-   Validate JWT token
-   Check xem có phải user mới không
-   Nếu user mới → tạo User với Google ID
-   Nếu user cũ → check xem có đang trong phòng chat nào không
-   Nếu có active room → trả về lịch sử tin nhắn + profile
-   Nếu không có room → chỉ trả về profile

---

### 2. 🙍 User Profile Management

-   Sau khi đăng nhập, người dùng có thể cập nhật thông tin ẩn danh:
    -   `age`, `city`, `gender` (`is_male`)
-   Các trường này lưu vào bảng `profiles` (tách riêng với bảng `users`)
-   Có thể cài đặt `is_hidden` để ẩn thông tin khi chat

---

### 3. 🔁 Matchmaking Queue (In-Memory)

-   Người dùng chọn 1 **chat category** (VD: `polite`)
-   Hàng chờ xử lý bằng bộ nhớ (RAM), không lưu vào DB
-   Ghép đôi dựa trên:
    -   Giới tính
    -   Cùng category
-   Khi match, tạo room (`rooms`) giữa 2 user
-   Gợi ý dùng `map[category][]*UserWaiting` + `sync.Mutex`

---

### 4. 🔌 WebSocket Chat

-   Sau khi match, tạo room và mở WebSocket cho cả hai client
-   Gửi/nhận tin nhắn qua `gorilla/websocket`
-   Tin nhắn được lưu vào bảng `messages`
-   Gợi ý:
    -   Mỗi user có 1 goroutine lắng nghe socket
    -   Gửi tin nhắn qua channel → DB

---

### 5. 🧹 Post-Chat Cleanup Logic

-   Khi cả hai rời phòng:
    -   Hệ thống phân tích toàn bộ tin nhắn trong `messages`
    -   Nếu **có chứa từ khóa nhạy cảm** (ví dụ `hà nội`, `bikini`) → gán `rooms.is_sensitive = true`
    -   Nếu không → xóa toàn bộ tin nhắn của phòng
-   Gợi ý: dùng Go routine chạy cleanup logic sau `room.ended_at`

---

### 6. 🌐 Multi-Tab Support

-   Khi user mở tab mới:
    -   Frontend gửi lại JWT token → Backend xác thực
    -   Backend tự động check user state (active room, messages, profile)
    -   Nếu room đang còn hoạt động → khôi phục kết nối và tiếp tục chat
    -   Nếu không có room → user có thể join queue hoặc setup profile

---

### 7. ⚙️ Config

Tạo file cấu hình Go:

go

CopyEdit

`var ChatConfig = struct { 	Categories        []string 	SensitiveKeywords []string 	MatchTimeoutSec   int }{ 	Categories:        []string{"polite"}, 	SensitiveKeywords: []string{"hà nội", "bikini"}, 	MatchTimeoutSec:   30, }`

---

### 📝 Notes

-   Không có chức năng report
-   Không dùng Supabase SDK → Kết nối PostgreSQL trực tiếp bằng `pgx` hoặc `gorm/sqlc`
-   ORM gợi ý:
    -   Muốn nhanh: `gorm`
    -   Muốn tối ưu và type-safe: `sqlc`
