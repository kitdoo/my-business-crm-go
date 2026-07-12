package entities

// NotificationMessage is a generic outbound message: the body is opaque,
// already composed by the caller (e.g. web-public/'s contact/dealer
// forms) — see notification.Service. It carries no id/timestamps — it is
// never persisted, only emailed on.
type NotificationMessage struct {
	// Subject is the email subject line. Empty falls back to a generic
	// subject — see notification.Service.Send.
	Subject string
	// Message is the full message body, already composed by the caller.
	Message string
	// ReplyTo is set as the email's Reply-To address.
	ReplyTo string
}
