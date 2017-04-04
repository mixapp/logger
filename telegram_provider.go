package logger

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

const PROVIDER_TELEGRAM = "telegram"

type TelegramProvider struct {
	url     string
	chatIds []string
}

func NewTelegramProvider(url string, chatIds []string) (*TelegramProvider, error) {

	if len(url) == 0 {
		return nil, errors.New("Empty telegram url.")
	} else if chatIds == nil {
		return nil, errors.New("Empty telegram chat ids.")
	}

	provider := &TelegramProvider{
		url:     url,
		chatIds: chatIds,
	}

	return provider, nil
}

func (p TelegramProvider) GetID() string {
	return PROVIDER_TELEGRAM
}

func (p TelegramProvider) Log(msg []byte) {
	p.send("Log message", msg)
}

func (p TelegramProvider) Error(msg []byte) {
	p.send("Error message", msg)
}

func (p TelegramProvider) Fatal(msg []byte) {
	p.send("Fatal message", msg)
}

func (p TelegramProvider) Debug(msg []byte) {
	p.send("Debug message", msg)
}

func (p TelegramProvider) send(subject string, body []byte) {
	go tg_send(p.url, p.chatIds, subject, body)
}

func tg_send(url string, chatIds []string, subject string, body []byte) {
	msg := map[string]interface{}{
		"chat_id": "",
		"text":    subject + "\n" + string(body),
	}
	client := &http.Client{}

	for _, chatId := range chatIds {
		msg["chat_id"] = chatId
		jsonStr, err := json.Marshal(msg)
		if err != nil {
			return
		}

		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return
		}

		defer resp.Body.Close()
	}
}
