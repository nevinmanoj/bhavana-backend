package mail

import (
	"bytes"
	"embed"
	htmltemplate "html/template"
	"net/url"
	"strings"
	texttemplate "text/template"
)

// The wording of every email lives in templates/. Each email is a .html and a
// .txt file with the same name; a change to one usually belongs in the other.
// layout.html wraps every HTML email, and login_details is shared by the two
// emails that carry a password.
//
// The files are compiled into the binary, because the runtime image holds only
// the binary. html/template escapes values for us; the .txt files go through
// text/template so a name like St Mary's stays as typed.
//
//go:embed templates
var templateFiles embed.FS

var (
	htmlTemplates = htmltemplate.Must(htmltemplate.ParseFS(templateFiles, "templates/*.html"))
	textTemplates = texttemplate.Must(texttemplate.ParseFS(templateFiles, "templates/*.txt"))
)

// emailData is what the templates can read. Each email fills in what it needs.
type emailData struct {
	Name       string
	Email      string
	SchoolName string
	Password   string
	LoginURL   string
	ResetLink  string
}

func RegistrationReceivedEmail(email, name, schoolName string) Message {
	return message("registration_received", "Bhavana: we received your school registration", emailData{
		Name:       name,
		Email:      email,
		SchoolName: schoolName,
	})
}

func AcceptedLoginEmail(appURL, email, name, schoolName, password string) Message {
	return message("accepted_login", "Bhavana: your school registration is accepted", emailData{
		Name:       name,
		Email:      email,
		SchoolName: schoolName,
		Password:   password,
		LoginURL:   baseURL(appURL) + "/login",
	})
}

func NewPasswordEmail(appURL, email, name, schoolName, password string) Message {
	return message("new_password", "Bhavana: your new sign-in password", emailData{
		Name:       name,
		Email:      email,
		SchoolName: schoolName,
		Password:   password,
		LoginURL:   baseURL(appURL) + "/login",
	})
}

func PasswordResetEmail(appURL, email, name, token string) Message {
	return message("password_reset", "Bhavana: reset your password", emailData{
		Name:      name,
		Email:     email,
		ResetLink: baseURL(appURL) + "/reset-password?token=" + url.QueryEscape(token),
	})
}

// message renders one email. Rendering can only fail on a mistake in a template
// file, such as a misspelled field, so it panics rather than making every
// caller handle an error. The tests render every email and catch those.
func message(template, subject string, data emailData) Message {
	var text, body, page bytes.Buffer
	must(textTemplates.ExecuteTemplate(&text, template+".txt", data))
	must(htmlTemplates.ExecuteTemplate(&body, template+".html", data))
	// body is already escaped, so layout.html must not escape it again.
	must(htmlTemplates.ExecuteTemplate(&page, "layout.html", htmltemplate.HTML(body.String())))

	return Message{
		To:        data.Email,
		ToName:    data.Name,
		Subject:   subject,
		PlainText: text.String(),
		HTML:      page.String(),
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func baseURL(appURL string) string { return strings.TrimRight(appURL, "/") }
