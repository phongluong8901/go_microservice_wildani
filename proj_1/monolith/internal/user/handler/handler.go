package handler

import (
	"net/http"

	customError "github.com/bashocode/gowallet/monolith/internal/errors"
	"github.com/bashocode/gowallet/monolith/internal/user/model"
	"github.com/bashocode/gowallet/monolith/internal/user/service"
	"github.com/gin-gonic/gin"
)

// Struct gom nhóm các hàm xử lý HTTP, chứa dependency svc service.UserService để giao tiếp với tầng nghiệp vụ.
type UserHandler struct {
	svc service.UserService
}

// Khởi tạo một đối tượng UserHandler mới bằng cách truyền vào tầng service (Dependency Injection).
func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// Xử lý HTTP POST Đăng ký
func (h *UserHandler) Register(c *gin.Context) {
	var req model.CreateUserRequest
	//Đọc và parse dữ liệu JSON từ HTTP Body của client vào struct CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		//register the error input to gin context
		c.Error(customError.NewAppError(http.StatusBadRequest, "INVALID_ID_INPUT", err.Error()))
		return
	}

	//Gọi tầng service để thực hiện logic đăng ký.
	user, err := h.svc.Register(c.Request.Context(), req)
	if err != nil {
		//register the error to middleware
		c.Error(err)
		return
	}

	//Đăng ký thành công, trả về dữ liệu user vừa tạo kèm mã trạng thái 201 Created.
	c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	//Lấy giá trị tham số id từ đường dẫn URL (ví dụ /api.v1.users/123).
	id := c.Param("id")
	//Gọi service để lấy thông tin user từ database
	user, err := h.svc.GetProfile(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	//Thành công, trả về thông tin user kèm mã 200 OK
	c.JSON(http.StatusOK, user)
}

// Xử lý HTTP PUT Cập nhật thông tin
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	//Lấy ID của user cần cập nhật từ đường dẫn URL.
	id := c.Param("id")
	//Parse dữ liệu JSON cập nhật mới từ client vào struct UpdateUserRequest.
	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(customError.NewAppError(http.StatusBadRequest, "INVALID_ID_INPUT", err.Error()))
		return
	}

	//Gọi tầng service để cập nhật thông tin trong database và trả về user sau khi đã thay đổi thành công kèm mã 200 OK.
	user, err := h.svc.UpdateProfile(c.Request.Context(), id, req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, user)
}
