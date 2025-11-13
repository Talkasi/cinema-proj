package utils

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
)

type EmailSender interface {
	SendEmail(to, subject, body string) error
}

type SMTPSender struct {
	config EmailConfig
}

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	FromEmail    string
}

func NewSMTPSender(config EmailConfig) *SMTPSender {
	return &SMTPSender{
		config: config,
	}
}
func (s *SMTPSender) SendEmail(to, subject, body string) error {
	auth := smtp.PlainAuth("", s.config.SMTPUser, s.config.SMTPPassword, s.config.SMTPHost)
	to = s.config.SMTPUser

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s\r\n",
		s.config.FromEmail, to, subject, body)

	switch s.config.SMTPPort {
	case "465":
		return s.sendSMTPS(auth, to, msg)
	case "587", "25":
		return s.sendSTARTTLS(auth, to, msg)
	default:
		return fmt.Errorf("unsupported SMTP port: %s", s.config.SMTPPort)
	}
}

func (s *SMTPSender) sendSMTPS(auth smtp.Auth, to, msg string) error {
	addr := fmt.Sprintf("%s:465", s.config.SMTPHost)

	conn, err := tls.Dial("tcp", addr, &tls.Config{
		ServerName: s.config.SMTPHost,
	})
	if err != nil {
		return fmt.Errorf("tls dial failed: %v", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.config.SMTPHost)
	if err != nil {
		return fmt.Errorf("smtp client failed: %v", err)
	}
	defer client.Close()

	return s.authenticateAndSend(client, auth, to, msg)
}

func (s *SMTPSender) sendSTARTTLS(auth smtp.Auth, to, msg string) error {
	addr := fmt.Sprintf("%s:%s", s.config.SMTPHost, s.config.SMTPPort)

	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("dial failed: %v", err)
	}
	defer client.Close()

	if err = client.StartTLS(&tls.Config{
		ServerName: s.config.SMTPHost,
	}); err != nil {
		return fmt.Errorf("starttls failed: %v", err)
	}

	return s.authenticateAndSend(client, auth, to, msg)
}

func (s *SMTPSender) authenticateAndSend(client *smtp.Client, auth smtp.Auth, to, msg string) error {
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("auth failed: %v", err)
	}

	if err := client.Mail(s.config.FromEmail); err != nil {
		return fmt.Errorf("mail failed: %v", err)
	}

	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("rcpt failed: %v", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("data failed: %v", err)
	}

	if _, err := w.Write([]byte(msg)); err != nil {
		return fmt.Errorf("write failed: %v", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("close failed: %v", err)
	}

	return client.Quit()
}

func GetEmailConfigFromEnv() EmailConfig {
	return EmailConfig{
		SMTPHost:     GetEnv("SMTP_HOST"),
		SMTPPort:     GetEnv("SMTP_PORT"),
		SMTPUser:     GetEnv("SMTP_USER"),
		SMTPPassword: GetEnv("SMTP_PASS"),
		FromEmail:    GetEnv("FROM_EMAIL"),
	}
}
