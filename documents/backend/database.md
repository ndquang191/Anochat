## Anonymous Chat App – Database Schema

> Thiết kế này ưu tiên đơn giản, rõ ràng, dễ mở rộng, phục vụ cho hệ thống backend NestJS sử dụng Supabase (PostgreSQL) và Drizzle ORM.

---

### 📍 1. `users`

Lưu thông tin xác thực từ Google login.

```sql
id             UUID (PK)
name           TEXT (nullable)
email          TEXT (nullable)
avatar_url     TEXT (nullable)
is_active      BOOLEAN DEFAULT false
is_deleted     BOOLEAN DEFAULT false
created_at     TIMESTAMP DEFAULT now()
```

---

### 📍 2. `profiles`

Thông tin ẩn danh của người dùng.

```sql
user_id         UUID (PK, FK -> users.id)
is_male         BOOLEAN
age             INTEGER (nullable)
city            TEXT (nullable)
is_hidden   BOOLEAN DEFAULT false
updated_at      TIMESTAMP DEFAULT now()
```

---

### 📍 3. `rooms`

Thông tin phòng chat giữa 2 người.

```sql
id                           UUID (PK)
user1_id                     UUID (FK -> users.id)
user2_id                     UUID (FK -> users.id)
category                     TEXT DEFAULT 'polite'
created_at                   TIMESTAMP DEFAULT now()
ended_at                     TIMESTAMP (nullable)
is_sensitive                 BOOLEAN DEFAULT false
user1_last_read_message_id   UUID (nullable)
user2_last_read_message_id   UUID (nullable)
is_deleted                   BOOLEAN DEFAULT false
```

---

### 📍 4. `messages`

Lưu tin nhắn của từng phòng.

```sql
id           UUID (PK)
room_id      UUID (FK -> rooms.id)
sender_id    UUID (FK -> users.id)
content      TEXT
created_at   TIMESTAMP DEFAULT now()
```

---

### ✅ Ghi chú bổ sung

-   Không có bảng `queues` → hàng chờ được xử lý bằng RAM/WebSocket
-   `is_sensitive` trong `rooms` dùng để đánh dấu phòng nên được giữ lại sau khi phân tích nội dung.
-   `sensitiveKeywords` được định nghĩa trong `chat.config.ts`, không lưu trong DB.
-   Soft delete (`is_deleted`) áp dụng cho các bảng cần khả năng khôi phục hoặc lọc.
-   Enum như `status`, `category` nên định nghĩa rõ trong Drizzle bằng `pg.enum()`.

---

Bạn có thể dùng schema này làm nền tảng để tạo Drizzle migration hoặc generate types.
