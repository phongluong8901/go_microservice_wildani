package repository

import (
	"context"
	"database/sql"

	"github.com/bashocode/gowallet/monolith/internal/wallet/model"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
)

type MockWalletRepository struct {
	mock.Mock
}

// Hàm giả lập tạo mới ví bên trong một transaction cơ sở dữ liệu.
func (m *MockWalletRepository) CreateTx(ctx context.Context, tx *sql.Tx, w *model.Wallet) error {
	args := m.Called(ctx, tx, w)
	return args.Error(0)
}

// Hàm giả lập lấy thông tin ví dựa theo userID, trả về con trỏ ví và lỗ
func (m *MockWalletRepository) GetByUserID(ctx context.Context, userID string) (*model.Wallet, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Wallet), args.Error(1)
}

// Hàm giả lập cập nhật số dư ví bên trong transaction (áp dụng Optimistic Locking).
func (m *MockWalletRepository) UpdateBalanceTx(ctx context.Context, tx *sql.Tx, walletID string, newBalance decimal.Decimal, currentVersion int) error {
	args := m.Called(ctx, tx, walletID, newBalance, currentVersion)
	return args.Error(0)
}
