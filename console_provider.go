package logger

import "fmt"

const PROVIDER_CONSOLE = "console"

type ConsoleProvider struct {
}

func (p ConsoleProvider) GetID() string {
	return PROVIDER_CONSOLE
}

func (p ConsoleProvider) Log(msg []byte) {
	fmt.Println(string(msg))
}

func (p ConsoleProvider) Error(msg []byte) {
	fmt.Println(string(msg))
}

func (p ConsoleProvider) Fatal(msg []byte) {
	fmt.Println(string(msg))
}

func (p ConsoleProvider) Debug(msg []byte) {
	fmt.Println(string(msg))
}
