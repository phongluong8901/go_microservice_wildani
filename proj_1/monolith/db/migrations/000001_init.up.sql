CREATE TABLE users (
    id VARCHAR(36) PRIMARY KEY,
    full_name VARCHAR(100) NOT NULL,
    email VARCHAR(150) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE TABLE wallets (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) UNIQUE NOT NULL,
    balance DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
    currency VARCHAR(3) NOT NULL DEFAULT 'IDR',
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active, frozen
    version INT NOT NULL DEFAULT 1, -- used for optimistic locking
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE transactions (
    id VARCHAR(36) PRIMARY KEY,
    sender_wallet_id VARCHAR(36) NULL,
    receiver_wallet_id VARCHAR(36) NOT NULL,
    amount DECIMAL(15, 2) NOT NULL,
    description TEXT NULL,
    idempotency_key VARCHAR(100) UNIQUE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'success', -- success, pending, failed
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sender_wallet_id) REFERENCES wallets(id),
    FOREIGN KEY (receiver_wallet_id) REFERENCES wallets(id)
);

CREATE TABLE ledger_entries (
    id VARCHAR(36) PRIMARY KEY,
    wallet_id VARCHAR(36) NOT NULL,
    transaction_id VARCHAR(36) NOT NULL,
    entry_type VARCHAR(10) NOT NULL, -- credit (+), debit (-)
    amount DECIMAL(15, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (wallet_id) REFERENCES wallets(id),
    FOREIGN KEY (transaction_id) REFERENCES transactions(id)
);

-- expalain
-- 1. Bảng users (Quản lý thông tin người dùng)
-- CREATE TABLE users: Khởi tạo bảng lưu trữ thông tin tài khoản người dùng hệ thống.

-- id VARCHAR(36) PRIMARY KEY: Khóa chính định danh người dùng. Dùng chuỗi 36 ký tự để lưu chuẩn UUID v4 (ví dụ: 123e4567-e89b-12d3-a456-426614174000).

-- full_name VARCHAR(100) NOT NULL: Họ và tên đầy đủ của người dùng, tối đa 100 ký tự và bắt buộc phải có (NOT NULL).

-- email VARCHAR(150) UNIQUE NOT NULL: Địa chỉ email đăng nhập. Phải là duy nhất (UNIQUE, không trùng lặp giữa các user) và bắt buộc nhập.

-- password_hash VARCHAR(255) NOT NULL: Lưu chuỗi mật khẩu đã được mã hóa băm (ví dụ dùng bcrypt hoặc argon2), độ dài 255 ký tự để đủ chứa chuỗi mã hóa dài.

-- created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP: Thời điểm bản ghi được tạo, mặc định tự động lấy thời gian hiện tại.

-- updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP: Thời điểm cập nhật thông tin gần nhất; MySQL sẽ tự động đổi thời gian này thành hiện tại mỗi khi dòng dữ liệu bị thay đổi.

-- deleted_at TIMESTAMP NULL: Dùng cho kỹ thuật Soft Delete (xóa mềm). Nếu cột này có giá trị (không NULL), hiểu là tài khoản đã bị khóa/xóa, tránh việc xóa vĩnh viễn dữ liệu khỏi database.

-- 2. Bảng wallets (Quản lý ví tiền của người dùng)
-- CREATE TABLE wallets: Khởi tạo bảng chứa số dư tài khoản gắn với từng user.

-- id VARCHAR(36) PRIMARY KEY: Khóa chính của ví (UUID v4, 36 ký tự).

-- user_id VARCHAR(36) UNIQUE NOT NULL: Liên kết ví với bảng users. Ràng buộc UNIQUE đảm bảo mỗi user chỉ có tối đa 1 ví chính trong hệ thống (tùy nghiệp vụ).

-- balance DECIMAL(15, 2) NOT NULL DEFAULT 0.00: Số dư ví. Dùng kiểu DECIMAL (15 chữ số tổng cộng, trong đó có 2 chữ số phần thập phân) để tránh sai số tính toán tiền tệ so với kiểu FLOAT/DOUBLE. Mặc định ban đầu là 0.00.

-- currency VARCHAR(3) NOT NULL DEFAULT 'IDR': Đơn vị tiền tệ (ví dụ: IDR - Rupiah Indonesia, USD, VND), mặc định là IDR.

-- status VARCHAR(20) NOT NULL DEFAULT 'active': Trạng thái ví (active: hoạt động, frozen: bị đóng băng/khóa).

-- version INT NOT NULL DEFAULT 1: Biến đếm phiên bản dùng cho Optimistic Locking (khóa lạc quan) nhằm chống tranh chấp dữ liệu (race condition) khi có nhiều luồng cùng rút/nạp tiền vào một ví cùng lúc.

-- created_at / updated_at / deleted_at: Tương tự bảng users (thời điểm tạo, sửa, xóa mềm).

-- FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE: Khóa ngoại liên kết với bảng users. ON DELETE CASCADE nghĩa là nếu tài khoản user bị xóa vĩnh viễn, ví tiền gắn liền với user đó cũng sẽ tự động bị xóa theo.

-- 3. Bảng transactions (Lịch sử giao dịch chuyển tiền)
-- CREATE TABLE transactions: Lưu lịch sử mỗi lần tiền di chuyển từ ví này sang ví khác.

-- id VARCHAR(36) PRIMARY KEY: Mã giao dịch độc nhất (UUID v4).

-- sender_wallet_id VARCHAR(36) NULL: ID ví của người gửi tiền. Cho phép NULL (vì có trường hợp tiền được nạp từ hệ thống bên ngoài vào, không có người gửi cụ thể).

-- receiver_wallet_id VARCHAR(36) NOT NULL: ID ví của người nhận tiền (bắt buộc phải có người nhận).

-- amount DECIMAL(15, 2) NOT NULL: Số tiền giao dịch trong lần này.

-- description TEXT NULL: Ghi chú nội dung giao dịch (có thể bỏ trống).

-- idempotency_key VARCHAR(100) UNIQUE NOT NULL: Khóa chống trùng lặp giao dịch (Idempotency Key). Khi mạng lag và user bấm nút "Chuyển tiền" 2 lần liên tục, khóa này giúp hệ thống nhận diện và chặn không cho trừ tiền 2 lần.

-- status VARCHAR(20) NOT NULL DEFAULT 'success': Trạng thái giao dịch (success: thành công, pending: đang xử lý, failed: thất bại).

-- created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP: Thời điểm phát sinh giao dịch.

-- FOREIGN KEY ...: Khóa ngoại nối sender_wallet_id và receiver_wallet_id về bảng wallets.

-- 4. Bảng ledger_entries (Sổ cái kế toán chi tiết dòng tiền)
-- CREATE TABLE ledger_entries: Bảng ghi lại mọi biến động chi tiết (vào/ra) của từng ví đối với mỗi giao dịch nhằm đảm bảo tính toàn vẹn tài chính theo chuẩn kế toán kép (Double-entry bookkeeping).

-- id VARCHAR(36) PRIMARY KEY: Khóa chính của bản ghi sổ cái (UUID v4).

-- wallet_id VARCHAR(36) NOT NULL: Ví nào chịu ảnh hưởng bởi biến động này.

-- transaction_id VARCHAR(36) NOT NULL: Biến động này thuộc về mã giao dịch nào ở bảng transactions.

-- entry_type VARCHAR(10) NOT NULL: Loại biến động (credit: cộng tiền vào ví, debit: trừ tiền khỏi ví).

-- amount DECIMAL(15, 2) NOT NULL: Số tiền tác động trong bút toán này.

-- created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP: Thời điểm ghi sổ.

-- FOREIGN KEY ...: Các khóa ngoại liên kết ngược lại với bảng wallets và transactions.