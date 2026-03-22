package smtp

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/smtp"
	"strings"

	"sourcecraft.dev/benzo/bff/internal/application/ports"
	"sourcecraft.dev/benzo/bff/internal/domain"
)

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	UseTLS   bool
}

type Sender struct {
	host     string
	port     int
	username string
	password string
	from     string
	useTLS   bool
	disabled bool
}

func NewSender(cfg Config) *Sender {
	if strings.TrimSpace(cfg.Host) == "" {
		return &Sender{disabled: true}
	}

	from := strings.TrimSpace(cfg.From)
	if from == "" {
		if strings.TrimSpace(cfg.Username) != "" {
			from = strings.TrimSpace(cfg.Username)
		} else {
			from = "no-reply@profdnk.local"
		}
	}

	return &Sender{
		host:     strings.TrimSpace(cfg.Host),
		port:     cfg.Port,
		username: strings.TrimSpace(cfg.Username),
		password: cfg.Password,
		from:     from,
		useTLS:   cfg.UseTLS,
	}
}

func (s *Sender) SendReport(ctx context.Context, message ports.ReportEmail) error {
	if s.disabled {
		return domain.ErrReportDeliveryDisabled
	}

	address := net.JoinHostPort(s.host, fmt.Sprintf("%d", s.port))
	dialer := &net.Dialer{}
	rawConn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("dial smtp server: %w", err)
	}
	defer rawConn.Close()

	var conn net.Conn = rawConn
	if s.useTLS {
		tlsConn := tls.Client(rawConn, &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: s.host,
		})

		if err := tlsConn.HandshakeContext(ctx); err != nil {
			return fmt.Errorf("smtp tls handshake: %w", err)
		}

		conn = tlsConn
	}

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return fmt.Errorf("create smtp client: %w", err)
	}
	defer client.Close()

	if !s.useTLS {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(&tls.Config{
				MinVersion: tls.VersionTLS12,
				ServerName: s.host,
			}); err != nil {
				return fmt.Errorf("starttls: %w", err)
			}
		}
	}

	if s.username != "" {
		auth := smtp.PlainAuth("", s.username, s.password, s.host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	if err := client.Mail(s.from); err != nil {
		return fmt.Errorf("smtp MAIL FROM: %w", err)
	}

	if err := client.Rcpt(message.To); err != nil {
		return fmt.Errorf("smtp RCPT TO: %w", err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp DATA: %w", err)
	}

	payload := buildMessage(s.from, message)
	if _, err := writer.Write(payload); err != nil {
		_ = writer.Close()
		return fmt.Errorf("write smtp payload: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("close smtp writer: %w", err)
	}

	if err := client.Quit(); err != nil {
		return fmt.Errorf("smtp quit: %w", err)
	}

	return nil
}

func buildMessage(from string, message ports.ReportEmail) []byte {
	boundary := "profdnk-boundary"
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("From: %s\r\n", from))
	buffer.WriteString(fmt.Sprintf("To: %s\r\n", message.To))
	buffer.WriteString(fmt.Sprintf("Subject: %s\r\n", message.Subject))
	buffer.WriteString("MIME-Version: 1.0\r\n")
	buffer.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%q\r\n", boundary))
	buffer.WriteString("\r\n")

	buffer.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	buffer.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	buffer.WriteString("Content-Transfer-Encoding: 7bit\r\n")
	buffer.WriteString("\r\n")
	buffer.WriteString(message.Body)
	buffer.WriteString("\r\n")

	buffer.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	buffer.WriteString(fmt.Sprintf("Content-Type: %s\r\n", message.ContentType))
	buffer.WriteString("Content-Transfer-Encoding: base64\r\n")
	buffer.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=%q\r\n", message.FileName))
	buffer.WriteString("\r\n")
	buffer.WriteString(wrapBase64(message.Attachment))
	buffer.WriteString("\r\n")
	buffer.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	return buffer.Bytes()
}

func wrapBase64(content []byte) string {
	encoded := base64.StdEncoding.EncodeToString(content)
	if encoded == "" {
		return ""
	}

	var lines []string
	for len(encoded) > 76 {
		lines = append(lines, encoded[:76])
		encoded = encoded[76:]
	}
	lines = append(lines, encoded)

	return strings.Join(lines, "\r\n")
}
