ALTER TABLE users DROP COLUMN avatar_url;

-- Tiến lên (Up Migration - Version 2): Khi bạn đánh số phiên bản là version 2, hệ thống quản lý migration (như golang-migrate) sẽ ghi nhận đây là bước tiếp theo sau version 1. Câu lệnh ALTER TABLE users ADD COLUMN avatar_url VARCHAR(255) NULL AFTER password_hash; được đặt trong file migration up, dùng để tiến hành thay đổi cấu trúc bảng bằng cách thêm một trường dữ liệu mới mà không làm ảnh hưởng đến các dữ liệu bản ghi cũ đang tồn tại.

-- Quay lại (Down Migration / Rollback): Hoàn toàn có thể. Mỗi file migration tiêu chuẩn luôn được thiết kế theo cặp (một file để nâng cấp và một file hoặc một hàm tương ứng để hạ cấp). Nếu quá trình triển khai gặp sự cố hoặc bạn muốn đưa database về trạng thái trước khi có cột avatar_url, bạn có thể chạy lệnh rollback (migrate down).
