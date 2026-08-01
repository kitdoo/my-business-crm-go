// Package mailer defines the Service interface for sending outbound email
// over SMTP; the implementation lives in the mailer subpackage.
package mailer

import "context"

// Message is a plain-text email to send.
type Message struct {
	To      string
	Subject string
	Body    string
	// FromName, when set, is used as the display name on the From header
	// (e.g. "TOM STUDIO 021 DOO <you@gmail.com>") instead of the bare
	// configured From address — the mailbox itself is unchanged, but the
	// recipient no longer sees it as coming from a personal account.
	FromName string
	// ReplyTo, when set, lets the recipient reply directly to the form
	// submitter instead of the configured From address.
	ReplyTo string
	// Attachments are sent alongside Body, e.g. a generated invoice PDF
	// (see internal/services/invoice). Empty for every other caller today.
	Attachments []Attachment
}

// Attachment is one file attached to a Message.
type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

// Service sends outbound email.
type Service interface {
	Send(ctx context.Context, msg Message) error
}
