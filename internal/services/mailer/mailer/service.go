// Package mailer implements the mailer.Service interface over the
// standard library's net/smtp.
package mailer

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/kitdoo/my-business-crm-go/internal/errs"
	mailersvc "github.com/kitdoo/my-business-crm-go/internal/services/mailer"
)

var _ mailersvc.Service = (*Service)(nil)

// Config is the resolved SMTP connection/auth data a Service sends
// through. Built from appconfig.SMTPConfig in internal/fx.
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// Service sends email over SMTP, auto-selecting implicit TLS (port 465,
// dialed directly with crypto/tls) vs STARTTLS (any other port,
// negotiated by net/smtp.SendMail itself when the server advertises it).
type Service struct {
	cfg *Config
}

// New builds a Service. cfg == nil yields a Service whose Send always
// returns errs.ErrSMTPNotConfigured — see CRMConfig.Smtp being optional.
func New(cfg *Config) *Service {
	return &Service{cfg: cfg}
}

// Send delivers msg synchronously. ctx is accepted for interface
// symmetry with other services but does not cancel the underlying
// net/smtp call — the standard library offers no hook for that.
func (s *Service) Send(_ context.Context, msg mailersvc.Message) error {
	if s.cfg == nil {
		return errs.ErrSMTPNotConfigured
	}

	body := buildMessage(s.cfg.From, msg)
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)

	var auth smtp.Auth
	if s.cfg.Username != "" {
		auth = smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	}

	if s.cfg.Port == 465 {
		return sendImplicitTLS(addr, s.cfg.Host, auth, s.cfg.From, msg.To, body)
	}
	return smtp.SendMail(addr, auth, s.cfg.From, []string{msg.To}, body)
}

func buildMessage(from string, msg mailersvc.Message) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", from)
	fmt.Fprintf(&b, "To: %s\r\n", msg.To)
	fmt.Fprintf(&b, "Subject: %s\r\n", msg.Subject)
	if msg.ReplyTo != "" {
		fmt.Fprintf(&b, "Reply-To: %s\r\n", msg.ReplyTo)
	}
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	b.WriteString("\r\n")
	b.WriteString(msg.Body)
	return []byte(b.String())
}

// sendImplicitTLS handles port 465, which expects TLS from the first
// byte of the connection (unlike 587/25's plaintext-then-STARTTLS, which
// smtp.SendMail already handles on its own).
func sendImplicitTLS(addr, host string, auth smtp.Auth, from, to string, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12})
	if err != nil {
		return fmt.Errorf("dial smtp over tls: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("create smtp client: %w", err)
	}
	defer client.Close()

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt to: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("write smtp body: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("close smtp body: %w", err)
	}
	return client.Quit()
}
