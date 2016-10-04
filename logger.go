package logger

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

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

type Logger struct {
	providers      map[string]*ProviderInterface
	logProviders   []string
	errorProviders []string
	fatalProviders []string
	debugProviders []string
	level          int
}

func NewLogger() *Logger {
	return &Logger{
		providers: make(map[string]*ProviderInterface, 0),
	}
}

func (l *Logger) SetLevel(level int) {
	l.level = level
}

func (l *Logger) RegisterProvider(p ProviderInterface) {
	l.providers[p.GetID()] = &p
}

func (l *Logger) AddLogProvider(provIDs ...string) {
	l.addProvider("log", provIDs...)
}

func (l *Logger) AddErrorProvider(provIDs ...string) {
	l.addProvider("error", provIDs...)
}

func (l *Logger) AddFatalProvider(provIDs ...string) {
	l.addProvider("fatal", provIDs...)
}

func (l *Logger) AddDebugProvider(provIDs ...string) {
	l.addProvider("debug", provIDs...)
}

func (l *Logger) addProvider(providerType string, providersIDs ...string) {

	var IDs *[]string
	switch providerType {
	case "debug":
		IDs = &l.debugProviders
	case "log":
		IDs = &l.logProviders
	case "error":
		IDs = &l.errorProviders
	case "fatal":
		IDs = &l.fatalProviders
	default:
		panic("Wrong type of the provider.")
	}

	alreadyRegistred := func(id string, idsList *[]string) bool {
		for _, val := range *idsList {
			if val == id {
				return true
			}
		}
		return false
	}

	for _, id := range providersIDs {

		provider, bFound := l.providers[id]
		if bFound {
			pID := (*provider).GetID()
			if !alreadyRegistred(pID, IDs) {
				*IDs = append(*IDs, pID)
			}
		}
	}
}

func (l *Logger) Log(err ...interface{}) {
	if l.level < LEVEL_INFO {
		return
	}
	msg := makeMessage("LOG", err)
	for _, pID := range l.logProviders {
		p, bFound := l.providers[pID]
		if bFound {
			(*p).Log(msg)
		}
	}
}

func (l *Logger) Error(err ...interface{}) {

	msg := makeMessage("ERROR", err)
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
	msg := makeMessage("DEBUG", err)
	for _, pID := range l.debugProviders {
		p, bFound := l.providers[pID]
		if bFound {
			(*p).Debug(msg)
		}
	}
}

func (l *Logger) Fatal(err ...interface{}) {
	msg := makeMessage("FATAL", err)
	for _, pID := range l.fatalProviders {
		p, bFound := l.providers[pID]
		if bFound {
			(*p).Fatal(msg)
		}
	}

	os.Exit(1)
}

var (
	HOST              string
	MESSAGE_REPLACER  = strings.NewReplacer("\r", "", "\n", "\t")
	MESSAGE_SEPARATOR = []byte("\t")
)

func makeMessage(typeLog string, err []interface{}) []byte {

	if len(HOST) == 0 {
		HOST, _ = os.Hostname()
	}

	buf := bytes.NewBuffer(nil)
	prefix := fmt.Sprintf("%s: %s %s ", typeLog, time.Now().Format(time.RFC3339), HOST)
	logger := log.New(buf, prefix, log.Lshortfile)

	msg := bytes.NewBuffer(nil)
	for i, v := range err {
		if i > 0 {
			msg.Write(MESSAGE_SEPARATOR)
		}
		fmt.Fprint(msg, v)
	}

	logger.Output(3, MESSAGE_REPLACER.Replace(msg.String()))

	return bytes.Replace(buf.Bytes(), []byte("\n"), []byte{}, -1)
}
