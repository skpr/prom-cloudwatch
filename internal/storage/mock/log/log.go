package log

import "fmt"

// Logger for testing the storage package.
type Logger struct {
	Messages []string
}

// New mock logger.
func New() *Logger {
	return &Logger{}
}

// Infof mock implementation.
func (l *Logger) Infof(format string, args ...interface{}) {
	l.Messages = append(l.Messages, fmt.Sprintf(format, args...))
}
