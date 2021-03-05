package logger

import (
	"fmt"
)

type stdWriter struct{}

func (l *stdWriter) write(msg *message) {
	fmt.Print(string(msg.bytes()))
}

func (l *stdWriter) close() {}
