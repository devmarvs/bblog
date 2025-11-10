package utils

import (
	"errors"
	"fmt"
	"net/smtp"
	"os"
	"strings"
)

var ErrMissingSMTPConfig = errors.New("smtp configuration incomplete")

func SendVerificationEmail(toEmail, verifyURL string) error {
	subject := "Verify your bblog account"
	body := fmt.Sprintf("Hello,\n\nPlease verify your bblog account by clicking the link below.\n\n%s\n\nIf you did not create this account, you can ignore this email.\n", verifyURL)
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

	return smtp.SendMail(addr, auth, from, []string{toEmail}, msg)
}
