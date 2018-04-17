package logger

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

const PROVIDER_TELEGRAM = "telegram"

type TelegramProvider struct {
	url       string
	chatIds   []string
	httplient *http.Client
	debugMode bool
}

func NewTelegramProvider(conn string, chatIds []string) (*TelegramProvider, error) {

	if len(conn) == 0 {
		return nil, errors.New("Empty telegram connection string")
	} else if len(chatIds) == 0 {
		return nil, errors.New("Empty telegram chat ids")
	}

	var (
		botUrl    string
		httplient = &http.Client{
			Timeout: time.Second * 10,
		}
	)

	connParts := strings.Split(conn, `|`)

	switch len(connParts) {
	case 1:
		// example: 'https://api.telegram.org/bot<token>/sendMessage'
		botUrl = connParts[0]

	case 2:
		// example: '<token>|<schema>://user:password@ip:port'
		botToken := connParts[0]
		proxyUrl := connParts[1]

		botUrl = fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

		transport, err := httpTransport(proxyUrl)
		if err != nil {
			return nil, err
		}

		httplient.Transport = transport

	default:
		return nil, errors.New("Invalid connection string format")
	}

	provider := &TelegramProvider{
		url:       botUrl,
		chatIds:   chatIds,
		httplient: httplient,
	}

	return provider, nil
}

func (p TelegramProvider) GetID() string {
	return PROVIDER_TELEGRAM
}

func (p TelegramProvider) Log(msg []byte) {
	p.send("INFO:", msg)
}

func (p TelegramProvider) Error(msg []byte) {
	p.send("ERROR:", msg)
}

func (p TelegramProvider) Fatal(msg []byte) {
	p.send("FATAL:", msg)
}

func (p TelegramProvider) Debug(msg []byte) {
	p.send("DEBUG:", msg)
}

func (p *TelegramProvider) send(subject string, body []byte) error {

	buf := bytes.NewBuffer(nil)
	err := json.NewEncoder(buf).Encode(map[string]interface{}{
		"chat_id": "unknown",
		"text":    subject + "\n" + string(body),
	})
	if err != nil {
		log.Println(err)
		return err
	}

	chatIdTemplate := []byte(`"chat_id":"unknown"`)

	for _, chatId := range p.chatIds {
		js := bytes.Replace(buf.Bytes(), chatIdTemplate, []byte(`"chat_id":"`+chatId+`"`), 1)

		req, err := http.NewRequest(http.MethodPost, p.url, bytes.NewBuffer(js))
		if err != nil {
			log.Println(err)
			return err
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := p.httplient.Do(req)
		if err != nil {
			log.Println(err)
			return err
		}

		if p.debugMode {
			type Answer struct {
				Ok bool `json:"ok"`
			}

			answer := &Answer{}
			err := json.NewDecoder(resp.Body).Decode(answer)
			if err != nil {
				return err
			}

			if !answer.Ok {
				return errors.New("Invalid response data")
			}
		} else {
			resp.Body.Close()
		}
	}

	return nil
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
