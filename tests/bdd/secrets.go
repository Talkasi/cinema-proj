package bdd

import (
	cwUtils "cw/internal/utils"
)

type TestSecrets struct {
	TestDBURL    string
	SMTPHost     string
	IMAPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	FromEmail    string
	JWTSecret    string
}

func GetTestSecretsFromEnv() TestSecrets {
	return TestSecrets{
		TestDBURL:    cwUtils.GetEnv("TEST_DATABASE_URL"),
		SMTPHost:     cwUtils.GetEnv("SMTP_HOST"),
		IMAPHost:     cwUtils.GetEnv("IMAP_HOST"),
		SMTPPort:     cwUtils.GetEnv("SMTP_PORT"),
		SMTPUser:     cwUtils.GetEnv("SMTP_USER"),
		SMTPPassword: cwUtils.GetEnv("SMTP_PASS"),
		FromEmail:    cwUtils.GetEnv("FROM_EMAIL"),
		JWTSecret:    cwUtils.GetEnv("JWT_SECRET"),
	}
}
