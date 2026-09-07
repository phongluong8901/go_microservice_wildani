package proxy // Khai báo package proxy để xử lý chuyển hướng request giữa các microservice

import (
	"net/http"          // Thư viện chuẩn xử lý HTTP server và request/response
	"net/http/httputil" // Thư viện chuẩn của Go hỗ trợ xây dựng Reverse Proxy
	"net/url"           // Thư viện chuẩn phân tích cú pháp và xử lý URL

	"github.com/google/uuid" // Thư viện tạo mã định danh duy nhất UUID
)

type ReverseProxy struct { // Định nghĩa cấu trúc quản lý proxy
	target *url.URL               // URL đích mà proxy sẽ chuyển tiếp request tới
	proxy  *httputil.ReverseProxy // Đối tượng ReverseProxy có sẵn của thư viện chuẩn Go
}

func NewReverseProxy(targetURL string) (*ReverseProxy, error) { // Hàm khởi tạo một instance ReverseProxy từ chuỗi URL đích
	url, err := url.Parse(targetURL) // Phân tích chuỗi URL đích thành kiểu *url.URL
	if err != nil {                  // Kiểm tra nếu cấu trúc URL không hợp lệ
		return nil, err // Trả về lỗi nếu quá trình parse thất bại
	}

	// Create Go's built-in reverse proxy with Rewrite only
	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(url)
			r.Out.Header.Set("X-Forwarded-Host", r.In.Header.Get("Host"))

			// Inject & forward Request Correlation ID for distributed logging
			corID := r.In.Header.Get("X-Correlation-ID")
			if corID == "" {
				corID = uuid.New().String()
			}
			r.Out.Header.Set("X-Correlation-ID", corID)
		},
	}

	return &ReverseProxy{ // Trả về con trỏ cấu trúc ReverseProxy đã cấu hình hoàn chỉnh
		target: url,
		proxy:  proxy,
	}, nil
}

func (p *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) { // Triển khai interface http.Handler để nhận và xử lý request HTTP
	p.proxy.ServeHTTP(w, r) // Ủy quyền việc chuyển tiếp request và trả về response cho proxy mặc định của Go
}
