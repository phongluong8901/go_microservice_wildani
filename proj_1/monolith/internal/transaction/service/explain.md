Khởi tạo Database Transaction & Lấy ví

tx, err := s.db.BeginTx(ctx, nil): Bắt đầu một Database Transaction để đảm bảo tính nguyên tử (Atomicity - thành công hết hoặc rollback hết).

defer tx.Rollback(): Đảm bảo nếu có lỗi xảy ra giữa chừng mà chưa kịp Commit, transaction sẽ tự động được hoàn tác (rollback) để tránh lệch dữ liệu.

Lấy thông tin ví của người gửi và người nhận dựa trên ID tương ứng. Kiểm tra điều kiện: người gửi và người nhận không được trùng nhau (senderWallet.ID == receiverWallet.ID).

Kiểm tra số dư: Nếu số dư ví người gửi nhỏ hơn số tiền muốn chuyển (senderWallet.Balance.LessThan(req.Amount)), trả về lỗi INSUFFICIENT_BALANCE.

Cơ chế Optimistic Locking (Khóa lạc quan)

Trừ tiền người gửi & Cộng tiền người nhận:
Sử dụng UpdateBalanceTx kèm theo số phiên bản (version) hiện tại của ví để cập nhật vào database.

Nếu trong lúc cập nhật mà version trong database đã bị thay đổi bởi một tiến trình khác chạy song song, hàm sẽ trả về lỗi xung đột (CONCURRENCY_CONFLICT / HTTP 409) để yêu cầu client thử lại sau.

Ghi nhận Giao dịch và Sổ cái (Ledger)

transactionID := uuid.New().String(): Sinh mã giao dịch mới.

s.txRepo.CreateTx(...): Lưu thông tin chung của giao dịch vào bảng transactions với trạng thái "success".

Tạo bút toán sổ cái kép (Double-entry bookkeeping):

Tạo bản ghi debit (ghi nợ) trừ tiền vào ví người gửi.

Tạo bản ghi credit (ghi có) cộng tiền vào ví người nhận.

Cả hai bút toán này đều được ghi vào bảng ledger thông qua s.ledgerRepo.CreateTx.

Hoàn tất Transaction

if err := tx.Commit(); err != nil: Nếu tất cả các bước trên thành công, tiến hành Commit để lưu chính thức xuống cơ sở dữ liệu.

Trả về thông tin transaction vừa hoàn thành.