package mail

import (
	"strings"
	"testing"
)

func TestLoginEmails(t *testing.T) {
	for _, m := range []Message{
		RegistrationReceivedEmail("https://app.example/", "anu@example.com", "Anu", "St Mary's", "abcd2345efgh6789jkmn"),
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

// A school in review can sign in but only look, so the email has to say so.
func TestRegistrationEmailExplainsWhatWorksDuringReview(t *testing.T) {
	m := RegistrationReceivedEmail("https://app.example", "anu@example.com", "Anu", "St Mary's", "pw")
	for _, part := range []string{m.PlainText, m.HTML} {
		for _, want := range []string{"sign in using the credentials below", "Once event registrations open, you’ll be able to add students"} {
			if !strings.Contains(part, want) {
				t.Errorf("missing %q: %s", want, part)
			}
		}
	}
}

func TestRegistrationReceivedSubjectNamesTheSchool(t *testing.T) {
	m := RegistrationReceivedEmail("https://app.example", "anu@example.com", "Anu", "St Mary's", "pw")
	if want := "RCI Bhavana 2027 | School Registration | St Mary's"; m.Subject != want {
		t.Errorf("subject = %q, want %q", m.Subject, want)
	}
}

func TestRegistrationAcceptedEmail(t *testing.T) {
	m := RegistrationAcceptedEmail("https://app.example/", "anu@example.com", "Anu", "St Mary's")
	for _, part := range []string{m.PlainText, m.HTML} {
		for _, want := range []string{"anu@example.com", "https://app.example/login", "Forgot password?"} {
			if !strings.Contains(part, want) {
				t.Errorf("missing %q", want)
			}
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
		RegistrationReceivedEmail("https://app.example", "anu@example.com", "Anu", "St Mary's", "pw"),
		RegistrationAcceptedEmail("https://app.example", "anu@example.com", "Anu", "St Mary's"),
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
		if !strings.Contains(m.HTML, "Hello") || !strings.HasPrefix(m.PlainText, "Hello") {
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
		RegistrationReceivedEmail("https://app.example", "a@b.c", name, school, "pw"),
		RegistrationAcceptedEmail("https://app.example", "a@b.c", name, school),
		AcceptedLoginEmail("https://app.example", "a@b.c", name, school, "pw"),
		NewPasswordEmail("https://app.example", "a@b.c", name, school, "pw"),
		PasswordResetEmail("https://app.example", "a@b.c", name, "t"),
	} {
		if strings.Contains(m.HTML, "<script>") || strings.Contains(m.HTML, "<b>Evil</b>") {
			t.Errorf("%q: unescaped name in HTML", m.Subject)
		}
	}
	if m := AcceptedLoginEmail("https://app.example", "a@b.c", name, school, "pw"); !strings.Contains(m.PlainText, school) {
		t.Error("plain text should carry the name as typed")
	}
}
