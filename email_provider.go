package logger

import (
	"errors"
	"github.com/IntelliQru/mail"
)

const PROVIDER_EMAIL = "email"

type EmailProvider struct {
	address    string
	smtpClient *mail.SmtpClient
}

func NewEmailProvider(address string, smtpClient *mail.SmtpClient) (*EmailProvider, error) {

	if len(address) == 0 {
		return nil, errors.New("Empty email address.")
	} else if smtpClient == nil {
		return nil, errors.New("Empty smtp client.")
	}

	provider := &EmailProvider{
		address:    address,
		smtpClient: smtpClient,
	}

	return provider, nil
}

func (p EmailProvider) GetID() string {
	return PROVIDER_EMAIL
}

func (p EmailProvider) Log(msg []byte) {
	p.send("Log message", msg)
}

func (p EmailProvider) Error(msg []byte) {
	p.send("Error message", msg)
}

func (p EmailProvider) Fatal(msg []byte) {
	p.send("Fatal message", msg)
}

func (p EmailProvider) Debug(msg []byte) {
	p.send("Debug message", msg)
}

func (p EmailProvider) send(subject string, body []byte) {
	message := mail.NewMessage(p.smtpClient, p.address, subject, string(body))
	message.BodyContentType = "text/plain"
	go message.SendMail()
}
