package utils

import (
	"errors"
	"fmt"
	"net/smtp"
	"os"
	"strings"
	"time"
)

var ErrMissingSMTPConfig = errors.New("smtp configuration incomplete")

// smtpSendMail is assignable for tests to avoid real SMTP calls.
var smtpSendMail = smtp.SendMail

// SetSendMail overrides the SMTP send function (intended for tests).
func SetSendMail(fn func(addr string, a smtp.Auth, from string, to []string, msg []byte) error) {
	if fn == nil {
		smtpSendMail = smtp.SendMail
		return
	}
	smtpSendMail = fn
}

func SendVerificationEmail(toEmail, code string, expiresIn time.Duration) error {
	subject := "Verify your bblog account"
	body := fmt.Sprintf(
		"Hello,\n\nHere is your bblog verification code: %s\nThis code expires in %d minutes.\n\nIf you did not create this account, you can ignore this email.\n",
		code,
		int64(expiresIn.Minutes()),
	)
	return sendEmail(toEmail, subject, body)
}

func SendPasswordResetEmail(toEmail, resetURL string) error {
	subject := "Reset your bblog password"
	body := fmt.Sprintf("Hello,\n\nA password reset was requested for your bblog account. If this was you, click the link below to set a new password.\n\n%s\n\nIf you didn't request this, you can ignore the email.\n", resetURL)
	return sendEmail(toEmail, subject, body)
}

func sendEmail(toEmail, subject, body string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	username := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASS")
	from := os.Getenv("MAIL_FROM")

	if host == "" || port == "" || username == "" || password == "" || from == "" {
		return ErrMissingSMTPConfig
	}

	addr := fmt.Sprintf("%s:%s", host, port)
	auth := smtp.PlainAuth("", username, password, host)

	sanitizedBody := strings.ReplaceAll(body, "\r", "")
	sanitizedSubject := strings.ReplaceAll(subject, "\r", "")

	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"utf-8\"\r\n\r\n%s", from, toEmail, sanitizedSubject, sanitizedBody))

	return smtpSendMail(addr, auth, from, []string{toEmail}, msg)
}
