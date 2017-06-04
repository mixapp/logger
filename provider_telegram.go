package logger

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
)

const PROVIDER_TELEGRAM = "telegram"

type TelegramProvider struct {
	ProviderInterface

	req        *http.Request
	httpClient *http.Client
	chatIds    []string
	mu         sync.Mutex
}

func NewTelegramProvider(token string, chatIds []string) (*TelegramProvider, error) {

	if len(token) == 0 {
		return nil, errors.New("Telegram token is empty.")
	}
	if chatIds == nil {
		return nil, errors.New("Empty telegram chat ids.")
	}

	var url string
	if strings.HasPrefix(token, "https://api.telegram.org/bot") && strings.HasSuffix(token, "sendMessage") {
		url = token // OLD API
	} else {
		url = fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	}

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	provider := &TelegramProvider{
		req:        req,
		httpClient: new(http.Client),
		chatIds:    chatIds,
	}

	return provider, nil
}

func (p *TelegramProvider) GetID() string {
	return PROVIDER_TELEGRAM
}

func (p *TelegramProvider) Write(data []byte) (n int, err error) {

	if len(data) == 0 {
		return 0, nil
	}

	// Telegram response example:
	// {
	//      "ok":false,
	//      "error_code":404,
	//      "description":"Not Found: method not found"
	//      "result":{"id":290082045,....}
	// }

	type Response struct {
		Ok          bool        `json:"ok"`
		ErrorCode   int32       `json:"error_code,omitempty"`
		Description string      `json:"description,omitempty"`
		Parameters  interface{} `json:"parameters,omitempty"`
		Result      interface{} `json:"result,omitempty"`
	}

	jd, err := json.Marshal(string(data))
	if err != nil {
		return 0, err
	}

	buf := bytes.NewBuffer(make([]byte, 0, len(jd)+100))

	p.mu.Lock()
	defer p.mu.Unlock()

	for _, chatId := range p.chatIds {
		buf.Reset()
		buf.WriteByte('{')
		buf.WriteString(`"chat_id":`)
		buf.WriteString(chatId)
		buf.WriteByte(',')
		buf.WriteString(`"text":`)
		buf.Write(jd)
		buf.WriteByte('}')

		body := bytes.NewReader(buf.Bytes())

		p.req.Body = ioutil.NopCloser(body)
		p.req.ContentLength = int64(buf.Len())
		p.req.GetBody = func() (io.ReadCloser, error) {
			return ioutil.NopCloser(body), nil
		}

		resp, err := p.httpClient.Do(p.req)
		if err != nil {
			return 0, err
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			continue
		}

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return 0, err
		}

		if len(respBody) == 0 {
			return 0, fmt.Errorf("Failed send message to telegram: %+v", resp.Header)
		} else {
			r := new(Response)
			if err := json.Unmarshal(respBody, r); err != nil {
				r.Description = err.Error()
			}

			return 0, fmt.Errorf("Failed send message to telegram: %+v", r)
		}
	}

	return len(data), nil
}
