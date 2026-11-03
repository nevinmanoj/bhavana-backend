package auth

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"
)

func TestGeneratePassword(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 1000; i++ {
		p, err := GeneratePassword()
		if err != nil {
			t.Fatal(err)
		}
		if len(p) != generatedPasswordLength {
			t.Fatalf("length %d, want %d: %q", len(p), generatedPasswordLength, p)
		}
		for _, c := range p {
			if !strings.ContainsRune(generatedPasswordAlphabet, c) {
				t.Fatalf("character %q outside the alphabet in %q", c, p)
			}
		}
		if seen[p] {
			t.Fatalf("duplicate password %q", p)
		}
		seen[p] = true
	}
}

func TestGeneratedPasswordRoundTripsThroughBcrypt(t *testing.T) {
	p, _ := GeneratePassword()
	hash, err := HashPassword(p)
	if err != nil {
		t.Fatal(err)
	}
	if CheckPassword(p, hash) != nil {
		t.Error("generated password does not match its own hash")
	}
}

// "" in users.password_hash means "no password yet"; login relies on bcrypt
// rejecting it whatever is typed.
func TestEmptyHashRejectsEveryPassword(t *testing.T) {
	for _, p := range []string{"", "secret1"} {
		if CheckPassword(p, "") == nil {
			t.Errorf("empty hash accepted %q", p)
		}
	}
}

func TestNewResetToken(t *testing.T) {
	token, hash, err := NewResetToken()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(HashResetToken(token), hash) {
		t.Error("HashResetToken(token) differs from the returned hash")
	}
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(raw) != 32 {
		t.Errorf("token %q is not 32 bytes of URL-safe base64", token)
	}
	other, _, _ := NewResetToken()
	if other == token {
		t.Error("two tokens are equal")
	}
}
