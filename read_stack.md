# --- lib


# --- stack


# --- more
1. Ledger system

Ledger System (Hệ thống sổ cái) là một mô hình kiến trúc dữ liệu ghi lại toàn bộ mọi giao dịch tài chính hoặc thay đổi trạng thái dưới dạng các bút toán (entries) bất biến (immutable) và không bao giờ được phép sửa hay xóa dữ liệu cũ.

Double-entry bookkeeping (Kế toán kép): Mỗi giao dịch tài chính phải được ghi nhận tối thiểu qua hai bút toán: Debit (Nợ) tài khoản này và Credit (Có) tài khoản kia. Tổng số dư của toàn hệ thống luôn phải cân bằng.

Append-only (Chỉ ghi nối tiếp): Dữ liệu trong ledger không bao giờ bị UPDATE hay DELETE. Nếu có sai sót hoặc cần hoàn tiền (refund), hệ thống sẽ tạo một giao dịch nghịch đảo (compensating transaction) ghi đè lên chứ không sửa bản ghi gốc. Điều này đảm bảo tính minh bạch và vết kiểm toán (audit trail) hoàn hảo.

immutability
create leadger entries, the money transaction records
balance an wallets, how about when got hacked? how to know this user has 10 million money, wher is this comming from
every money  come in called credit
come out it would be debit, we can't have update/delete query in this ledger_entries

2. optimistic locking in databse, pessimistic locking
Optimistic locking (khóa lạc quan) và pessimistic locking (khóa bi quan) là hai chiến lược kiểm soát đồng thời (concurrency control) trong cơ sở dữ liệu để giải quyết tranh chấp khi nhiều giao dịch cùng truy cập và cập nhật một bản ghi dữ liệu.

Pessimistic Locking (Khóa bi quan)
Nguyên lý: Giả định rằng xung đột dữ liệu chắc chắn sẽ xảy ra. Khi một transaction đọc một bản ghi để sửa, nó sẽ khóa (lock) bản ghi đó lại ngay lập tức trên database để các transaction khác không thể đọc hoặc ghi cho đến khi transaction hiện tại hoàn tất.

Optimistic Locking (Khóa lạc quan)
Nguyên lý: Giả định rằng xung đột rất ít khi xảy ra. Không có khóa nào được đặt trên database trong lúc đọc dữ liệu. Thay vào đó, mỗi bản ghi sẽ có một trường phiên bản (thường gọi là version kiểu số nguyên hoặc timestamp). Khi cập nhật, hệ thống kiểm tra xem version trong DB có khớp với lúc đọc ban đầu không. Nếu khớp thì cho update và tăng version lên; nếu không khớp (có người khác đã sửa trước), transaction sẽ bị hủy hoặc phải retry.

Idempotency Key (Khóa đồng nhất) là một cơ chế thiết kế API giúp đảm bảo rằng dù một yêu cầu (request) được client gửi đi gửi lại nhiều lần do lỗi mạng, timeout hoặc retry tự động, hệ thống phía server chỉ thực hiện hành động đó đúng một lần duy nhất.

Database Locking (Optimistic/Pessimistic): Giải quyết tranh chấp khi nhiều user khác nhau cùng cố gắng cập nhật một bản ghi dữ liệu tại cùng một thời điểm.

Idempotency Key: Giải quyết vấn đề một user gửi lặp lại chính request của họ do mất kết nối Internet hoặc app tự động retry.

race condition, , double transaction/spending
inwallets table we add 'version' column
make transaction process is save and with databse transaction sql.Tx
two scenario that occured in the mili-second time
1.user A has balance 100
2.user A transfer to user B about 70
3.in the same time, user A withdraw 50
120 transaction, 120 > 100
if server let them go pararelly without concurrency control
-thread 1 reading balance user A 100, reduce it with 70 so equal 30
-thread 2 also reading balence A 100, (before thread 1 done wirting)
    reduce 50 so equal 50
-the real transaction should be balance 100-70-50 = -20 (rejected)
-balance can be 30 or 50 if this not prevented
there have 2 solution to solve this race condition"

pessimistic locking->locking table while process, less concurrent
lock the database, until tracsaction done, it would be save but so slow
blocking other thread, queueing

optimistic locking ->optimistic locking is that we use version column in database table
faster because no blocking other thread, onlt check version before update

if the other update the data, version in db should be incremented, our reading version 
not equal to db version so the update should be failed
user should re-do transaction

3. soft delete
delete from the db, but with timestamp in the delete_at column guarantee referential integrity
we dont want to delete the important data, transfer history, ledger if we delete user
literally, we will face error for the foreign key

DELETE FROM users;
is deleted = true/false, deleted at nullable, if has time then it should be falg as deleted

pagnition & sorting
milion data, so slow and so long for the query and it affect api response

file upload
i just want to share how to save the avater/profile picture of the user

4. unit test
ensure the quality code
mocking, is act like a database