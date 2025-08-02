✅ 1. Auth – Đăng nhập
Method	Route	Mô tả
GET	/auth/google	Redirect đến Google OAuth
GET	/auth/callback	Nhận code và xử lý login

✅ 2. User & Profile
Method	Route	Mô tả
GET	/user	Lấy thông tin người dùng hiện tại
GET	/profile/:user_id	Lấy profile ẩn danh của người khác
PUT	/profile	Cập nhật profile ẩn danh người dùng

✅ 3. Room & Chat
Method	Route	Mô tả
POST	/room	Tạo phòng khi ghép đôi thành công (WebSocket gọi nội bộ cũng được)
PUT	/room/leave	Người dùng rời khỏi phòng
PUT	/room/update	Cập nhật trạng thái phòng (VD: có lưu message không)

✅ 4. Message
Method	Route	Mô tả
POST	/message	Lưu tin nhắn mới
PUT	/message/seen	Cập nhật trạng thái đã xem (tuỳ chọn)

✅ 5. WebSocket
Không cần REST, nhưng nên tài liệu rõ endpoint WebSocket:

ws://host/ws (gửi JWT khi kết nối)

Tin nhắn gửi qua socket: send_message, leave_room, v.v.

✅ 6. Health / Internal
Method	Route	Mô tả
GET	/healthz	Kiểm tra server còn sống không

🔍 Bạn có thể thêm:
API	Giải thích
GET /room	Lấy room hiện tại (nếu user đang chat)
GET /room/history/:room_id	Xem lại lịch sử tin nhắn (nếu là room sensitive)

Nếu bạn cần mình viết luôn bản Swagger mô tả các API trên (dưới dạng comment Go), chỉ cần nói nhé.








You said:
tạo 1 phiên bản markdown trước để tôi lưu lại


ChatGPT said:
Dưới đây là phiên bản tài liệu API cho app chat ẩn danh của bạn, viết dưới dạng Markdown, chưa có phần chi tiết input/output – chỉ là mô tả sơ bộ để bạn lưu lại:

📘 Anonymous Chat App – API Specification
Dành cho backend Golang (Gin), dùng Google OAuth2 và WebSocket.

🔐 1. Authentication
Method	Endpoint	Mô tả
GET	/auth/google	Chuyển hướng đến Google OAuth
GET	/auth/callback	Nhận mã từ Google, xử lý đăng nhập

🙍 2. User & Profile
Method	Endpoint	Mô tả
GET	/user	Lấy thông tin người dùng hiện tại
GET	/profile/:user_id	Lấy profile ẩn danh của người khác
PUT	/profile	Cập nhật profile ẩn danh của chính mình

💬 3. Room Management
Method	Endpoint	Mô tả
POST	/room	Tạo room khi ghép đôi thành công
PUT	/room/leave	Người dùng rời khỏi phòng
PUT	/room/update	Cập nhật room (đánh dấu lưu message nếu nhạy cảm)
GET	/room	Lấy room hiện tại của người dùng (nếu có)
GET	/room/history/:id	Lấy lịch sử tin nhắn của room (nếu là sensitive)

✉️ 4. Message
Method	Endpoint	Mô tả
POST	/message	Lưu tin nhắn mới
PUT	/message/seen	Cập nhật trạng thái đã xem (optional)

🔌 5. WebSocket
Endpoint	Mô tả
ws://host/ws	WebSocket endpoint, gửi token JWT khi kết nối

Sự kiện WebSocket có thể bao gồm:

send_message

receive_message

partner_left

you_matched

room_closed

🧪 6. Health Check
Method	Endpoint	Mô tả
GET	/healthz	Kiểm tra server còn sống

📌 Notes
Tất cả các endpoint yêu cầu JWT trừ /auth/*

Hệ thống ghép đôi, queue xử lý bằng WebSocket + RAM

Tất cả các route đều tuân thủ RESTful design

Bạn có thể paste file này vào GitHub README hoặc Notion để quản lý tài liệu. Nếu cần mình thêm phần Request / Response mẫu sau này thì cứ nói nhé.