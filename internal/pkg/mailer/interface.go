// Package mailer defines the Service interface for sending outbound email
// over SMTP; the implementation lives in the mailer subpackage.
package mailer

import "context"

// Message is a plain-text email to send.
type Message struct {
	To      string
	Subject string
	Body    string
	// ReplyTo, when set, lets the recipient reply directly to the form
	// submitter instead of the configured From address.
	ReplyTo string
}

// Service sends outbound email.
type Service interface {
	Send(ctx context.Context, msg Message) error
}
