package logger

import (
	"context"
	"log/slog"
	"os"
)

var Log *slog.Logger

const CorrelationIDKey = "correlation_id"

// InitLogger khởi tạo cấu hình logger mặc định ghi log dưới dạng JSON ra chuẩn đầu ra tiêu chuẩn (stdout).
func InitLogger() {
	// set default structured JSON handler to stdout
	// Thiết lập JSON handler với mức độ lọc log tối thiểu là LevelInfo.
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo, // show the log level info
	})
	// Tạo instance logger mới từ handler vừa cấu hình.
	Log = slog.New(handler)
	// Đặt làm logger mặc định toàn cục của Go package slog.
	slog.SetDefault(Log)
}

// helper for log with context that automatically includes correlation id
// getLogArgs là hàm bổ trợ (helper) tự động trích xuất correlation_id từ context và gộp vào danh sách tham số log nếu có.
func getLogArgs(ctx context.Context, args []any) []any {
	// Kiểm tra xem trong context có lưu giá trị correlation_id dạng string hay không.
	if ctx != nil {
		if cid, ok := ctx.Value(CorrelationIDKey).(string); ok {
			return append(args, slog.String("correlation_id", cid))
		}
	}
	return args
}

// Info ghi log ở cấp độ thông tin (INFO) kèm theo ngữ cảnh
func Info(ctx context.Context, msg string, args ...any) {
	Log.InfoContext(ctx, msg, getLogArgs(ctx, args)...)
}

// Error ghi log ở cấp độ lỗi (ERROR) kèm theo ngữ cảnh context
func Error(ctx context.Context, msg string, args ...any) {
	Log.ErrorContext(ctx, msg, getLogArgs(ctx, args)...)
}

// Warn ghi log ở cấp độ cảnh báo (WARN) kèm theo ngữ cảnh context
func Warn(ctx context.Context, msg string, args ...any) {
	Log.WarnContext(ctx, msg, getLogArgs(ctx, args)...)
}
