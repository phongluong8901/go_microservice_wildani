package grpc // Khai báo package grpc chứa logic xử lý gRPC server

import (
	"context"

	"github.com/bashocode/gowallet/microservices/ledger-service/internal/ledger/model"
	"github.com/bashocode/gowallet/microservices/ledger-service/internal/ledger/repository"
	pb "github.com/bashocode/gowallet/microservices/ledger-service/proto/ledger"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ledgerGRPCServer triển khai gRPC interface của LedgerService
type ledgerGRPCServer struct {
	pb.UnimplementedLedgerServiceServer                             // Bắt buộc phải nhúng để tương thích ngược gRPC trong Go
	repo                                repository.LedgerRepository // Repository thao tác với cơ sở dữ liệu
}

// NewLedgerGRPCServer khởi tạo một instance mới của ledgerGRPCServer
func NewLedgerGRPCServer(repo repository.LedgerRepository) pb.LedgerServiceServer {
	return &ledgerGRPCServer{repo: repo}
}

// RecordLedgerEntry xử lý yêu cầu ghi 1 bút toán đơn lẻ
func (s *ledgerGRPCServer) RecordLedgerEntry(ctx context.Context, req *pb.RecordEntryRequest) (*pb.RecordEntryResponse, error) {
	// Chuyển đổi chuỗi số tiền sang kiểu Decimal để tránh sai số tính toán tài chính
	amount, err := decimal.NewFromString(req.GetAmount())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid amount: %v", err)
	}

	// Khởi tạo model dữ liệu để chuẩn bị lưu xuống database
	entry := &model.LedgerEntry{
		ID:            uuid.New().String(), // Tạo UUID ngẫu nhiên cho bản ghi sổ cái
		TransactionID: req.GetTransactionId(),
		WalletID:      req.GetWalletId(),
		EntryType:     req.GetType(),
		Amount:        amount,
	}

	// Gọi repository để lưu bản ghi vào DB
	if err := s.repo.Create(ctx, entry); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to record ledger entry: %v", err)
	}

	// Trả về kết quả thành công kèm ID bản ghi sổ cái
	return &pb.RecordEntryResponse{
		EntryId: entry.ID,
		Success: true,
	}, nil
}

// RecordLedgerEntries xử lý yêu cầu ghi hàng loạt bút toán cùng lúc
func (s *ledgerGRPCServer) RecordLedgerEntries(ctx context.Context, req *pb.RecordEntriesRequest) (*pb.RecordEntriesResponse, error) {
	entries := make([]*model.LedgerEntry, 0, len(req.GetEntries()))
	for _, e := range req.GetEntries() {
		amount, err := decimal.NewFromString(e.GetAmount())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid amount: %v", err)
		}
		entries = append(entries, &model.LedgerEntry{
			ID:            uuid.New().String(),
			TransactionID: e.GetTransactionId(),
			WalletID:      e.GetWalletId(),
			EntryType:     e.GetType(),
			Amount:        amount,
		})
	}

	// Gọi repository thực hiện ghi hàng loạt (batch insert) tối ưu hiệu năng
	if err := s.repo.CreateBatch(ctx, entries); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to record ledger entries: %v", err)
	}

	return &pb.RecordEntriesResponse{Success: true}, nil
}

// GetBalanceFromLedger truy vấn và tính toán số dư hiện tại của ví
func (s *ledgerGRPCServer) GetBalanceFromLedger(ctx context.Context, req *pb.GetBalanceRequest) (*pb.BalanceResponse, error) {
	balance, err := s.repo.GetBalanceByWalletID(ctx, req.GetWalletId())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "failed to calculate balance: %v", err)
	}

	// Chuyển kiểu Decimal về chuỗi để trả về qua gRPC
	return &pb.BalanceResponse{
		CalculatedBalance: balance.String(),
	}, nil
}

// GetEntriesByWalletID lấy toàn bộ danh sách lịch sử bút toán của một ví
func (s *ledgerGRPCServer) GetEntriesByWalletID(ctx context.Context, req *pb.GetEntriesRequest) (*pb.EntriesResponse, error) {
	entries, err := s.repo.GetEntriesByWalletID(ctx, req.GetWalletId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get ledger entries: %v", err)
	}

	// Map danh sách model trong DB sang cấu trúc gRPC message tương ứng
	pbEntries := make([]*pb.LedgerEntry, 0, len(entries))
	for _, e := range entries {
		pbEntries = append(pbEntries, &pb.LedgerEntry{
			Id:            e.ID,
			TransactionId: e.TransactionID,
			WalletId:      e.WalletID,
			EntryType:     e.EntryType,
			Amount:        e.Amount.String(),
			CreatedAt:     e.CreatedAt.Format("2006-01-02T15:04:05Z07:00"), // Format thời gian chuẩn ISO 8601
		})
	}

	return &pb.EntriesResponse{Entries: pbEntries}, nil
}
