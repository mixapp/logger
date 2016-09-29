package logger

import (
	"github.com/IntelliQru/config"
	"github.com/IntelliQru/mail"
)

const PROVIDER_EMAIL = "email"

type emailProvider struct {
}

func (p emailProvider) GetID() string {
	return PROVIDER_EMAIL
}

func (p emailProvider) Log(msg []byte) {
	send("Log message", msg)
}

func (p emailProvider) Error(msg []byte) {
	send("Error message", msg)
}

func (p emailProvider) Fatal(msg []byte) {
	send("Fatal message", msg)
}

func (p emailProvider) Debug(msg []byte) {
	send("Debug message", msg)
}

func send(subject string, body []byte) {

	to := config.CFG.Str(config.CFG_ADMIN_EMAIL)

	message := mail.NewMail(to, subject, string(body))
	message.BodyContentType = "text/plain"
	message.SendMail()
}
