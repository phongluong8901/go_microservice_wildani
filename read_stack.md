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

3. gRPC
gRPC đóng vai trò là giao thức giao tiếp liên dịch vụ (Inter-service Communication) cực kỳ nhanh chóng và hiệu quả giữa các microservices (ví dụ: giữa API Gateway và các service phía sau như Auth Service, User Service) trong dự án này.

Giao tiếp nội bộ tốc độ cao: Thay vì dùng HTTP/JSON truyền thống (REST API) vốn nặng nề và tốn tài nguyên khi gọi qua lại giữa các service, gRPC sử dụng HTTP/2 và định dạng nhị phân Protocol Buffers (Protobuf) giúp nén dữ liệu nhỏ hơn nhiều lần và truyền tải nhanh hơn.

Định nghĩa hợp đồng rõ ràng (Contract-First): Thông qua các file .proto, dự án định nghĩa sẵn cấu trúc dữ liệu và các hàm (RPC methods) mà các service cung cấp. Từ đó, mã nguồn client và server cho các microservices được sinh ra tự động, đảm bảo tính đồng nhất và an toàn kiểu dữ liệu (type-safe).

Hỗ trợ Streaming: gRPC hỗ trợ các luồng dữ liệu thời gian thực (Client/Server/Bi-directional Streaming), rất hữu ích cho các tính năng ví điện tử cần thông báo số dư hoặc giao dịch biến động theo thời gian thực.

4. Saga Pattern
Saga Pattern là một mẫu thiết kế kiến trúc phần mềm dùng để quản lý giao dịch phân tán (distributed transactions) trong các hệ thống microservices.

Trong hệ thống microservices của bạn, mỗi service (như wallet-service, ledger-service, user-service) có cơ sở dữ liệu riêng biệt. Khi một nghiệp vụ lớn xảy ra (ví dụ: chuyển tiền từ ví này sang ví khác), nó cần thao tác qua nhiều service. Bạn không thể dùng lệnh COMMIT hoặc ROLLBACK truyền thống của cơ sở dữ liệu trên nhiều database độc lập như vậy được. Saga sinh ra để giải quyết bài toán này.

Cách hoạt động của Saga:
Saga chia một giao dịch lớn thành một chuỗi các bước cục bộ (local transactions) nhỏ thực hiện tuần tự qua các service:

Bước 1: Service A thực hiện transaction của nó và phát ra một sự kiện (event) hoặc lời gọi tới Service B.

Bước 2: Service B nhận được, thực hiện transaction của nó.

Nếu tất cả thành công: Giao dịch hoàn tất.

Nếu một bước bị lỗi (Failure): Saga sẽ kích hoạt các giao dịch bù trừ (compensating transactions) để đi ngược lại các bước đã làm trước đó nhằm hoàn tác (rollback) dữ liệu về trạng thái ban đầu, đảm bảo tính nhất quán cuối cùng (eventual consistency).

Vai trò cụ thể của Saga trong mã nguồn của bạn:
Đảm bảo tính nhất quán tài chính: Trong ứng dụng ví điện tử (gowallet), việc sai sót lệch tiền giữa các tài khoản là tối kỵ. Saga đóng vai trò điều phối dòng tiền đi qua các bước (ví dụ: trừ tiền ví nguồn $\rightarrow$ cộng tiền ví đích $\rightarrow$ ghi nhận sổ cái ledger).

Xử lý lỗi phân tán (Failure Recovery): Nếu hệ thống ghi sổ cái (ledger-service) bị lỗi sau khi ví đã bị trừ tiền, Saga đảm bảo tiền trong ví sẽ được tự động hoàn trả thay vì bị mất tích giữa chừng.

Giải phóng ràng buộc Database: Cho phép các microservices hoạt động độc lập, dùng database riêng mà vẫn phối hợp chặt chẽ được với nhau trong các nghiệp vụ phức tạp.

5. Outbox Pattern
Transactional Outbox Pattern là một design pattern trong kiến trúc phần mềm (đặc biệt phổ biến trong Microservices) dùng để giải quyết bài toán: Làm thế nào để lưu dữ liệu vào Database chính VÀ gửi message/event đi (qua Kafka, RabbitMQ, v.v.) một cách đáng tin cậy, đảm bảo tính nhất quán (Atomicity)

Vấn đề gặp phải nếu không dùng Outbox Pattern:Khi hệ thống cần thực hiện 2 hành động cùng lúc:Lưu transaction vào Database của dịch vụ (ví dụ: trừ tiền ví, tạo đơn hàng).Phát một message/event lên Message Broker (ví dụ: thông báo WalletDeducted để các service khác xử lý).Nếu viết code tuần tự như sau:Go// Cách làm dễ bị lỗi
db.Save(walletTransaction) // 1. Lưu DB thành công
messageBroker.Publish("wallet.deducted", event) // 2. LỖI MẠNG / BROKER SẬP!
Rủi ro: Bước 1 thành công nhưng bước 2 thất bại (do sập mạng, lỗi broker), dữ liệu trong DB đã đổi nhưng hệ thống bên ngoài không nhận được event $\rightarrow$ Dữ liệu bị lệch (Inconsistency).Ngược lại, nếu gửi message trước rồi lưu DB sau thì khi lưu DB lỗi, message đã bị phát đi oan uổng.

Cách Outbox Pattern giải quyết:
Ghi vào bảng Outbox: Thay vì gọi trực tiếp sang Message Broker, ứng dụng gom việc lưu dữ liệu nghiệp vụ và lưu event cần gửi vào cùng một Database Transaction.

SQL


BEGIN TRANSACTION;
  -- 1. Lưu giao dịch ví
  INSERT INTO transactions (id, user_id, amount) VALUES (...);

  -- 2. Lưu event vào bảng outbox (chung 1 DB transaction)
  INSERT INTO outbox_messages (id, event_type, payload, status) VALUES (...);
COMMIT;
Đẩy event đi sau: Một tiến trình nền riêng biệt (gọi là Message Relay hoặc Outbox Processor) sẽ đọc các bản ghi chưa gửi trong bảng outbox_messages, tiến hành gửi lên Message Broker (Kafka/RabbitMQ), sau đó cập nhật trạng thái thành sent hoặc xóa đi.

Tác dụng của Outbox Pattern trong project gowallet
Đảm bảo tính nhất quán dữ liệu tài chính (Atomicity): Các giao dịch tiền tệ đòi hỏi độ chính xác tuyệt đối. Khi một ví tiền thực hiện chuyển khoản hoặc thanh toán, thay đổi số dư và sự kiện phát sinh (WalletBalanceUpdated, TransactionCreated) phải luôn đồng bộ. Outbox pattern ngăn chặn hoàn toàn tình trạng tiền đã bị trừ trong database của ví nhưng hệ thống thông báo/lịch sử giao dịch không nhận được event.

Chống mất mát tin nhắn (At-least-once delivery): Do các message được lưu sẵn xuống database dưới dạng bảng trung gian (outbox), nếu Message Broker gặp sự cố sập nguồn hoặc mất mạng tạm thời, tiến trình background worker của outbox sẽ retry (thử lại) liên tục cho đến khi message được gửi thành công, đảm bảo không có sự kiện giao dịch nào bị bỏ sót.

Tách biệt luồng xử lý (Decoupling): Giúp API endpoint xử lý ví phản hồi nhanh hơn, không bị nghẽn cổ chai hay phụ thuộc vào độ trễ của mạng kết nối tới Message Broker bên ngoài trong lúc người dùng đang thực hiện request.

6. RabbitMQ và Event Publishing
RabbitMQ: Là một Message Broker (phần mềm trung gian quản lý hàng đợi tin nhắn) cho phép các ứng dụng hoặc microservices trao đổi thông tin với nhau một cách bất đồng bộ (asynchronous) thông qua các hàng đợi (queues) và luồng trao đổi (exchanges).

Event Publishing (Phát sự kiện): Là hành động một service (Publisher) phát đi một thông báo dạng sự kiện (ví dụ: TransactionCreated, WalletUpdated) lên Message Broker ngay sau khi một hành động nghiệp vụ hoàn tất, thay vì phải gọi trực tiếp sang các service khác.

Vai trò của RabbitMQ & Event Publishing trong project gowallet
Tách biệt hệ thống (Decoupling): Giúp transaction-service hoặc wallet-service không bị phụ thuộc chặt chẽ (tightly coupled) vào các service phụ trợ như thông báo, lịch sử, hoặc xử lý hóa đơn. Service chỉ cần thực hiện xong nhiệm vụ của mình và bắn event lên RabbitMQ.

Xử lý bất đồng bộ và tăng hiệu năng: Các tác vụ nặng hoặc không cần phản hồi tức thì (như gửi email thông báo, ghi log kiểm toán phức tạp) được đẩy vào queue để các worker service xử lý ngầm, giúp API phản hồi nhanh hơn cho người dùng.

Đảm bảo tính nhất quán (Eventual Consistency): Khi một giao dịch tài chính xảy ra, event phát đi đảm bảo rằng các service liên quan (như cập nhật số dư, ghi sổ cái ledger) đều nhận được dữ liệu và tự đồng bộ trạng thái của mình theo.

Tích hợp với Outbox Pattern: Làm cầu nối để đẩy các message được lưu tạm trong bảng outbox của database lên hệ thống message queue một cách an toàn, giải quyết triệt để vấn đề mất dữ liệu khi mạng hoặc broker gặp sự cố gián đoạn.

Khả năng chịu lỗi (Fault Tolerance): Nếu một microservice tiêu thụ (consumer) bị sập nguồn, RabbitMQ sẽ giữ lại các message trong queue và tiếp tục gửi lại khi service đó hồi phục, ngăn ngừa việc thất thoát giao dịch.

7. Audit cùng Mongo
Audit cùng Mongo là việc sử dụng cơ sở dữ liệu NoSQL MongoDB để lưu trữ nhật ký kiểm toán (audit logs), vết lịch sử hoạt động, hoặc các bản ghi sự kiện của hệ thống.

Trong kiến trúc microservices như gowallet, cách tiếp cận này mang lại những đặc điểm sau:
Linh hoạt về cấu trúc (Schemaless): Dữ liệu kiểm toán hoặc payload của các sự kiện thường có định dạng thay đổi tùy theo loại hành động (ví dụ: log của giao dịch nạp tiền sẽ khác với log đổi mật khẩu). MongoDB cho phép lưu trữ trực tiếp dưới dạng JSON/BSON mà không cần định nghĩa schema cứng nhắc hay thực hiện lệnh ALTER TABLE.

Hiệu năng ghi cao (High Write Throughput): Các hệ thống tài chính phát sinh lượng log và vết sự kiện rất lớn mỗi giây. Việc ghi log sang MongoDB giúp giảm tải cho cơ sở dữ liệu quan hệ chính (như PostgreSQL hoặc MySQL đang chuyên xử lý số dư ví và lệnh chuyển tiền cốt lõi).

Tách biệt dữ liệu: Tách bạch rõ ràng giữa cơ sở dữ liệu nghiệp vụ giao dịch (Transactional DB) và hệ thống lưu trữ nhật ký phân tích/kiểm tra (Audit/Log DB).

Cách sử dụng trong project gowallet
Lưu vết sự kiện hệ thống: Ghi nhận lại các mốc thời gian, trạng thái thay đổi của ví, hoặc các luồng sự kiện đi qua microservices để phục vụ cho việc tra soát khi xảy ra lỗi.
Lịch sử hoạt động: Lưu trữ các hành động của người dùng hoặc các yêu cầu API quan trọng để phục vụ công tác bảo mật và kiểm tra (auditing).



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

12. Circuit Breaker (Mạch ngắt) và DLQ (Dead Letter Queue - Hàng đợi thư chết)
Circuit Breaker (Mạch ngắt) và DLQ (Dead Letter Queue - Hàng đợi thư chết) là hai mẫu thiết kế (design patterns) cực kỳ quan trọng trong kiến trúc microservices nhằm đảm bảo tính chịu lỗi (fault tolerance) và độ tin cậy của hệ thống.
---
Circuit Breaker trong Go

Khái niệm: Là một lớp bảo vệ nằm giữa các microservice (ví dụ khi Gateway hoặc Ledger gọi sang Wallet service). Nó hoạt động giống như cầu dao điện tự động ngắt mạch khi phát hiện lỗi hệ thống liên tiếp.

Cơ chế hoạt động (3 trạng thái):

Closed (Đóng): Hoạt động bình thường, các request được gọi đi qua.

Open (Mở): Khi số lượng lỗi vượt ngưỡng cho phép, Circuit Breaker ngắt kết nối ngay lập tức, trả về lỗi giả lập (hoặc fallback) mà không gọi sang service đang chết, giúp service đích có thời gian hồi phục và tránh sập dây chuyền (cascading failure).

Half-Open (Nửa mở): Sau một khoảng thời gian, nó cho phép một vài request thử nghiệm đi qua. Nếu thành công, chuyển về trạng thái Closed; nếu tiếp tục lỗi, quay lại trạng thái Open.

Thư viện phổ biến trong Go: sony/gobreaker hoặc afex/hystrix-go.

DLQ (Dead Letter Queue) trong Go

Khái niệm: Là một hàng đợi phụ (queue/topic phụ) trong hệ thống message broker (như RabbitMQ hoặc Kafka mà bạn đang dùng) dùng để lưu trữ các thông điệp (messages) không thể xử lý thành công sau một số lần thử lại (retry) nhất định.
---
Lý do cần DLQ:

Tránh hiện tượng poison message (thông điệp lỗi làm consumer bị lặp vô hạn hoặc crash liên tục).

Giúp giữ lại dữ liệu lỗi để lập trình viên phân tích, debug hoặc xử lý thủ công (replay) sau khi đã sửa lỗi code mà không làm gián đoạn luồng xử lý chính.

Cách áp dụng: Khi cấu hình RabbitMQ/Kafka consumer trong Go, bạn thiết lập chính sách x-dead-letter-exchange hoặc cơ chế retry. Nếu bản tin xử lý lỗi vượt quá max_retries, broker sẽ tự động chuyển bản tin đó sang hàng đợi DLQ.

Trong hệ thống microservices tài chính ví điện tử (gowallet), việc tích hợp Circuit Breaker và DLQ (Dead Letter Queue) đóng vai trò sống còn để đảm bảo hệ thống không bị sập dây chuyền và không bị mất mát dữ liệu giao dịch.

Vai trò của Circuit Breaker trong hệ thống
---
Chống sập dây chuyền (Cascading Failures): Khi một service phụ trợ (ví dụ: Ledger Service hoặc Wallet Service) gặp sự cố hoặc quá tải, Circuit Breaker nằm ở các service gọi đến (như Transaction Service hoặc API Gateway) sẽ tự động ngắt kết nối, ngăn chặn việc tiếp tục bắn request dồn dập gây cạn kiệt tài nguyên (thread/connection pool).

Cơ chế phản hồi nhanh (Fail Fast): Thay vì bắt client hoặc service gọi chờ timeout dài gây nghẽn luồng, mạch mở giúp trả về lỗi ngay lập tức để hệ thống xử lý fallback hoặc hiển thị thông báo gián đoạn tạm thời.

Tạo khoảng lặng hồi phục: Cho phép service đích (bị lỗi) có thời gian tự phục hồi ổn định mà không bị áp lực từ lượng request khổng lồ đổ vào liên tục.

Vai trò của DLQ (Dead Letter Queue) trong hệ thống

Xử lý sự cố tin nhắn (Poison Messages): Trong kiến trúc hướng sự kiện (Event-driven qua RabbitMQ), khi một bản tin (message) xử lý thất bại do lỗi định dạng hoặc lỗi logic nghiệp vụ, cơ chế retry sẽ thử lại. Nếu vượt ngưỡng, tin nhắn sẽ bị loại bỏ khỏi hàng đợi chính để tránh làm đơ (block) hệ thống tiêu thụ (consumer).

Bảo vệ dữ liệu tài chính: Thay vì làm mất bản tin (gây lệch số dư hoặc mất vết giao dịch), hệ thống đẩy bản tin lỗi đó vào Dead Letter Queue.

Hỗ trợ kiểm tra và khôi phục (Replay): Giúp đội ngũ kỹ thuật giữ lại toàn bộ các giao dịch/sự kiện thất bại trong DLQ để phân tích log, sửa lỗi code, sau đó bơm ngược (replay) các bản tin đó trở lại hệ thống xử lý mà không sợ thất lạc dữ liệu.


13. 

















