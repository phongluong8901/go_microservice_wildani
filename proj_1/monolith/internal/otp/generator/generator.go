package generator

import (
	"crypto/rand"
)

// GenerateOTP generates a cryptographically secure random numeric OTP of the specified length.
// It avoids using rand.Int by reading from crypto/rand.Reader directly and using rejection
// sampling to eliminate modulo bias.
// GenerateOTP tạo một mã OTP số ngẫu nhiên an toàn mật mã
func GenerateOTP(length int) (string, error) {
	const charset = "0123456789" // Định nghĩa tập ký tự chữ số dùng để tạo mã OTP.
	otp := make([]byte, length)  // Khởi tạo một slice byte để chứa mã OTP
	num := make([]byte, 1)       // Khởi tạo một slice byte để đọc dữ liệu ngẫu nhiên
	for i := 0; i < length; {
		_, err := rand.Read(num) // Đọc dữ liệu ngẫu nhiên từ crypto/rand.Reader
		if err != nil {
			return "", err
		}
		val := num[0] // Lấy giá trị byte ngẫu nhiên
		// 256 % 10 = 6. 256 - 6 = 250.
		// To avoid modulo bias, discard values >= 250. // Để tránh sai lệch do phép chia lấy dư, loại bỏ các giá trị >= 250.
		if val < 250 {
			otp[i] = charset[val%10]
			i++
		}
	}
	return string(otp), nil
}
