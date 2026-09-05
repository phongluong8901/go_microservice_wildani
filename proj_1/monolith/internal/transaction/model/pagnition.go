package model

type PaginationParams struct {
	Page   int    `form:"page,default=1"`
	Limit  int    `form:"limit,default=10"`
	Sort   string `form:"sort,default=created_at"`
	Order  string `form:"order,default=desc"`
	Status string `form:"status"` // optional filter status
}

// Xác định vị trí bản ghi bắt đầu trong cơ sở dữ liệu để phục vụ cho câu lệnh LIMIT ... OFFSET .... Ví dụ, nếu người dùng xem trang 2 với limit 10, offset sẽ là (2 - 1) * 10 = 10, nghĩa là bỏ qua 10 bản ghi đầu để lấy 10 bản ghi tiếp theo.
func (p *PaginationParams) Offset() int {
	return (p.Page - 1) * p.Limit
}

type PaginationMeta struct {
	Page      int   `json:"page"`
	Limit     int   `json:"limit"`
	Total     int64 `json:"total"`
	TotalPage int   `json:"total_page"`
}

type PaginatedResponse struct {
	Success bool           `json:"success"`
	Data    any            `json:"data"`
	Meta    PaginationMeta `json:"meta"`
}

// Các struct tags form:"...,default=..." hỗ trợ các web framework (như Gin) tự động ánh xạ dữ liệu từ request và gán sẵn giá trị mặc định (trang 1, mỗi trang 10 bản ghi, sắp xếp theo created_at giảm dần) nếu người dùng không truyền lên.
