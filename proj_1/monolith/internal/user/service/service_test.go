package service

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	userModel "github.com/bashocode/gowallet/monolith/internal/user/model"
	userRepo "github.com/bashocode/gowallet/monolith/internal/user/repository"
	walletRepo "github.com/bashocode/gowallet/monolith/internal/wallet/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegister_Success(t *testing.T) {
	// sql mock
	//Tạo ra một kết nối database giả (db) và đối tượng dbMock để ép buộc Service phải thực hiện đúng các lệnh transaction mong muốn.
	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// initiate mock repositories
	//Khởi tạo các bản mock repository để chúng ta dễ dàng "bơm" dữ liệu đầu vào hoặc kết quả trả về theo ý muốn.
	mockUserRepo := new(userRepo.MockUserRepository)
	mockWalletRepo := new(walletRepo.MockWalletRepository)
	svc := NewUserService(db, mockUserRepo, mockWalletRepo)

	ctx := context.TODO()
	req := userModel.CreateUserRequest{
		FullName: "John Doe",
		Email:    "phongluong3366@gmail.com",
		Password: "123456",
	}

	// expect sql transaction
	//Ép service phải gọi mở transaction và commit thành công.
	dbMock.ExpectBegin()
	dbMock.ExpectCommit()

	// expect mock
	mockUserRepo.On("GetByEmail", ctx, "john.doe@example.com").Return(nil, errors.New("user not found"))
	mockUserRepo.On("CreateTx", ctx, mock.Anything, mock.Anything).Return(nil)
	mockWalletRepo.On("CreateTx", ctx, mock.Anything, mock.Anything).Return(nil)

	expectedUser := &userModel.User{
		ID:       "some-uuid",
		FullName: req.FullName,
		Email:    req.Email,
	}
	mockUserRepo.On("GetByID", ctx, mock.Anything).Return(expectedUser, nil)

	// run the function
	user, err := svc.Register(ctx, req)

	// verify result
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, req.FullName, user.FullName)
	assert.Equal(t, req.Email, user.Email)

	// make sure all mock if called and expected
	mockUserRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()

	mockUserRepo := new(userRepo.MockUserRepository)
	mockWalletRepo := new(walletRepo.MockWalletRepository)
	svc := NewUserService(db, mockUserRepo, mockWalletRepo)

	ctx := context.TODO()
	req := userModel.CreateUserRequest{
		FullName: "John Doe",
		Email:    "phongluong3366@gmail.com",
		Password: "123456",
	}

	// email already in db
	existingUser := &userModel.User{
		FullName: "John Doe",
		Email:    "phongluong3366@gmail.com",
	}
	mockUserRepo.On("GetByEmail", ctx, req.Email).Return(existingUser, nil)

	// run the function
	user, err := svc.Register(ctx, req)

	// verify result should be error and no return the user object
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "this email already registered.", err.Error())

	// make sure all mock if called and expected
	mockUserRepo.AssertExpectations(t)
}
