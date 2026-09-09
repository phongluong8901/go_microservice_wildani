các biến port gRPC trong file .env (như LEDGER_GRPC_PORT=50054 hay LEDGER_GRPC_ADDR=ledger-service:50054) là các cổng mạng TCP dùng riêng cho giao tiếp nội bộ giữa các microservices với nhau, tách biệt hoàn toàn với cổng HTTP/REST thông thường.

Cấu hình Port lắng nghe cho Server (LEDGER_GRPC_PORT): Khi khởi động một microservice (ví dụ Ledger Service), ứng dụng sẽ dùng biến này để mở cổng mạng (lắng nghe ở port 50054), sẵn sàng nhận các gói tin gRPC từ các service khác gửi đến.

Cấu hình Địa chỉ kết nối cho Client (LEDGER_GRPC_ADDR): Khi một service khác (như Service gọi sang Ledger) muốn gọi hàm gRPC, nó sẽ nhìn vào biến này để biết chính xác tên container/host và số port cần kết nối (ví dụ: ledger-service:50054).

Phân tách loại traffic: Giúp hệ thống phân định rõ ràng:

Cổng HTTP (Ví dụ: LEDGER_PORT=8085): Dành cho API Gateway hoặc HTTP client gọi vào.

Cổng gRPC (Ví dụ: LEDGER_GRPC_PORT=50054): Dành riêng cho giao tiếp gRPC nội bộ (sử dụng giao thức HTTP/2 và mã hóa nhị phân Protobuf với hiệu suất cực cao và tốc độ nhanh hơn REST/JSON).