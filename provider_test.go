package logger

import (
	"bytes"
	"testing"
	"time"

	"github.com/IntelliQru/mail"
)

func TestRemoveNewLinesInText(t *testing.T) {

	for _, td := range []struct {
		Src []byte
		Res []byte
	}{
		{[]byte("He\r\nll\ro, \n世\r界\n"), []byte("He  ll o,  世 界\n")},
		{[]byte("При\r\nве\rт, \n世\r界\n"), []byte("При  ве т,  世 界\n")},
	} {

		res := make([]byte, len(td.Src))
		copy(res, td.Src)

		removeNewLinesInText(res)
		if !bytes.Equal(res, td.Res) {
			t.Errorf("%v != %v", td.Res, res)
		}
	}

}

func TestEmail(t *testing.T) {

	// You can use that 'https://hub.docker.com/r/velaluqa/iredmail/' mail server for tests
	cl := mail.SmtpClient{
		Host:     "localhost",
		Port:     "587",
		User:     "postmaster@example.org",
		Password: "teivVedJin",
		From:     "postmaster@example.org",
	}

	prv, err := NewEmailProvider("postmaster@example.org", &cl)
	if err != nil {
		t.Fatal(err)
	}

	for _, txt := range []string{
		"",
		"Telegram provider test:" + time.Now().Format(time.RFC3339Nano),
	} {
		n, err := prv.Write([]byte(txt))
		if err != nil {
			t.Error(err)
		} else if n != len([]byte(txt)) {
			t.Error("Fail")
		}
	}
}
