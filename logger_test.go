package logger

import (
	"fmt"
	"regexp"
	"testing"
)

func TestAddProvider(t *testing.T) {

	providerInfo := Provider{Level: LEVEL_INFO}
	providerError := Provider{Level: LEVEL_ERROR}
	providerDebug := Provider{Level: LEVEL_DEBUG}

	l := NewLogger()
	l.RegisterProvider(&providerInfo)
	l.RegisterProvider(&providerError)
	l.RegisterProvider(&providerDebug)

	l.AddErrorProvider(providerError.GetID())
	if len(l.errorProviders) != 1 || l.errorProviders[0] != providerError.GetID() {
		t.Error("Failed registration of the provider 'error'.")
	}

	l.AddLogProvider(providerInfo.GetID(), providerInfo.GetID())
	if len(l.logProviders) != 1 || l.logProviders[0] != providerInfo.GetID() {
		t.Error("Failed registration of the provider 'log'.")
	}

	l.AddFatalProvider(providerInfo.GetID(), providerError.GetID())
	if len(l.fatalProviders) != 2 || l.fatalProviders[0] != providerInfo.GetID() || l.fatalProviders[1] != providerError.GetID() {
		t.Error("Failed registration of the provider 'fatal'.")
	}

	l.AddDebugProvider(providerDebug.GetID(), providerDebug.GetID(), providerError.GetID())
	if len(l.debugProviders) != 2 || l.debugProviders[0] != providerDebug.GetID() || l.debugProviders[1] != providerError.GetID() {
		t.Error("Failed registration of the provider 'debug'.")
	}
}

func TestMessage(t *testing.T) {

	// example:
	// Err: 2016-09-29T14:32:49+03:00 MyHost testing.go:610: text1 text2 text3text4
	// Info: 2016-09-29T14:32:49+03:00 MyHost testing.go:610: text1 text2 text3text4
	expr := regexp.MustCompile(`: \d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[+-]\d{2}:\d{2} [^ ]+ [^:]+:\d+:`)

	testData := []interface{}{"line1\nline2", "line3\r\nline4", "text1", "text2", 10}

	for _, prefix := range []string{"Err", "Info"} {
		msg := string(makeMessage(prefix, testData).Bytes())
		msg = expr.ReplaceAllString(msg, "")

		etalon := prefix + " line1\tline2\tline3\tline4\ttext1\ttext2\t10\n"
		if msg != etalon {
			t.Errorf("Failed massage: '%s' != '%s'", msg, etalon)
		}
	}
}

func TestPringMessage(t *testing.T) {

	fn := func() {
		if str := recover(); str != nil {
			t.Errorf("%#v", str)
		}
	}

	for _, level := range []int{LEVEL_ERROR, LEVEL_INFO, LEVEL_DEBUG} {
		provider := Provider{Level: level}

		l := NewLogger()
		l.SetLevel(level)
		l.RegisterProvider(&provider)

		l.AddLogProvider(provider.GetID())
		l.AddErrorProvider(provider.GetID())
		l.AddFatalProvider(provider.GetID())
		l.AddDebugProvider(provider.GetID())

		defer fn()
		l.Log("text")

		defer fn()
		l.Debug("text")
	}
}

type Provider struct {
	ProviderInterface
	Level int
}

func (p Provider) GetID() string {
	return fmt.Sprintf("%d", p.Level)
}

func (p *Provider) Log(msg []byte) {
	if p.Level < LEVEL_INFO {
		panic("call LOG")
	}
}

func (p *Provider) Debug(msg []byte) {
	if p.Level < LEVEL_DEBUG {
		panic("call Debug")
	}
}
