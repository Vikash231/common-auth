package email

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strconv"
	"strings"
)

type Mailer struct {
	host     string
	port     int
	user     string
	password string
	from     string
}

var Default *Mailer

func Init() {
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if port == 0 {
		port = 587
	}
	Default = &Mailer{
		host:     os.Getenv("SMTP_HOST"),
		port:     port,
		user:     os.Getenv("SMTP_USER"),
		password: os.Getenv("SMTP_PASSWORD"),
		from:     os.Getenv("SMTP_FROM"),
	}
	if Default.from == "" {
		Default.from = "noreply@common-auth.local"
	}
}

func (m *Mailer) Send(to, subject, body string) error {
	if m.host == "" {
		// Log the email to console when SMTP is not configured
		log.Printf("\n📧 ─── EMAIL ───────────────────────\nTO: %s\nSUBJECT: %s\n\n%s\n────────────────────────────────────\n",
			to, subject, stripHTML(body))
		return nil
	}

	addr := fmt.Sprintf("%s:%d", m.host, m.port)
	var auth smtp.Auth
	if m.user != "" && m.password != "" {
		auth = smtp.PlainAuth("", m.user, m.password, m.host)
	}

	msg := buildMessage(m.from, to, subject, body)
	return smtp.SendMail(addr, auth, m.from, []string{to}, []byte(msg))
}

func buildMessage(from, to, subject, body string) string {
	return fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		from, to, subject, body)
}

// stripHTML removes HTML tags for console output
func stripHTML(s string) string {
	result := strings.ReplaceAll(s, "<br>", "\n")
	result = strings.ReplaceAll(result, "<br/>", "\n")
	result = strings.ReplaceAll(result, "<p>", "\n")
	result = strings.ReplaceAll(result, "</p>", "")
	result = strings.ReplaceAll(result, "<h2>", "\n")
	result = strings.ReplaceAll(result, "</h2>", "\n")
	// Simple tag stripper
	inTag := false
	var b strings.Builder
	for _, ch := range result {
		if ch == '<' {
			inTag = true
			continue
		}
		if ch == '>' {
			inTag = false
			continue
		}
		if !inTag {
			b.WriteRune(ch)
		}
	}
	return strings.TrimSpace(b.String())
}

func SendVerificationEmail(to, name, token string) error {
	baseURL := os.Getenv("APP_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:5173"
	}
	link := fmt.Sprintf("%s/set-password?token=%s", baseURL, token)
	subject := "Verify your email and set your password"
	body := fmt.Sprintf(`
<h2>Welcome to Common Auth, %s!</h2>
<p>Click the link below to verify your email and create your password:</p>
<p><a href="%s">Set Password: %s</a></p>
<p>This link expires in 24 hours.</p>
`, name, link, link)
	return Default.Send(to, subject, body)
}

func SendInviteEmail(to, adminName, token string) error {
	baseURL := os.Getenv("APP_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:5173"
	}
	link := fmt.Sprintf("%s/set-password?token=%s", baseURL, token)
	subject := "You've been invited to Common Auth"
	body := fmt.Sprintf(`
<h2>Hello!</h2>
<p>%s has invited you to Common Auth.</p>
<p>Click the link below to set your password and activate your account:</p>
<p><a href="%s">Set Password: %s</a></p>
<p>This link expires in 72 hours.</p>
`, adminName, link, link)
	return Default.Send(to, subject, body)
}
