package utils

import (
	"fmt"
	"io"
	"os"
	"time"
)

type logLevel string

// Declare typed constants each with type of status
const (
	Warn  logLevel = "WARN"
	Info  logLevel = "INFO"
	Error logLevel = "ERROR"
	Debug logLevel = "DEBUG"
)

type Logger interface {
	Info(template string, args ...any)

	Error(template string, args ...any)

	Warn(template string, args ...any)

	Debug(template string, args ...any)

	Log(level logLevel, template string, args ...any)

	Logger(name string) Logger
}

type logger struct {
	name string
	out  io.Writer
}

func NewStdLogger(name string) Logger {
	return logger{
		name: name,
		out:  os.Stdout,
	}
}

func (l logger) Log(level logLevel, template string, args ...any) {
	message := fmt.Sprintf("%s [%s] %s: "+template, append([]any{
		time.Now().Format(time.RFC3339),
		level,
		l.name,
	}, args...)...)
	fmt.Println(message)
}

func (l logger) Info(template string, args ...any) {
	l.Log(Info, template, args...)
}

func (l logger) Debug(template string, args ...any) {
	l.Log(Debug, template, args...)
}

func (l logger) Warn(template string, args ...any) {
	l.Log(Warn, template, args...)
}

func (l logger) Error(template string, args ...any) {
	l.Log(Error, template, args...)
}

func (l logger) Logger(name string) Logger {
	return logger{
		name: l.name + "::" + name,
		out:  l.out,
	}
}
