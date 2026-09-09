# --- tac dung khi gen ra 2 file _grpc.pb.go va .pb.go
Mục đích chính của 2 file được sinh ra không phải để tự tạo kết nối mạng, mà là để chuyển đổi file hợp đồng .proto thành mã nguồn Go chuẩn, cung cấp các cấu trúc dữ liệu và đoạn mã trung gian giúp các microservices gọi gRPC với nhau một cách an toàn và dễ dàng.

Cụ thể, vai trò của chúng trong mô hình gRPC gồm:

1. Biên dịch dữ liệu thành code Go (ledger.pb.go)

Thay vì viết thủ công các struct trong Go cho từng message (như RecordEntryRequest, LedgerEntry), trình biên dịch tự động sinh ra chúng.

Cung cấp các phương thức nén/giải nén dữ liệu sang định dạng nhị phân (Binary) cực nhanh của Protobuf, và các hàm tiện ích như GetAmount(), GetWalletId().

2. Tạo khung kết nối và gọi hàm (ledger_grpc.pb.go)

Phía Client: Sinh ra sẵn các hàm gọi mạng (Client Stub). Khi service khác muốn gọi Ledger Service, họ chỉ cần gọi hàm Go thông thường (ví dụ: client.RecordLedgerEntry(...)), phần code này sẽ tự động đóng gói dữ liệu thành gói tin nhị phân và gửi qua mạng.

Phía Server: Định nghĩa sẵn các interface (như LedgerServiceServer) để bắt buộc lập trình viên phải viết code khớp đúng với những gì đã định nghĩa trong file .proto.

Tóm lại quy trình:
Bạn viết file .proto (bản thiết kế chung) -> Dùng protoc sinh ra 2 file .pb.go (bản dịch sang ngôn ngữ Go) -> Bạn viết code Go (như file grpc server ở trên) để implement các interface đó và thực hiện kết nối gRPC thực tế.

# ---
Hai file ledger.pb.go và ledger_grpc.pb.go là mã nguồn Go được sinh ra tự động từ file .proto thông qua công cụ biên dịch protoc bằng các plugin của Go (protoc-gen-go và protoc-gen-go-grpc).

ledger.pb.go (Xử lý Data Structures)

Định nghĩa Struct: Chuyển đổi toàn bộ các khối message trong file .proto thành các cấu trúc struct trong Go (ví dụ: RecordEntryRequest, LedgerEntry, BalanceResponse,...).

Serialization / Deserialization: Cung cấp các phương thức để mã hóa (marshal) dữ liệu từ struct Go sang định dạng nhị phân Protocol Buffers và ngược lại (unmarshal) khi truyền qua mạng.

Tiện ích đi kèm: Tự động sinh ra các hàm Getter (như req.GetAmount(), req.GetWalletId()) và các hàm hỗ trợ thao tác với JSON.

ledger_grpc.pb.go (Xử lý gRPC Networking)

Service Interfaces: Định nghĩa các interface cho phía Server và Client dựa trên khối service trong file .proto.

LedgerServiceServer: Interface mà server phải implement (ví dụ như struct ledgerGRPCServer trong code của bạn chứa các hàm RecordLedgerEntry, GetBalanceFromLedger,...).

LedgerServiceClient: Interface và client stub để các service khác gọi đến Ledger Service qua mạng.

Server Registration: Cung cấp hàm RegisterLedgerServiceServer để gắn gRPC server vào grpc.Server chính của ứng dụng Go.

# --- 1,2,3,4
Các con số = 1, = 2, = 3,... trong Protocol Buffers được gọi là Field Number (Số nhận diện trường), và chúng đóng vai trò cốt lõi vì những lý do sau:

Tối ưu hóa băng thông (Binary Serialization): Khi truyền dữ liệu qua mạng, Protobuf không gửi tên trường (ví dụ: không gửi chữ "transaction_id" hay "wallet_id"). Thay vào đó, nó chỉ mã hóa các con số định danh này kèm theo giá trị. Điều này giúp gói tin siêu nhỏ, tiết kiệm băng thông và tăng tốc độ truyền tải mạng đáng kể so với JSON hay XML.

Đảm bảo tính tương thích ngược (Backward/Forward Compatibility): Các microservices trong hệ thống có thể được nâng cấp vào những thời điểm khác nhau. Nếu sau này bạn thay đổi tên biến hoặc thêm trường mới, miễn là field number được giữ nguyên, các service cũ và mới vẫn đọc và hiểu đúng dữ liệu của nhau mà không bị xung đột.

Độc lập về thứ tự: Khi sử dụng số định danh, bạn có thể thoải mái sắp xếp lại vị trí các dòng trong message ở các phiên bản file .proto sau mà không làm gián đoạn việc giải mã dữ liệu nhị phân.

Quy tắc quan trọng: Một khi các service đã chạy trên môi trường thực tế (Production), tuyệt đối không được thay đổi số thứ tự của các trường đã tồn tại, vì điều đó sẽ làm sai lệch hoàn toàn cách các service đọc và ghi dữ liệu của nhau.