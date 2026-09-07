# --- lib
"net/http/httputil" // Thư viện chuẩn của Go hỗ trợ xây dựng Reverse Proxy


# --- stack
1. go.work
go.work (Go Workspace) là tính năng được Go giới thiệu từ phiên bản 1.18 để quản lý đồng thời nhiều module độc lập trong cùng một dự án lớn (như kiến trúc Microservices).

Phát triển cục bộ (Local Development) dễ dàng: Khi bạn sửa code ở một module dùng chung (ví dụ: thư viện common hoặc pkg), các microservice khác đang dùng thư viện đó sẽ nhận ngay thay đổi mã nguồn ngay lập tức mà không cần phải go get hay đẩy lên Git rồi tải lại.

hay thế hoàn toàn lệnh replace trong go.mod: Tránh việc phải cấu hình replace ../pkg thủ công ở từng go.mod của mỗi service (vốn rất dễ bị nhầm lẫn và gây lỗi khi build trên môi trường Production).

Độc lập Dependency: Mỗi microservice vẫn giữ file go.mod riêng của nó để quản lý các thư viện bên thứ ba (như Gin, Gorm, JWT...), nhưng file go.work ở thư mục gốc sẽ đóng vai trò điều phối chung toàn bộ workspace.

2. reverse proxy (proxy/reverse_proxy.go)
Tác dụng chính của file reverse proxy (proxy/reverse_proxy.go) trong kiến trúc microservice là đóng vai trò như một cổng trung gian thông minh nhận toàn bộ request từ client (hoặc API Gateway) rồi âm thầm chuyển tiếp (forward) chúng đến các microservice backend phù hợp, cụ thể:

Điều hướng và gom kênh (Routing): Giúp API Gateway không cần viết code xử lý nghiệp vụ thủ công cho từng đường dẫn, mà tự động nhận diện và ném request qua đúng service đích (ví dụ: /api/v1/auth/* sang Auth Service, /api/v1/wallets/* sang Wallet Service).

Chuẩn hóa Request & Header (proxy.Rewrite):

Ghi đè lại URL đích chuẩn xác thông qua r.SetURL(url).

Lưu lại địa chỉ máy khách gốc vào header X-Forwarded-Host để service phía sau biết nguồn gốc request.

Theo dõi vết phân tán (X-Correlation-ID): Tự động kiểm tra xem client đã gửi kèm mã X-Correlation-ID chưa. Nếu chưa (request mới tinh từ client ngoài vào), nó tự sinh một mã UUID mới. Mã này được đính kèm vào header để chuyền xuyên suốt qua tất cả các microservice phía sau, giúp lập trình viên dễ dàng tra cứu log khi hệ thống bị lỗi.

Che giấu cấu trúc nội bộ: Các microservice con chạy ở các cổng nội bộ ẩn (như 8081, 8082, 8086) hoàn toàn bị che khuất với bên ngoài. Client bên ngoài chỉ giao tiếp duy nhất với API Gateway thông qua reverse proxy, giúp tăng tính bảo mật cho hệ thống mạng.

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

5. swagger - API documentation

6. redis
check in redis? if exist them return from the redis
if not then query from mysql, save to redis, return

always initiate on write, when balance increase/decresae (after transfer)
we must delete the cache key in redis, so the next balance will from mysql

7. rate limmiting and JWT blacklisting
secure API from brute forece/absue attack, redfis based rate limmiting
create logout feature, to blackist jwt in redis

API max 60 req/ minutes 429 too manu reqeust if exceeded
save the key rate_limit ip_address:minutes

jwt, token is blacklisted, 410 gone

8. security
bad approach
response like emaul not found

good approach
respponse such as if the email is registered, reset code has been sent

9. tac dung go routine trong gui mail otp
Tối ưu hiệu năng (Non-blocking response): Khách hàng không phải chờ đợi quá trình kết nối và truyền dữ liệu qua giao thức SMTP (vốn có độ trễ cao từ vài trăm mili-giây đến vài giây). Giúp API phản hồi nhanh chóng (giảm thiểu HTTP latency).

Cô lập vòng đời nhờ Context độc lập (bgCtx): Khi HTTP request của client kết thúc hoặc bị hủy, context gốc sẽ tự động bị hủy (cancel). Nếu dùng trực tiếp context gốc, tiến trình gửi email đang chạy ngầm sẽ bị ngắt đột ngột giữa chừng. Việc khởi tạo context.WithTimeout(context.Background(), 10*time.Second) giúp tiến trình nền có một không gian thời gian độc lập để hoàn thành việc truyền tải dữ liệu.

Phòng chống rò rỉ tài nguyên và treo tiến trình (Goroutine Leak / Hanging): Giới hạn thời gian cứng 10 giây ngăn chặn tình trạng goroutine bị kẹt vĩnh viễn trong trạng thái chờ (waiting) nếu máy chủ SMTP gặp sự cố mất kết nối hoặc phản hồi cực kỳ chậm.

Bảo vệ luồng nghiệp vụ chính (Fault Isolation): Sự cố phát sinh từ dịch vụ bên thứ ba hoặc máy chủ email (như lỗi mạng SMTP) được cô lập hoàn toàn, không làm gián đoạn hay trả về lỗi thất bại cho các nghiệp vụ cốt lõi quan trọng của người dùng (như đăng ký tài khoản hay yêu cầu đổi mật khẩu).

10. concept Refresh Token Rotation (RTR) & resuse Detection
if RT stolen by hacker, hacker can generate new refresh token  forever
RTR, RT-A used to generate new refresh token RT-B, we need to revoke the RT-A
Reuse detection, RT-A that  has been revoked, if used again, system should detectthat this RT-A was stolen, delete all refresh token from the user so the real user, and hacker automatically logged out


Refresh Token Rotation (RTR) và Reuse Detection là cơ chế bảo mật nâng cao nhằm bảo vệ phiên đăng nhập của người dùng khỏi việc bị kẻ gian đánh cắp và lợi dụng (Token Hijacking).

1. Refresh Token Rotation (Xoay vòng Refresh Token)
Trong mô hình xác thực thông thường, mỗi khi Refresh Token được dùng để cấp lại Access Token mới, hệ thống sẽ trả về cùng một Refresh Token cũ cho đến khi nó hết hạn (thường kéo dài nhiều ngày hoặc vài tuần).

RTR hoạt động ngược lại: Mỗi khi client dùng một Refresh Token cũ để xin cấp Access Token mới, server sẽ thu hồi (revoke) token cũ đó ngay lập tức và phát hành một Refresh Token hoàn toàn mới.

Lợi ích: Giảm thiểu khoảng thời gian tồn tại của một Refresh Token. Kẻ tấn công nếu có lỡ lấy được một Refresh Token cũ, chúng chỉ dùng được một lần duy nhất.

2. Reuse Detection (Phát hiện tái sử dụng token độc hại)
Khi áp dụng RTR, server cần lưu trạng thái hoặc chuỗi liên kết (family) của các token.

Cách hoạt động: Server ghi nhận token nào đã bị sử dụng rồi. Nếu một ngày nào đó, một Refresh Token đã bị đánh dấu là "đã dùng/đã bị thay thế" mà lại xuất hiện thêm một yêu cầu đổi token lần nữa (thường xảy ra do kẻ tấn công đang cố dùng token cũ đã đánh cắp, trong khi người dùng thực sự vẫn đang dùng token mới), hệ thống sẽ lập tức nhận diện đây là hành vi tấn công chiếm đoạt phiên (Token Reuse Attack).

Hành động phản ứng: Khi phát hiện tái sử dụng, hệ thống sẽ lập tức vô hiệu hóa toàn bộ chuỗi token đó, đồng thời thu hồi toàn bộ phiên đăng nhập của user trên mọi thiết bị (xóa sạch các key tương ứng trên Redis) và buộc người dùng phải đăng nhập lại từ đầu.

11. Role Based Access Control (RBAC)
admin -> can see all the user wallet and balance, freeze, log system
user -> only see their wallet and balance

Role-Based Access Control (RBAC) trong Go là mô hình phân quyền truy cập hệ thống dựa trên vai trò của người dùng (ví dụ: admin, user, merchant). Thay vì gán quyền trực tiếp cho từng cá nhân, ứng dụng sẽ gán quyền cho các role, sau đó phân role cho người dùng.

các thành phần cốt lõi của RBAC
User: Người dùng thực hiện request vào hệ thống.

Role: Nhóm chức vụ hoặc tập hợp các quyền (Ví dụ: admin toàn quyền, user chỉ được xem và quản lý ví cá nhân).

Permission: Quyền hạn chi tiết trên từng tài nguyên (Ví dụ: wallet:read, wallet:transfer).
