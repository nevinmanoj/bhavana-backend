package mail

import (
	"fmt"
	"html"
	"net/url"
	"strings"
)

func RegistrationReceivedEmail(email, name, schoolName string) Message {
	text := fmt.Sprintf(`Hello %s,

We've received the registration for %s. An admin will review it.

Once it's accepted, we'll email your sign-in password to this address. You don't need to do anything now.

If you didn't register this school, you can ignore this email.
`, name, schoolName)

	body := greeting(name) +
		p("We've received the registration for <strong>"+h(schoolName)+"</strong>. An admin will review it.") +
		p("Once it's accepted, we'll email your sign-in password to this address. You don't need to do anything now.") +
		p("If you didn't register this school, you can ignore this email.")

	return Message{
		To:        email,
		ToName:    name,
		Subject:   "Bhavana: we received your school registration",
		PlainText: text,
		HTML:      page(body),
	}
}

func AcceptedLoginEmail(appURL, email, name, schoolName, password string) Message {
	return loginEmail(appURL, email, name, password,
		"Bhavana: your school registration is accepted",
		fmt.Sprintf("The registration for %s has been accepted. You can now sign in.", schoolName),
		"The registration for <strong>"+h(schoolName)+"</strong> has been accepted. You can now sign in.",
	)
}

func NewPasswordEmail(appURL, email, name, schoolName, password string) Message {
	return loginEmail(appURL, email, name, password,
		"Bhavana: your new sign-in password",
		fmt.Sprintf("A new password has been created for %s. Your old password no longer works.", schoolName),
		"A new password has been created for <strong>"+h(schoolName)+"</strong>. Your old password no longer works.",
	)
}

func loginEmail(appURL, email, name, password, subject, introText, introHTML string) Message {
	loginURL := baseURL(appURL) + "/login"
	const keep = `Please keep this email. You can change this password after signing in, using "Change password" at the bottom of the menu. If you forget it, use "Forgot password?" on the sign-in page.`

	text := fmt.Sprintf(`Hello %s,

%s

Sign-in page: %s
Email: %s
Password: %s

%s
`, name, introText, loginURL, email, password, keep)

	body := greeting(name) +
		p(introHTML) +
		p(`Sign-in page: <a href="`+h(loginURL)+`">`+h(loginURL)+`</a><br>Email: `+h(email)) +
		`<p style="margin:0 0 16px">Password:<br><span style="display:inline-block;margin-top:6px;padding:8px 12px;background:#f3f4f6;border-radius:6px;font-family:Consolas,Menlo,monospace;font-size:20px;letter-spacing:1px">` + h(password) + `</span></p>` +
		p(h(keep))

	return Message{To: email, ToName: name, Subject: subject, PlainText: text, HTML: page(body)}
}

func PasswordResetEmail(appURL, email, name, token string) Message {
	link := baseURL(appURL) + "/reset-password?token=" + url.QueryEscape(token)

	text := fmt.Sprintf(`Hello %s,

Someone asked to reset the password for %s. To choose a new password, open this link:

%s

The link works once and expires in 24 hours. If you didn't ask for this, you can ignore this email. Your password hasn't changed.
`, name, email, link)

	body := greeting(name) +
		p("Someone asked to reset the password for "+h(email)+". To choose a new password, open this link:") +
		p(`<a href="`+h(link)+`">Choose a new password</a>`) +
		p("The link works once and expires in 24 hours. If you didn't ask for this, you can ignore this email. Your password hasn't changed.")

	return Message{
		To:        email,
		ToName:    name,
		Subject:   "Bhavana: reset your password",
		PlainText: text,
		HTML:      page(body),
	}
}

func baseURL(appURL string) string { return strings.TrimRight(appURL, "/") }

func h(s string) string { return html.EscapeString(s) }

func p(inner string) string { return `<p style="margin:0 0 16px">` + inner + `</p>` }

func greeting(name string) string { return p("Hello " + h(name) + ",") }

func page(body string) string {
	return `<!DOCTYPE html><html><body style="font-family:Arial,Helvetica,sans-serif;font-size:15px;line-height:1.5;color:#111827">` + body + `</body></html>`
}
