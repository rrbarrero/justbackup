package smtp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/rrbarrero/justbackup/internal/notification/domain/entities"
	"github.com/rrbarrero/justbackup/internal/notification/domain/interfaces"
	"github.com/rrbarrero/justbackup/internal/notification/domain/valueobjects"
)

func init() {
	interfaces.RegisterProvider("smtp", func(config json.RawMessage) (interfaces.NotificationProvider, bool, error) {
		var c entities.SMTPConfig
		if err := json.Unmarshal(config, &c); err != nil {
			return nil, false, err
		}
		return NewSMTPProvider(c.Host, c.Port, c.User, c.Password, c.From, c.To), c.NotifyOnSuccess, nil
	})
}

type SMTPProvider struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
	To       []string
}

func NewSMTPProvider(host string, port int, user, password, from string, to []string) *SMTPProvider {
	return &SMTPProvider{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		From:     from,
		To:       to,
	}
}

func (p *SMTPProvider) Send(ctx context.Context, title, message string, level valueobjects.NotificationLevel) error {
	auth := smtp.PlainAuth("", p.User, p.Password, p.Host)
	addr := fmt.Sprintf("%s:%d", p.Host, p.Port)

	subject := fmt.Sprintf("[%s] %s", strings.ToUpper(string(level)), title)

	// Construct email headers
	headers := make(map[string]string)
	headers["From"] = p.From
	headers["To"] = strings.Join(p.To, ", ")
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/plain; charset=\"utf-8\""
	headers["Date"] = time.Now().Format(time.RFC1123Z)
	headers["Message-ID"] = fmt.Sprintf("<%d.%s@%s>", time.Now().UnixNano(), "justbackup", p.Host)

	messageBody := ""
	for k, v := range headers {
		messageBody += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	messageBody += "\r\n" + message

	err := smtp.SendMail(addr, auth, p.From, p.To, []byte(messageBody))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (p *SMTPProvider) Type() string {
	return "smtp"
}
