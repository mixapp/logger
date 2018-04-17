package logger

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"time"

	"golang.org/x/net/proxy"
)

const PROVIDER_TELEGRAM = "telegram"

type TelegramProvider struct {
	buf        *bytes.Buffer
	req        *http.Request
	httpClient *http.Client
	chatIds    []string
	mu         sync.Mutex
	once       sync.Once
}

func NewTelegramProvider(conn string, chatIds []string) (*TelegramProvider, error) {

	if len(conn) == 0 {
		return nil, errors.New("Empty telegram connection string")
	} else if len(chatIds) == 0 {
		return nil, errors.New("Empty telegram chat ids")
	}

	var (
		botUrl     string
		httpClient = &http.Client{
			Timeout: time.Second * 10,
		}
	)

	connParts := strings.Split(conn, `|`)

	switch len(connParts) {
	case 1:
		// example: 'https://api.telegram.org/bot<token>/sendMessage'

		if strings.HasPrefix(conn, "https://api.telegram.org/bot") {
			botUrl = connParts[0] // OLD API
		} else {
			botUrl = fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", connParts[0])
		}

	case 2:
		// example: '<token>|<schema>://user:password@ip:port'
		botToken := connParts[0]
		proxyUrl := connParts[1]

		botUrl = fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

		transport, err := httpTransport(proxyUrl)
		if err != nil {
			return nil, err
		}

		httpClient.Transport = transport

	default:
		return nil, errors.New("Invalid connection string format")
	}

	req, err := http.NewRequest(http.MethodPost, botUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	provider := &TelegramProvider{
		req:        req,
		chatIds:    chatIds,
		httpClient: httpClient,
		buf:        &bytes.Buffer{},
	}

	provider.runFlushing()

	return provider, nil
}

func (p *TelegramProvider) GetID() string {
	return PROVIDER_TELEGRAM
}

var (
	_MESSAGE_PREFIX = []byte("=== ")
	_MESSAGE_SUFFIX = []byte(" ===\n")
)

func (p *TelegramProvider) Write(data []byte) (n int, err error) {

	if len(data) == 0 {
		return 0, nil
	}

	p.mu.Lock()
	p.buf.Write(_MESSAGE_PREFIX)
	p.buf.WriteString(time.Now().In(time.UTC).Format("2006-01-02 15:04:05.0000000 -07:00"))
	p.buf.Write(_MESSAGE_SUFFIX)
	p.buf.Write(data)
	p.mu.Unlock()

	return len(data), nil
}

func (p *TelegramProvider) send() (n int, err error) {

	p.mu.Lock()
	defer p.mu.Unlock()

	dataLen := p.buf.Len()
	if dataLen == 0 {
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

	text := p.buf.String()

	for _, chatId := range p.chatIds {
		jd, err := json.Marshal(map[string]string{
			"chat_id": chatId,
			"text":    text,
		})

		if err != nil {
			return 0, err
		}

		body := bytes.NewReader(jd)

		p.req.Body = ioutil.NopCloser(body)
		p.req.ContentLength = int64(len(jd))
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

	p.buf.Reset()
	return dataLen, nil
}

func (p *TelegramProvider) runFlushing() {
	go func() {
		for {
			time.Sleep(time.Second)
			p.send()
		}
	}()
}

func httpTransport(uri string) (*http.Transport, error) {

	proxyUrl, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	var transport *http.Transport

	switch {
	case strings.HasPrefix(proxyUrl.Scheme, "http"):

		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
		}

		transport = &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSClientConfig:       tlsConfig,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 2 * time.Second,
		}

	case proxyUrl.Scheme == "socks5":

		var auth *proxy.Auth
		if proxyUrl.User != nil {
			pass, _ := proxyUrl.User.Password()
			auth = &proxy.Auth{
				User:     proxyUrl.User.Username(),
				Password: pass,
			}
		}

		dialSocksProxy, err := proxy.SOCKS5("tcp", proxyUrl.Host, auth, proxy.Direct)
		if err != nil {
			return nil, err
		}

		transport = &http.Transport{
			Dial: dialSocksProxy.Dial,
		}
	}

	if transport == nil {
		return nil, errors.New("Invalid proxy schema")
	}

	return transport, nil
}
