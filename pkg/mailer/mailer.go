package mailer

import (
	"fmt"
	"net/smtp"
)

const (
	GmailSMTPHost = "smtp.gmail.com"
	GmailSMTPPort = 587 
)

type GmailSMTPMailer struct {
	From string
	Auth smtp.Auth
}

type Mailer interface {
	// Отправляет сообщение одному получателю
	SendMail(to string, subject string, message string) error
}

func (m *GmailSMTPMailer) SendMail(to string, subject string, message string) error {
	msg := fmt.Sprintf(
		"To: %s\r\nSubject: %s\r\n\r\n%s\r\n",
		to,
		subject,
		message,
	)
	return smtp.SendMail(
		fmt.Sprintf("%s:%d", GmailSMTPHost, GmailSMTPPort),
		m.Auth,
		m.From,
		[]string{to},
		[]byte(msg),
	)
}

// Создает экземпляет Mailer работающий через smtp.gmail.com
func NewMailer(from string, pass string) Mailer {
	return &GmailSMTPMailer{
		From: from,
		Auth: smtp.PlainAuth("", from, pass, GmailSMTPHost),
	}
}
