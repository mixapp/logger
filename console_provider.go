package logger

import "fmt"

const PROVIDER_CONSOLE = "console"

type consoleProvider struct {
}

func (p consoleProvider) GetID() string {
	return PROVIDER_CONSOLE
}

func (p consoleProvider) Log(msg []byte) {

	fmt.Println(string(msg))

}

func (p consoleProvider) Error(msg []byte) {

	fmt.Println(string(msg))
}

func (p consoleProvider) Fatal(msg []byte) {

	fmt.Println(string(msg))
}

func (p consoleProvider) Debug(msg []byte) {

	fmt.Println(string(msg))
}
