--- chuyen sang gitbash

1. Register (Đăng ký tài khoản)

Đăng ký tài khoản mới thành công (201 Created):

Bash


curl -X POST http://localhost:8080/api.v1.users -H "Content-Type: application/json" -d '{"full_name": "Nguyen Van A", "email": "user@example.com", "password": "password123"}'
Đăng ký lỗi trùng email (409 Conflict):

Bash


curl -X POST http://localhost:8080/api.v1.users -H "Content-Type: application/json" -d '{"full_name": "Nguyen Van B", "email": "user@example.com", "password": "password456"}'
Đăng ký lỗi validate dữ liệu/sai định dạng email hoặc mật khẩu quá ngắn (400 Bad Request):

Bash


curl -X POST http://localhost:8080/api.v1.users -H "Content-Type: application/json" -d '{"full_name": "Nguyen Van C", "email": "invalid-email", "password": "123"}'

2. Get Profile (Lấy thông tin cá nhân)

Lấy thông tin user theo ID hợp lệ (200 OK):

Bash


curl http://localhost:8080/api.v1.users/1f13bcec-747e-4602-a990-0c1b7b868595

3. Update Profile (Cập nhật thông tin)

Cập nhật full_name thành công (200 OK):

Bash


curl -X PUT http://localhost:8080/api.v1.users/1f13bcec-747e-4602-a990-0c1b7b868595 -H "Content-Type: application/json" -d '{"full_name": "Nguyen Van Updated"}'