package mail

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	acsAPIVersion  = "2023-03-31"
	acsSendTimeout = 10 * time.Second
)

// acsMailer sends through the Azure Communication Services Email REST API,
// authenticated with the HMAC scheme from the resource's connection string.
// There is no official Go SDK for ACS Email.
type acsMailer struct {
	scheme string
	host   string
	key    []byte
	sender string
	client *http.Client
}

// NewACSMailer takes the Communication Service's connection string
// ("endpoint=https://...;accesskey=...") and the MailFrom address.
func NewACSMailer(connectionString, sender string) (Mailer, error) {
	var endpoint, accessKey string
	for _, part := range strings.Split(connectionString, ";") {
		// Cut at the first '=' only: the base64 key can end in '=' padding
		k, v, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		switch strings.ToLower(strings.TrimSpace(k)) {
		case "endpoint":
			endpoint = strings.TrimSpace(v)
		case "accesskey":
			accessKey = strings.TrimSpace(v)
		}
	}
	if endpoint == "" || accessKey == "" {
		return nil, errors.New("ACS_CONNECTION_STRING must contain endpoint and accesskey")
	}
	u, err := url.Parse(endpoint)
	if err != nil || u.Host == "" {
		return nil, fmt.Errorf("ACS_CONNECTION_STRING has an invalid endpoint %q", endpoint)
	}
	key, err := base64.StdEncoding.DecodeString(accessKey)
	if err != nil {
		return nil, errors.New("ACS_CONNECTION_STRING has an accesskey that is not valid base64")
	}
	if sender == "" {
		return nil, errors.New("MAIL_SENDER is required when MAIL_MODE=acs")
	}
	return &acsMailer{
		scheme: u.Scheme,
		host:   u.Host,
		key:    key,
		sender: sender,
		client: &http.Client{},
	}, nil
}

type acsAddress struct {
	Address     string `json:"address"`
	DisplayName string `json:"displayName,omitempty"`
}

type acsRequest struct {
	SenderAddress string `json:"senderAddress"`
	Recipients    struct {
		To []acsAddress `json:"to"`
	} `json:"recipients"`
	Content struct {
		Subject   string `json:"subject"`
		PlainText string `json:"plainText"`
		HTML      string `json:"html,omitempty"`
	} `json:"content"`
}

func (m *acsMailer) Send(ctx context.Context, msg Message) error {
	var payload acsRequest
	payload.SenderAddress = m.sender
	payload.Recipients.To = []acsAddress{{Address: msg.To, DisplayName: msg.ToName}}
	payload.Content.Subject = msg.Subject
	payload.Content.PlainText = msg.PlainText
	payload.Content.HTML = msg.HTML
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, acsSendTimeout)
	defer cancel()

	pathAndQuery := "/emails:send?api-version=" + acsAPIVersion
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.scheme+"://"+m.host+pathAndQuery, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	m.sign(req, pathAndQuery, body, time.Now())

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("acs email: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusAccepted:
		return nil
	case http.StatusTooManyRequests:
		return ErrRateLimited
	}
	// the error body never echoes the message content, so it is safe to log
	var e struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(&e)
	return fmt.Errorf("acs email: status %d: %s %s", resp.StatusCode, e.Error.Code, e.Error.Message)
}

// sign adds the HMAC-SHA256 headers ACS expects:
// StringToSign = VERB \n path?query \n date;host;content-hash
func (m *acsMailer) sign(req *http.Request, pathAndQuery string, body []byte, now time.Time) {
	date := now.UTC().Format(http.TimeFormat)
	sum := sha256.Sum256(body)
	contentHash := base64.StdEncoding.EncodeToString(sum[:])

	toSign := req.Method + "\n" + pathAndQuery + "\n" + date + ";" + m.host + ";" + contentHash
	mac := hmac.New(sha256.New, m.key)
	mac.Write([]byte(toSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	req.Header.Set("x-ms-date", date)
	req.Header.Set("x-ms-content-sha256", contentHash)
	req.Header.Set("Authorization", "HMAC-SHA256 SignedHeaders=x-ms-date;host;x-ms-content-sha256&Signature="+signature)
}
