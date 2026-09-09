Trong đoạn code bạn cung cấp, các hàm đang thực hiện gọi gRPC (thông qua walletClient) nằm ở tầng Service, cụ thể là:

GetMutationHistory: Gọi s.walletClient.GetWalletByUserID(...) để lấy thông tin ví của người dùng từ Wallet Service qua gRPC.

ReconcileWalletBalance: Gọi s.walletClient.GetWalletByUserID(...) để lấy số dư ví hiện tại từ Wallet Service qua gRPC nhằm đối soát với dữ liệu sổ cái.

Thông tin liên quan:

Client gRPC: Sử dụng pbWallet.WalletServiceClient được inject vào cấu trúc ledgerService thông qua hàm khởi tạo NewLedgerService.

Các tầng khác không gọi gRPC trực tiếp: Handler xử lý HTTP/Gin, còn Repository thao tác trực tiếp với cơ sở dữ liệu MySQL.