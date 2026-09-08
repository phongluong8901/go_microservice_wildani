Các bước cài đặt gRPC và công cụ biên dịch Protocol Buffers trên Windows:

Cài đặt qua Winget (Khuyên dùng)

Mở PowerShell hoặc Command Prompt và chạy lệnh sau để tự động cài đặt protoc:

PowerShell


winget install protobuf
Sau khi cài xong, hãy tắt và mở lại PowerShell để hệ thống cập nhật đường dẫn (PATH).

Cài đặt các Go Plugins cho protoc

Chạy các lệnh sau trong terminal để tải trình biên dịch dành riêng cho Go:

PowerShell


go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
Thêm thư mục bin của Go vào biến môi trường PATH trên Windows

Đảm bảo rằng thư mục chứa các lệnh Go vừa cài (%GOPATH%\bin hoặc thường là C:\Users\<Tên_User>\go\bin) đã nằm trong biến môi trường PATH của Windows để hệ thống gọi được lệnh protoc-gen-go. Nếu chưa có, bạn có thể thêm thủ công trong phần Environment Variables.

Kiểm tra lại kết quả

Mở cửa sổ terminal mới và gõ các lệnh sau để xác nhận:

PowerShell


protoc --version
protoc-gen-go --version
protoc-gen-go-grpc --version