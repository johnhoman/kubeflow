package logging

import (
	"github.com/go-logr/logr"
)

type Logger interface {
	Info(msg string, keysAndValues ...interface{})
	Debug(msg string, keysAndValues ...interface{})
	WithValues(keysAndValues ...interface{}) Logger
}

type logger struct {
	log logr.Logger
}

func (l *logger) Debug(msg string, keysAndValues ...interface{}) {
	l.log.V(1).Info(msg, keysAndValues...)
}

func (l *logger) Info(msg string, keysAndValues ...interface{}) {
	l.log.Info(msg, keysAndValues...)
}

func (l *logger) WithValues(keysAndValues ...interface{}) Logger {
	return &logger{log: l.log.WithValues(keysAndValues...)}
}

func NewLogrLogger(log logr.Logger) *logger {
	return &logger{log: log}
}

var _ Logger = &logger{}

type nopLogger struct{}

func (*nopLogger) Debug(msg string, keysAndValues ...interface{}) {}
func (*nopLogger) Info(msg string, keysAndValues ...interface{})  {}

func (n *nopLogger) WithValues(keysAndValues ...interface{}) Logger {
	return n
}

func NewNopLogger() *nopLogger { return &nopLogger{} }

var _ Logger = &nopLogger{}
