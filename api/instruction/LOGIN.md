# Login Flow Documentation

## Tổng quan

Ứng dụng sử dụng OAuth 2.0 với Google để xác thực người dùng. Flow login được thiết kế để đảm bảo bảo mật và trải nghiệm người dùng tốt.

## Flow Login

### 1. Khởi tạo Login

-   User truy cập vào trang `/login`
-   Hiển thị form đăng nhập với Google OAuth button
-   User click vào "Login with Google"

### 2. OAuth Redirect

-   Frontend redirect user đến Google OAuth URL
-   Google xác thực user và redirect về backend với authorization code
-   Backend xử lý authorization code và tạo JWT token
-   Backend set JWT token trong HTTP-only cookie và redirect về frontend callback page

### 3. Callback Processing

-   Frontend nhận callback tại `/callback`
-   Frontend gọi API endpoint `/auth/user` để lấy user data
-   Backend trả về user data qua JSON response (token đã có trong HTTP-only cookie)
-   Frontend lưu user data vào auth context
-   Redirect user về trang chính `/`

### 4. Authentication Check

-   Middleware kiểm tra JWT token trong HTTP-only cookies
-   Nếu có token hợp lệ → cho phép truy cập
-   Nếu không có token → redirect về `/login`

## Các Components Tham Gia

### Frontend

-   **Login Page** (`/login`): Hiển thị form đăng nhập
-   **Callback Page** (`/callback`): Xử lý OAuth callback và gọi API
-   **Middleware**: Kiểm tra authentication
-   **Auth Context**: Quản lý trạng thái đăng nhập
-   **Loading Component**: Hiển thị loading state

### Backend

-   **OAuth Handler**: Xử lý Google OAuth flow
-   **JWT Service**: Tạo và validate JWT tokens
-   **User Service**: Quản lý thông tin user
-   **Auth API Endpoint**: `/auth/user` để trả về user data

## Security Features

### JWT Token Storage

-   **HTTP-only Cookies**: Token được lưu trong HTTP-only cookies (không thể truy cập từ JavaScript)
-   **Secure**: Token không xuất hiện trong URL params
-   **Automatic**: Backend tự động set cookie khi xác thực thành công
-   **Expiration**: Token có thời gian hết hạn (7 ngày)

### API Communication

-   **JSON Payload**: User data được truyền qua JSON response thay vì URL params
-   **Credentials**: Frontend gửi credentials để include cookies
-   **CORS**: Backend hỗ trợ CORS với credentials

### Middleware Protection

-   Tất cả routes (trừ public routes) đều được bảo vệ
-   Automatic redirect nếu chưa đăng nhập
-   Preserve original URL để redirect back sau login

### Public Routes

-   `/login`: Trang đăng nhập
-   `/callback`: OAuth callback
-   `/error`: Trang lỗi
-   `/auth/user`: API endpoint để lấy user data

## User Experience

### Loading States

-   Single loading component cho toàn bộ app
-   Hiển thị loading khi đang xác thực
-   Smooth transition giữa các states

### Error Handling

-   Hiển thị error message nếu authentication thất bại
-   Retry mechanism cho user
-   Graceful fallback cho các trường hợp lỗi

### Redirect Flow

-   Sau khi login thành công → redirect về trang chính
-   Nếu user truy cập protected route → redirect về login
-   Sau login → redirect về trang ban đầu

## Data Flow

```
User → Login Page → Google OAuth → Backend → Set HTTP-only Cookie → Callback → API Call → Main App
  ↓
Middleware checks JWT in cookies → Allow/Deny access
  ↓
Auth Context manages state → UI updates
```

## Security Improvements

### Before (Insecure)

-   Token và user data được gửi qua URL params
-   Token có thể bị lộ trong browser history
-   User data hiển thị trong URL
-   Có thể bị log bởi server logs, proxy

### After (Secure)

-   Token được lưu trong HTTP-only cookies
-   User data được truyền qua JSON API response
-   Token không xuất hiện trong URL
-   User data không hiển thị trong URL
-   Bảo vệ khỏi XSS attacks
-   Clean API communication

## API Endpoints

### Authentication Endpoints

-   `GET /auth/google`: Redirect to Google OAuth
-   `GET /auth/callback`: Handle OAuth callback
-   `POST /auth/logout`: Clear JWT token cookie

### Protected Endpoints

-   `GET /user/state`: Get user state (requires JWT)

## Best Practices

1. **Security First**: JWT trong HTTP-only cookies, user data qua JSON API
2. **User Experience**: Single loading state, smooth transitions
3. **Error Handling**: Comprehensive error messages và retry options
4. **Code Organization**: Tách biệt concerns, clean architecture
5. **Performance**: Minimal re-renders, efficient state management
6. **API Design**: RESTful endpoints với proper error handling
