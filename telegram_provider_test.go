package logger

import (
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
		require.Equal(t,
			&TelegramProvider{
				url:     "https://api.telegram.org/bot<id>/",
				chatIds: []string{"1"},
				httplient: &http.Client{
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

		require.Equal(t, "https://api.telegram.org/bot<bot_id>/sendMessage", p.url)
		require.Equal(t, []string{"1"}, p.chatIds)

		transportUrl, err := p.httplient.Transport.(*http.Transport).Proxy(nil)
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

	p.debugMode = true

	message := []byte("test")

	require.NoError(t, p.Debug(message))
	require.NoError(t, p.Error(message))
	require.NoError(t, p.Log(message))
	require.NoError(t, p.Fatal(message))
}

func TestTelegramSendMessageSocks5(t *testing.T) {

	botId := getBotId()
	proxyUrl := "socks5://" + getProxyUrl() + ":1080"
	chatId := getChatId()

	p, err := NewTelegramProvider(botId+"|"+proxyUrl, []string{chatId})
	require.NoError(t, err)

	p.debugMode = true

	message := []byte("test")

	require.NoError(t, p.Debug(message))
	require.NoError(t, p.Error(message))
	require.NoError(t, p.Log(message))
	require.NoError(t, p.Fatal(message))
}

func getBotId() string {
	return os.Getenv("TELEGRAM_BOT_TOKEN")
}

func getProxyUrl() string {
	return os.Getenv("PROXY_SETTINGS") // example: '<user>:<port>@<ip>:<port>'
}

func getChatId() string {
	return os.Getenv("TELEGRAM_CHAT_ID") // chat number
}
