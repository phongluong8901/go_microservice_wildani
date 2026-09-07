
CREATE TABLE refresh_tokens (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    token VARCHAR(500) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- VARCHAR(500) NOT NULL UNIQUE: Cho phép tạo khóa UNIQUE trực tiếp trên cột token giúp tăng tốc độ tìm kiếm bản ghi (GetByToken) và đảm bảo không bị trùng lặp. (Lưu ý: Nếu bạn dùng JWT tự ký có độ dài vượt quá 500 ký tự, hãy cân nhắc tăng lên VARCHAR(1000) hoặc dùng chuỗi opaque token ngắn gọn khoảng 64-128 ký tự).

-- TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP: Tự động quản lý mốc thời gian tạo và cập nhật bản ghi mà không cần code tầng ứng dụng phải truyền thủ công.

-- FOREIGN KEY ... ON DELETE CASCADE: Đảm bảo khi một user bị xóa khỏi hệ thống, toàn bộ các refresh token liên quan cũng sẽ được tự động dọn dẹp sạch sẽ khỏi database, tránh dữ liệu rác.