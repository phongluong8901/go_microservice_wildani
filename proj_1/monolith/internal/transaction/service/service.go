package service

import (
	"context"
	"database/sql"
	"net/http"

	customErr "github.com/bashocode/gowallet/monolith/internal/errors"
	ledgerModel "github.com/bashocode/gowallet/monolith/internal/ledger/model"
	ledgerRepo "github.com/bashocode/gowallet/monolith/internal/ledger/repository"
	"github.com/bashocode/gowallet/monolith/internal/transaction/model"
	"github.com/bashocode/gowallet/monolith/internal/transaction/repository"
	userRepo "github.com/bashocode/gowallet/monolith/internal/user/repository"
	walletRepo "github.com/bashocode/gowallet/monolith/internal/wallet/repository"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type TransactionService interface {
	Transfer(ctx context.Context, senderUserID string, req model.TransferRequest) (*model.Transaction, error)
	GetHistory(ctx context.Context, userID string, params model.PaginationParams) ([]model.Transaction, *model.PaginationMeta, error)
}

type transactionService struct {
	db         *sql.DB
	rdb        *redis.Client
	txRepo     repository.TransactionRepository
	userRepo   userRepo.UserRepository
	walletRepo walletRepo.WalletRepository
	ledgerRepo ledgerRepo.LedgerRepository
}

func NewTransactionService(
	db *sql.DB,
	rdb *redis.Client,
	txRepo repository.TransactionRepository,
	uRepo userRepo.UserRepository,
	wRepo walletRepo.WalletRepository,
	lRepo ledgerRepo.LedgerRepository,
) TransactionService {
	return &transactionService{
		db:         db,
		rdb:        rdb,
		txRepo:     txRepo,
		userRepo:   uRepo,
		walletRepo: wRepo,
		ledgerRepo: lRepo,
	}
}

// Transfer xử lý logic chuyển tiền giữa hai ví, áp dụng Transaction, khóa lạc quan (Optimistic Locking) và Idempotency Key.
func (s *transactionService) Transfer(ctx context.Context, senderUserID string, req model.TransferRequest) (*model.Transaction, error) {
	// check idempotency key (this is for checking to not reprocess the same request)
	// Kiểm tra IdempotencyKey để chống việc client gửi lặp lại một request giao dịch nhiều lần (double-spending/duplicate request).
	existing, _ := s.txRepo.GetByIdempotencyKey(ctx, req.IdempotencyKey)
	if existing != nil {
		return existing, nil
	}

	// look receiver by email // Tìm kiếm thông tin người nhận dựa vào email.
	receiverUser, err := s.userRepo.GetByEmail(ctx, req.ReceiverEmail)
	if err != nil {
		return nil, customErr.NewAppError(http.StatusNotFound, "RECEIVER_NOT_FOUND", "Receiver not found")
	}

	// start db transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, customErr.ErrInternalServer
	}
	// Đảm bảo gọi Rollback() khi hàm kết thúc (nếu không có Commit(), transaction sẽ tự hủy để đảm bảo tính toàn vẹn dữ liệu).
	defer tx.Rollback()

	// look for sender and receiver wallet // Lấy thông tin ví của người gửi (sender).
	senderWallet, err := s.walletRepo.GetByUserID(ctx, senderUserID)
	if err != nil {
		return nil, customErr.NewAppError(http.StatusNotFound, "SENDER_WALLET_NOT_FOUND", "Sender wallet not found")
	}

	// Lấy thông tin ví của người nhận (receiver).
	receiverWallet, err := s.walletRepo.GetByUserID(ctx, receiverUser.ID)
	if err != nil {
		return nil, customErr.NewAppError(http.StatusNotFound, "RECEIVER_WALLET_NOT_FOUND", "Receiver wallet not found")
	}

	// sender & receiver cannot be the same
	// Kiểm tra logic: Không cho phép tự chuyển tiền cho chính ví của mình.
	if senderWallet.ID == receiverWallet.ID {
		return nil, customErr.NewAppError(http.StatusBadRequest, "INVALID_TRANSFER", "Cannot transfer to self account")
	}

	// validate sender wallet balance is enough or not
	// Kiểm tra xem số dư ví người gửi có đủ tiền để chuyển hay không.
	if senderWallet.Balance.LessThan(req.Amount) {
		return nil, customErr.NewAppError(http.StatusBadRequest, "INSUFFICIENT_BALANCE", "Insufficient balance")
	}

	// OPTIMISTIC LOCKING
	// ----------------------------------------------------------------------
	// reduce sender wallet & add receiver wallet with tx and checking version
	// Trừ tiền ví người gửi đồng thời kiểm tra version để tránh xung đột đồng thời.
	newSenderBalance := senderWallet.Balance.Sub(req.Amount)
	err = s.walletRepo.UpdateBalanceTx(ctx, tx, senderWallet.ID, newSenderBalance, senderWallet.Version)
	if err != nil {
		// if failed because version mismatch, return special error so client can retry
		return nil, customErr.NewAppError(http.StatusConflict, "CONCURRENCY_CONFLICT", "Transaction is busy, please retry in the following minutes")
	}

	// Cộng tiền vào ví người nhận kèm theo kiểm tra version tương tự.
	newReceiverBalance := receiverWallet.Balance.Add(req.Amount)
	err = s.walletRepo.UpdateBalanceTx(ctx, tx, receiverWallet.ID, newReceiverBalance, receiverWallet.Version)
	if err != nil {
		return nil, customErr.NewAppError(http.StatusConflict, "CONCURRENCY_CONFLICT", "Transaction is busy, please retry in the following minutes")
	}

	// create data transaction record // Tạo bản ghi lịch sử giao dịch chính với trạng thái "success".
	transactionID := uuid.New().String()
	transaction := &model.Transaction{
		ID:               transactionID,
		SenderWalletID:   &senderWallet.ID,
		ReceiverWalletID: receiverWallet.ID,
		Amount:           req.Amount,
		Description:      req.Description,
		IdempotencyKey:   req.IdempotencyKey,
		Status:           "success",
	}
	if err = s.txRepo.CreateTx(ctx, tx, transaction); err != nil {
		return nil, customErr.ErrInternalServer
	}

	// create two ledger rows (debit for sender, and credit for receiver)
	debitEntry := &ledgerModel.LedgerEntry{
		ID:            uuid.New().String(),
		WalletID:      senderWallet.ID,
		TransactionID: transactionID,
		EntryType:     "debit",
		Amount:        req.Amount,
	}
	if err := s.ledgerRepo.CreateTx(ctx, tx, debitEntry); err != nil {
		return nil, customErr.ErrInternalServer
	}

	// Tạo 2 dòng ghi sổ cái (Ledger): một dòng Debit (ghi nợ) cho người gửi và một dòng Credit (ghi có) cho người nhận.
	creditEntry := &ledgerModel.LedgerEntry{
		ID:            uuid.New().String(),
		WalletID:      receiverWallet.ID,
		TransactionID: transactionID,
		EntryType:     "credit",
		Amount:        req.Amount,
	}
	if err := s.ledgerRepo.CreateTx(ctx, tx, creditEntry); err != nil {
		return nil, customErr.ErrInternalServer
	}

	// commit the db transaction // Xác nhận hoàn tất (Commit) toàn bộ database transaction.
	if err := tx.Commit(); err != nil {
		return nil, customErr.ErrInternalServer
	}

	// invalidate cache
	// Xóa cache số dư ví của cả người gửi và người nhận trên Redis (bất đồng bộ để không làm chậm response trả về cho client).
	senderCacheKey := "wallet:user:" + senderUserID
	receiverCacheKey := "wallet:user:" + receiverUser.ID

	// delete the cache keys asynchronously (don't block HTTP response)
	go func() {
		s.rdb.Del(context.Background(), senderCacheKey, receiverCacheKey)
	}()

	return transaction, nil
}

// GetHistory truy xuất lịch sử giao dịch của người dùng kèm theo tính toán phân trang.
func (s *transactionService) GetHistory(ctx context.Context, userID string, params model.PaginationParams) ([]model.Transaction, *model.PaginationMeta, error) {
	// Lấy thông tin ví của người dùng dựa vào userID.
	wallet, err := s.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, nil, customErr.NewAppError(http.StatusNotFound, "WALLET_NOT_FOUND", "Wallet not found")
	}

	// max limit
	// Giới hạn số lượng bản ghi tối đa trên một trang là 100
	if params.Limit > 100 {
		params.Limit = 100
	}

	// Gọi repository để lấy danh sách giao dịch phân trang và tổng số lượng bản ghi.
	txs, total, err := s.txRepo.GetHistory(ctx, wallet.ID, params)
	if err != nil {
		return nil, nil, customErr.ErrInternalServer
	}

	// Tính toán tổng số trang dựa trên tổng số bản ghi và giới hạn mỗi trang.
	totalPages := int(total / int64(params.Limit))
	if total%int64(params.Limit) != 0 {
		totalPages++
	}

	// Đóng gói thông tin metadata phục vụ phân trang cho client.
	meta := &model.PaginationMeta{
		Page:      params.Page,
		Limit:     params.Limit,
		Total:     total,
		TotalPage: totalPages,
	}

	return txs, meta, nil
}
