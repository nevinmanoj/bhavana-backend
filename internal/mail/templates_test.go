package mail

import (
	"strings"
	"testing"
)

func TestLoginEmails(t *testing.T) {
	for _, m := range []Message{
		AcceptedLoginEmail("https://app.example/", "anu@example.com", "Anu", "St Mary's", "abcd2345efgh6789jkmn"),
		NewPasswordEmail("https://app.example/", "anu@example.com", "Anu", "St Mary's", "abcd2345efgh6789jkmn"),
	} {
		for _, part := range []string{m.PlainText, m.HTML} {
			for _, want := range []string{"anu@example.com", "abcd2345efgh6789jkmn", "https://app.example/login"} {
				if !strings.Contains(part, want) {
					t.Errorf("%q: missing %q", m.Subject, want)
				}
			}
		}
		if m.To != "anu@example.com" || m.ToName != "Anu" {
			t.Errorf("%q: wrong recipient %q <%s>", m.Subject, m.ToName, m.To)
		}
	}
}

func TestPasswordResetEmail(t *testing.T) {
	m := PasswordResetEmail("https://app.example", "anu@example.com", "Anu", "tok-en_123")
	want := "https://app.example/reset-password?token=tok-en_123"
	if !strings.Contains(m.PlainText, want) || !strings.Contains(m.HTML, want) {
		t.Errorf("reset link %q missing", want)
	}
}

func TestEveryEmailRenders(t *testing.T) {
	for _, m := range []Message{
		RegistrationReceivedEmail("anu@example.com", "Anu", "St Mary's"),
		AcceptedLoginEmail("https://app.example", "anu@example.com", "Anu", "St Mary's", "pw"),
		NewPasswordEmail("https://app.example", "anu@example.com", "Anu", "St Mary's", "pw"),
		PasswordResetEmail("https://app.example", "anu@example.com", "Anu", "t"),
	} {
		if m.Subject == "" || m.To != "anu@example.com" || m.ToName != "Anu" {
			t.Errorf("%q: wrong headers: %+v", m.Subject, m)
		}
		if !strings.HasPrefix(m.HTML, "<!DOCTYPE html>") || !strings.Contains(m.HTML, "<body") {
			t.Errorf("%q: HTML is missing the layout", m.Subject)
		}
		if !strings.Contains(m.HTML, "Hello Anu,") || !strings.HasPrefix(m.PlainText, "Hello Anu,") {
			t.Errorf("%q: missing greeting", m.Subject)
		}
		if strings.Contains(m.PlainText, "<") {
			t.Errorf("%q: HTML in the plain text: %s", m.Subject, m.PlainText)
		}
		if strings.Contains(m.HTML, "&lt;p") {
			t.Errorf("%q: layout escaped the body", m.Subject)
		}
		for _, part := range []string{m.PlainText, m.HTML} {
			if strings.Contains(part, "{{") {
				t.Errorf("%q: unrendered placeholder", m.Subject)
			}
		}
	}
}

func TestTemplatesEscapeHTML(t *testing.T) {
	school, name := `<b>Evil</b> & Co`, `<script>x</script>`
	for _, m := range []Message{
		RegistrationReceivedEmail("a@b.c", name, school),
		AcceptedLoginEmail("https://app.example", "a@b.c", name, school, "pw"),
		NewPasswordEmail("https://app.example", "a@b.c", name, school, "pw"),
		PasswordResetEmail("https://app.example", "a@b.c", name, "t"),
	} {
		if strings.Contains(m.HTML, "<script>") || strings.Contains(m.HTML, "<b>Evil</b>") {
			t.Errorf("%q: unescaped name in HTML", m.Subject)
		}
	}
	if m := RegistrationReceivedEmail("a@b.c", name, school); !strings.Contains(m.PlainText, school) {
		t.Error("plain text should carry the name as typed")
	}
}
