// Package resend implements the mailer.Service interface over the
// Resend HTTP API (https://resend.com/docs/api-reference/emails/send-email),
// as an alternative to the mailer/mailer package's direct-SMTP delivery —
// direct SMTP to providers like Gmail is commonly blackholed from cloud/VPS
// source IPs on reputation grounds, where an HTTPS API call is not.
package resend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kitdoo/my-business-crm-go/internal/errs"
	mailersvc "github.com/kitdoo/my-business-crm-go/internal/pkg/mailer"
)

const apiURL = "https://api.resend.com/emails"

var _ mailersvc.Service = (*Service)(nil)

// Config is the resolved Resend API auth/from data a Service sends
// through. Built from appconfig.ResendConfig in internal/fx.
type Config struct {
	APIKey string
	From   string
}

// Service sends email through the Resend HTTP API.
type Service struct {
	cfg    *Config
	client *http.Client
}

// New builds a Service. cfg == nil yields a Service whose Send always
// returns errs.ErrMailerNotConfigured — see CRMConfig.Resend being optional.
func New(cfg *Config) *Service {
	return &Service{cfg: cfg, client: &http.Client{}}
}

type sendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Text    string   `json:"text"`
	ReplyTo string   `json:"reply_to,omitempty"`
}

// Send delivers msg via a single POST to the Resend emails endpoint.
func (s *Service) Send(ctx context.Context, msg mailersvc.Message) error {
	if s.cfg == nil {
		return errs.ErrMailerNotConfigured
	}

	body, err := json.Marshal(sendRequest{
		From:    s.cfg.From,
		To:      []string{msg.To},
		Subject: msg.Subject,
		Text:    msg.Body,
		ReplyTo: msg.ReplyTo,
	})
	if err != nil {
		return fmt.Errorf("marshal resend request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build resend request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.cfg.APIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("call resend api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("resend api returned %s: %s", resp.Status, respBody)
	}
	return nil
}
