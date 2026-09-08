
-- New database for Auth Service
CREATE DATABASE IF NOT EXISTS `gowallet_auth`;

-- New database for User Service
CREATE DATABASE IF NOT EXISTS `gowallet_user`;

-- New database for Wallet Service
-- CREATE DATABASE IF NOT EXISTS `gowallet_wallet`;

-- New database for Transaction Service
-- CREATE DATABASE IF NOT EXISTS `gowallet_transaction`;

-- New database for Payment Service
-- CREATE DATABASE IF NOT EXISTS `gowallet_payment`;

-- Grant full privileges to user gowallet_user
GRANT ALL PRIVILEGES ON `gowallet_auth`.* TO 'gowallet_user'@'%';
GRANT ALL PRIVILEGES ON `gowallet_user`.* TO 'gowallet_user'@'%';
-- GRANT ALL PRIVILEGES ON `gowallet_wallet`.* TO 'gowallet_user'@'%';
-- GRANT ALL PRIVILEGES ON `gowallet_transaction`.* TO 'gowallet_user'@'%';
-- GRANT ALL PRIVILEGES ON `gowallet_payment`.* TO 'gowallet_user'@'%';
FLUSH PRIVILEGES;



-- CREATE DATABASE IF NOT EXISTS \gowallet_auth`;**: Lệnh tạo một cơ sở dữ liệu mới mang tên gowallet_auth(dùng cho dịch vụ Xác thực). Cụm từIF NOT EXISTS` giúp tránh lỗi nếu cơ sở dữ liệu này đã tồn tại trước đó.
-- GRANT ALL PRIVILEGES ON \gowallet_auth`. TO 'gowallet_user'@'%';**: Cấp toàn bộ quyền quản trị, thao tác (đọc, ghi, sửa, xóa...) trên tất cả các bảng dữ liệu bên trong database gowallet_authcho tài khoản người dùng cơ sở dữ liệu có têngowallet_user. Ký tự %` nghĩa là user này có thể kết nối từ bất kỳ địa chỉ IP nào bên ngoài.
-- FLUSH PRIVILEGES;: Lệnh yêu cầu hệ thống MySQL/MariaDB tải lại và áp dụng ngay lập tức các thay đổi về quyền hạn vừa được cấp mà không cần khởi động lại máy chủ cơ sở dữ liệu.
