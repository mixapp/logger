package logger

import (
	"bytes"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTelegramProvider(t *testing.T) {

	{
		p, err := NewTelegramProvider("", []string{})
		require.EqualError(t, err, "Empty telegram connection string")
		require.Nil(t, p)
	}

	{
		p, err := NewTelegramProvider("-", []string{})
		require.EqualError(t, err, "Empty telegram chat ids")
		require.Nil(t, p)
	}

	{
		// old connection string
		p, err := NewTelegramProvider("https://api.telegram.org/bot<id>/", []string{"1"})
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "https://api.telegram.org/bot<id>/", nil)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		require.Equal(t,
			&TelegramProvider{
				req:     req,
				buf:     &bytes.Buffer{},
				chatIds: []string{"1"},
				httpClient: &http.Client{
					Timeout: time.Second * 10,
				},
			},
			p)
	}

	{
		// new connection string with proxy
		proxyUrl, err := url.Parse("https://user:password@ip:port")
		require.NoError(t, err)

		p, err := NewTelegramProvider("<bot_id>|"+proxyUrl.String(), []string{"1"})
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "https://api.telegram.org/bot<bot_id>/sendMessage", nil)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		require.Equal(t, req, p.req)
		require.Equal(t, []string{"1"}, p.chatIds)

		transportUrl, err := p.httpClient.Transport.(*http.Transport).Proxy(nil)
		require.NoError(t, err)
		require.Equal(t, "https://user:password@ip:port", transportUrl.String())
	}
}

func TestTelegramSendMessageHttpProxy(t *testing.T) {

	botId := getBotId()
	proxyUrl := "http://" + getProxyUrl() + ":8080"
	chatId := getChatId()

	p, err := NewTelegramProvider(botId+"|"+proxyUrl, []string{chatId})
	require.NoError(t, err)

	n, err := p.Write([]byte("test http proxy"))
	require.NoError(t, err)
	require.Equal(t, 15, n)

	n, err = p.send()
	require.NoError(t, err)
	require.Equal(t, 58, n)
}

func TestTelegramSendMessageSocks5(t *testing.T) {

	botId := getBotId()
	proxyUrl := "socks5://" + getProxyUrl() + ":1080"
	chatId := getChatId()

	p, err := NewTelegramProvider(botId+"|"+proxyUrl, []string{chatId})
	require.NoError(t, err)

	n, err := p.Write([]byte("test socks5 proxy"))
	require.NoError(t, err)
	require.Equal(t, 17, n)

	n, err = p.send()
	require.NoError(t, err)
	require.Equal(t, 60, n)
}

func getBotId() string {
	return os.Getenv("TELEGRAM_BOT_TOKEN")
}

func getProxyUrl() string {
	return os.Getenv("TELEGRAM_PROXY_SETTINGS") // example: '<user>:<port>@<ip>'
}

func getChatId() string {
	return os.Getenv("TELEGRAM_CHAT_ID") // chat number
}
