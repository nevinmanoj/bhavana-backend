package mail

import (
	"context"
	"log"
)

// logMailer prints emails instead of sending them. Local development only:
// the output includes generated passwords and reset links.
type logMailer struct{}

func NewLogMailer() Mailer {
	return logMailer{}
}

func (logMailer) Send(_ context.Context, msg Message) error {
	log.Printf("[mail:log] to=%s subject=%q\n%s", msg.To, msg.Subject, msg.PlainText)
	return nil
}
