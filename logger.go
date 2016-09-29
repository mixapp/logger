package logger

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

var LOGGER *Logger

const (
	LEVEL_ERROR int = iota
	LEVEL_INFO
	LEVEL_DEBUG
)

type ProviderInterface interface {
	GetID() string
	Log(msg []byte)
	Error(msg []byte)
	Fatal(msg []byte)
	Debug(msg []byte)
}

func init() {
	LOGGER = NewLogger()

	cp := consoleProvider{}
	ep := emailProvider{}

	LOGGER.RegisterProvider(cp)
	LOGGER.RegisterProvider(ep)

	LOGGER.AddLogProvider(PROVIDER_CONSOLE)
	LOGGER.AddErrorProvider(PROVIDER_CONSOLE, PROVIDER_EMAIL)
	LOGGER.AddFatalProvider(PROVIDER_CONSOLE, PROVIDER_EMAIL)
	LOGGER.AddDebugProvider(PROVIDER_CONSOLE)
}

type Logger struct {
	providers      map[string]*ProviderInterface
	logProviders   []string
	errorProviders []string
	fatalProviders []string
	debugProviders []string
	level          int
}

func NewLogger() *Logger {
	newLogger := Logger{
		providers: make(map[string]*ProviderInterface, 0),
	}

	return &newLogger
}

func (l *Logger) RegisterProvider(p ProviderInterface) {
	l.providers[p.GetID()] = &p
}

func (l *Logger) AddLogProvider(provIDs ...string) {

	for _, provID := range provIDs {
		p, bFound := l.providers[provID]

		if bFound {
			pID := (*p).GetID()

			for _, val := range l.logProviders {
				if val == pID {
					return
				}
			}

			l.logProviders = append(l.logProviders, pID)
		}
	}
}

func (l *Logger) AddErrorProvider(provIDs ...string) {

	for _, provID := range provIDs {

		p, bFound := l.providers[provID]

		if bFound {
			pID := (*p).GetID()

			for _, val := range l.errorProviders {
				if val == pID {
					return
				}
			}

			l.errorProviders = append(l.errorProviders, pID)
		}
	}
}

func (l *Logger) AddFatalProvider(provIDs ...string) {

	for _, provID := range provIDs {

		p, bFound := l.providers[provID]

		if bFound {
			pID := (*p).GetID()

			for _, val := range l.fatalProviders {
				if val == pID {
					return
				}
			}

			l.fatalProviders = append(l.fatalProviders, pID)
		}
	}
}

func (l *Logger) AddDebugProvider(provIDs ...string) {

	for _, provID := range provIDs {

		p, bFound := l.providers[provID]

		if bFound {
			pID := (*p).GetID()

			for _, val := range l.debugProviders {
				if val == pID {
					return
				}
			}

			l.debugProviders = append(l.debugProviders, pID)
		}
	}
}

func (l *Logger) SetLevel(level int) {
	l.level = level
}

var (
	HOST              string
	MESSAGE_REPLACER  = strings.NewReplacer("\r", "", "\n", "\t")
	MESSAGE_SEPARATOR = []byte("\t")
)

func (l *Logger) makeMessage(typeLog string, err []interface{}) *bytes.Buffer {

	if len(HOST) == 0 {
		HOST, _ = os.Hostname()
	}

	buf := bytes.NewBuffer(nil)
	prefix := fmt.Sprintf("%s: %s %s ", typeLog, time.Now().Format(time.RFC3339), HOST)
	logger := log.New(buf, prefix, log.Lshortfile)

	msg := bytes.NewBuffer(nil)
	for _, v := range err {
		if msg.Len() > 0 {
			msg.Write(MESSAGE_SEPARATOR)
		}
		fmt.Fprintf(msg, "%v", v)
	}

	logger.Output(3, MESSAGE_REPLACER.Replace(msg.String()))

	return buf
}

func (l *Logger) Log(err ...interface{}) {
	if l.level < LEVEL_INFO {
		return
	}
	msg := l.makeMessage("LOG", err).Bytes()
	for _, pID := range l.logProviders {
		p, bFound := l.providers[pID]
		if bFound {
			(*p).Log(msg)
		}
	}
}

func (l *Logger) Error(err ...interface{}) {

	msg := l.makeMessage("ERROR", err).Bytes()
	for _, pID := range l.errorProviders {
		p, bFound := l.providers[pID]
		if bFound {
			(*p).Error(msg)
		}
	}
}

func (l *Logger) Debug(err ...interface{}) {
	if l.level < LEVEL_DEBUG {
		return
	}
	msg := l.makeMessage("DEBUG", err).Bytes()
	for _, pID := range l.debugProviders {
		p, bFound := l.providers[pID]
		if bFound {
			(*p).Debug(msg)
		}
	}
}

func (l *Logger) Fatal(err ...interface{}) {
	msg := l.makeMessage("FATAL", err).Bytes()
	for _, pID := range l.fatalProviders {
		p, bFound := l.providers[pID]
		if bFound {
			(*p).Fatal(msg)
		}
	}

	os.Exit(1)
}
