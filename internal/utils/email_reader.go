package utils

import (
	"crypto/tls"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

type EmailReader interface {
	ReadRecentEmailWithSubject(toEmail, subject string) (string, error)
	ReadRecentEmailWithSubjectAndSender(toEmail, subject, sender string) (string, error)
}

type IMAPEmailReader struct {
	config EmailConfig
}

func NewIMAPEmailReader(config EmailConfig) *IMAPEmailReader {
	return &IMAPEmailReader{
		config: config,
	}
}

func (r *IMAPEmailReader) Connect() (*client.Client, error) {
	imapHost := determineIMAPHost(r.config.SMTPHost)
	if imapHost == "" {
		return nil, fmt.Errorf("could not determine IMAP host for %s", r.config.SMTPHost)
	}
	
	imapAddr := fmt.Sprintf("%s:993", imapHost)

	c, err := client.DialTLS(imapAddr, &tls.Config{ServerName: imapHost})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to IMAP server: %v", err)
	}

	if err := c.Login(r.config.SMTPUser, r.config.SMTPPassword); err != nil {
		c.Logout()
		return nil, fmt.Errorf("failed to login to IMAP server: %v", err)
	}

	return c, nil
}

func determineIMAPHost(smtpHost string) string {
	switch {
	case strings.Contains(strings.ToLower(smtpHost), "yandex"):
		return "imap.yandex.ru"
	case strings.Contains(strings.ToLower(smtpHost), "gmail"):
		return "imap.gmail.com"
	case strings.Contains(strings.ToLower(smtpHost), "outlook") || strings.Contains(strings.ToLower(smtpHost), "hotmail"):
		return "outlook.office365.com"
	case strings.Contains(strings.ToLower(smtpHost), "mail.ru"):
		return "imap.mail.ru"
	default:
		return smtpHost
	}
}

func (r *IMAPEmailReader) ReadRecentEmailWithSubject(toEmail, subject string) (string, error) {
	return r.ReadRecentEmailWithSubjectAndSender(toEmail, subject, "")
}

func (r *IMAPEmailReader) ReadRecentEmailWithSubjectAndSender(toEmail, subject, sender string) (string, error) {
	c, err := r.Connect()
	if err != nil {
		return "", err
	}
	defer c.Logout()

	mbox, err := c.Select("INBOX", false)
	if err != nil {
		return "", fmt.Errorf("failed to select INBOX: %v", err)
	}

	if mbox.Messages == 0 {
		return "", fmt.Errorf("no messages in INBOX")
	}

	criteria := imap.NewSearchCriteria()
	criteria.Since = time.Now().Add(-5 * time.Minute)
	
	criteria.Header = make(map[string][]string)
	criteria.Header["SUBJECT"] = []string{subject}
	
	if sender != "" {
		criteria.Header["FROM"] = []string{sender}
	}

	uids, err := c.UidSearch(criteria)
	if err != nil {
		return "", fmt.Errorf("failed to search emails: %v", err)
	}

	if len(uids) == 0 {
		return "", fmt.Errorf("no emails found with subject '%s' in the last 5 minutes", subject)
	}

	latestUID := uids[0]
	for _, uid := range uids {
		if uid > latestUID {
			latestUID = uid
		}
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(latestUID)

	bodySection := &imap.BodySectionName{}

	messages := make(chan *imap.Message, 1)
	done := make(chan error, 1)

	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{bodySection.FetchItem()}, messages)
	}()

	msg := <-messages
	if msg == nil {
		return "", fmt.Errorf("failed to fetch email")
	}

	if err := <-done; err != nil {
		return "", fmt.Errorf("fetch error: %v", err)
	}

	msgBody := msg.GetBody(bodySection)
	if msgBody == nil {
		return "", fmt.Errorf("could not retrieve message body")
	}

	body, err := io.ReadAll(msgBody)
	if err != nil {
		return "", fmt.Errorf("failed to read message body: %v", err)
	}

	emailBody := string(body)
	if emailBody == "" {
		return "", fmt.Errorf("could not find plain text body in email")
	}

	return emailBody, nil
}

func (r *IMAPEmailReader) Extract2FACode(emailBody string) (string, error) {
	re := regexp.MustCompile(`\b\d{6}\b`)
	matches := re.FindAllString(emailBody, -1)
	
	if len(matches) == 0 {
		return "", fmt.Errorf("no 6-digit code found in email body")
	}

	return matches[len(matches)-1], nil
}