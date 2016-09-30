package logger

import (
	"github.com/IntelliQru/config"
	"github.com/IntelliQru/mail"
)

const PROVIDER_EMAIL = "email"

type EmailProvider struct {
}

func (p EmailProvider) GetID() string {
	return PROVIDER_EMAIL
}

func (p EmailProvider) Log(msg []byte) {
	send("Log message", msg)
}

func (p EmailProvider) Error(msg []byte) {
	send("Error message", msg)
}

func (p EmailProvider) Fatal(msg []byte) {
	send("Fatal message", msg)
}

func (p EmailProvider) Debug(msg []byte) {
	send("Debug message", msg)
}

func send(subject string, body []byte) {

	to := config.CFG.Str(config.CFG_ADMIN_EMAIL)

	message := mail.NewMail(to, subject, string(body))
	message.BodyContentType = "text/plain"
	message.SendMail()
}
