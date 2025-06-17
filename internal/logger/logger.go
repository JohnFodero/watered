package logger

import (
	"log"
	"os"
)

// Logger provides structured logging for the application
type Logger struct {
	*log.Logger
}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "[WATERED] ", log.LstdFlags|log.Lshortfile),
	}
}

// Info logs informational messages
func (l *Logger) Info(msg string) {
	l.Printf("INFO: %s", msg)
}

// Error logs error messages
func (l *Logger) Error(msg string, err error) {
	if err != nil {
		l.Printf("ERROR: %s - %v", msg, err)
	} else {
		l.Printf("ERROR: %s", msg)
	}
}

// Debug logs debug messages
func (l *Logger) Debug(msg string) {
	l.Printf("DEBUG: %s", msg)
}