package mail

import (
	"context"
	"errors"
	"fmt"
)

type Message struct {
	To        string
	ToName    string
	Subject   string
	PlainText string
	HTML      string
}

type Mailer interface {
	Send(ctx context.Context, msg Message) error
}

var ErrRateLimited = errors.New("mail rate limited")

// New picks the mailer for MAIL_MODE. There is deliberately no default: a
// deployed backend with a missing setting must fail to start rather than fall
// back to printing passwords and reset links to its logs.
func New(mode, acsConnectionString, sender string) (Mailer, error) {
	switch mode {
	case "acs":
		return NewACSMailer(acsConnectionString, sender)
	case "log":
		return NewLogMailer(), nil
	default:
		return nil, fmt.Errorf("MAIL_MODE must be \"acs\" or \"log\", got %q", mode)
	}
}
