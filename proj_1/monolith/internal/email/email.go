package email

import (
	"context"
	"fmt"
	"net/smtp"

	"github.com/bashocode/gowallet/monolith/internal/logger"
)

type EmailSender interface {
	SendEmail(ctx context.Context, to string, subject string, body string) error
}

type smtpEmailSender struct {
	host string
	port string
	from string
}

func NewSMTPEmailSender(host string, port string, from string) EmailSender {
	return &smtpEmailSender{
		host: host,
		port: port,
		from: from,
	}
}

// SendEmail thực hiện soạn nội dung và gửi email tới người nhận thông qua SMTP protocol.
func (s *smtpEmailSender) SendEmail(ctx context.Context, to string, subject string, body string) error {
	// Định dạng nội dung email theo chuẩn RFC với tiêu đề To, Subject và phần thân body cách nhau bởi ký tự xuống dòng chuẩn (\r\n).
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s\r\n", to, subject, body))

	// for local testing with mailhog, we send without auth
	addr := s.host + ":" + s.port
	// Gửi email qua giao thức SMTP
	err := smtp.SendMail(addr, nil, s.from, []string{to}, msg)
	if err != nil {
		logger.Log.Error("Failed to send email via SMTP", "to", to, "error", err)
		return err
	}

	logger.Log.Info("Email sent successfully", "to", to)
	return nil
}
