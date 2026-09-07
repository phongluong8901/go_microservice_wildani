Hướng Dẫn Cấu Hình Google OAuth 2.0
Tài liệu này hướng dẫn các bước thiết lập trên Google Cloud Console để tích hợp tính năng xác thực Google OAuth 2.0 vào hệ thống GoWallet.

1. Cấu hình Màn hình đồng ý OAuth (OAuth Consent Screen)

Truy cập Google Cloud Console - Clients.
https://console.cloud.google.com/auth/clients?hl=vi&project=oauthaitesst123

Nếu hệ thống yêu cầu thiết lập lần đầu, chọn User Type là External.

Nhấn Create và điền các thông tin bắt buộc:

App name: GoWallet

User support email: Email hỗ trợ của dự án.

Developer contact information: Email nhà phát triển.

Nhấn Save and Continue qua các bước Scopes và Test users.

2. Tạo OAuth Client ID

Chuyển đến mục quản lý thông tin xác thực và chọn Create Credentials -> OAuth client ID.

Tại mục Application type, chọn Web application.

Đặt tên cho client (ví dụ: gowallet-web-client).

3. Cấu hình URI chuyển hướng (Authorized redirect URIs)

Tại phần Authorized redirect URIs, nhấn + Add URI.

Thêm đường dẫn callback nhận mã xác thực từ Google trỏ về API server Go:

Môi trường Local: http://localhost:8080/api/v1/auth/google/callback (thay đổi port theo cấu hình thực tế của ứng dụng Gin).

Nhấn Create để lưu lại và nhận Client ID cùng Client Secret.

4. Cấu hình Biến môi trường

Khai báo các thông tin vừa nhận vào file cấu hình .env trong mã nguồn dự án:

Đoạn mã


GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret
GOOGLE_REDIRECT_URL=http://localhost:8080/api/v1/auth/google/callback