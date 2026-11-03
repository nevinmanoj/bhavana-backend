package mail

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var testKey = base64.StdEncoding.EncodeToString([]byte("test-access-key-0123456789"))

func newTestMailer(t *testing.T, handler http.HandlerFunc) Mailer {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	m, err := NewACSMailer("endpoint="+srv.URL+"/;accesskey="+testKey, "DoNotReply@example.azurecomm.net")
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestACSSendSignsAndShapesTheRequest(t *testing.T) {
	m := newTestMailer(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		if r.Method != http.MethodPost || r.URL.Path != "/emails:send" || r.URL.Query().Get("api-version") != acsAPIVersion {
			t.Errorf("unexpected request %s %s", r.Method, r.URL)
		}

		sum := sha256.Sum256(body)
		contentHash := base64.StdEncoding.EncodeToString(sum[:])
		if r.Header.Get("x-ms-content-sha256") != contentHash {
			t.Error("content hash header does not match the body")
		}
		key, _ := base64.StdEncoding.DecodeString(testKey)
		toSign := "POST\n" + r.URL.RequestURI() + "\n" + r.Header.Get("x-ms-date") + ";" + r.Host + ";" + contentHash
		mac := hmac.New(sha256.New, key)
		mac.Write([]byte(toSign))
		want := "HMAC-SHA256 SignedHeaders=x-ms-date;host;x-ms-content-sha256&Signature=" + base64.StdEncoding.EncodeToString(mac.Sum(nil))
		if r.Header.Get("Authorization") != want {
			t.Errorf("Authorization = %q, want %q", r.Header.Get("Authorization"), want)
		}

		var got acsRequest
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatal(err)
		}
		if got.SenderAddress != "DoNotReply@example.azurecomm.net" ||
			len(got.Recipients.To) != 1 || got.Recipients.To[0].Address != "anu@example.com" || got.Recipients.To[0].DisplayName != "Anu" ||
			got.Content.Subject != "Hi" || got.Content.PlainText != "text" || got.Content.HTML != "<p>html</p>" {
			t.Errorf("unexpected payload %s", body)
		}
		w.WriteHeader(http.StatusAccepted)
	})

	err := m.Send(context.Background(), Message{To: "anu@example.com", ToName: "Anu", Subject: "Hi", PlainText: "text", HTML: "<p>html</p>"})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
}

func TestACSSendErrors(t *testing.T) {
	m := newTestMailer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	})
	if err := m.Send(context.Background(), Message{To: "a@b.c"}); !errors.Is(err, ErrRateLimited) {
		t.Errorf("429: got %v, want ErrRateLimited", err)
	}

	m = newTestMailer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":{"code":"InvalidSenderUserName","message":"bad sender"}}`))
	})
	err := m.Send(context.Background(), Message{To: "a@b.c"})
	if err == nil || !strings.Contains(err.Error(), "InvalidSenderUserName") || errors.Is(err, ErrRateLimited) {
		t.Errorf("400: got %v, want an error carrying the code", err)
	}
}

func TestNewACSMailerValidatesConfig(t *testing.T) {
	cases := map[string]struct{ conn, sender string }{
		"no endpoint":   {"accesskey=" + testKey, "x@y.z"},
		"no key":        {"endpoint=https://acs.example.com/", "x@y.z"},
		"key not b64":   {"endpoint=https://acs.example.com/;accesskey=***", "x@y.z"},
		"no sender":     {"endpoint=https://acs.example.com/;accesskey=" + testKey, ""},
		"relative host": {"endpoint=acs;accesskey=" + testKey, "x@y.z"},
	}
	for name, c := range cases {
		if _, err := NewACSMailer(c.conn, c.sender); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
	// base64 padding '=' in the key must survive parsing
	if _, err := NewACSMailer("endpoint=https://acs.example.com/;accesskey="+base64.StdEncoding.EncodeToString([]byte("k")), "x@y.z"); err != nil {
		t.Errorf("padded key rejected: %v", err)
	}
}

func TestNewRequiresAMode(t *testing.T) {
	if _, err := New("", "", ""); err == nil {
		t.Error("empty MAIL_MODE accepted")
	}
	if _, err := New("smtp", "", ""); err == nil {
		t.Error("unknown MAIL_MODE accepted")
	}
	if m, err := New("log", "", ""); err != nil || m == nil {
		t.Errorf("log mode: %v", err)
	}
}
