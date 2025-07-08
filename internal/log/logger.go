package log

import (
	"fmt"
	"log"
	"os"
)

// Logger defines the interface for logging
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
}

// LogLevel represents the logging level
type LogLevel int

const (
	// DEBUG level
	DEBUG LogLevel = iota
	// INFO level
	INFO
	// WARN level
	WARN
	// ERROR level
	ERROR
	// FATAL level
	FATAL
)

// DefaultLogger is a simple logger implementation
type DefaultLogger struct {
	level LogLevel
	debug *log.Logger
	info  *log.Logger
	warn  *log.Logger
	error *log.Logger
	fatal *log.Logger
}

// NewDefaultLogger creates a new default logger with the specified log level
func NewDefaultLogger(level LogLevel) Logger {
	return &DefaultLogger{
		level: level,
		debug: log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
		info:  log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		warn:  log.New(os.Stdout, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile),
		error: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		fatal: log.New(os.Stderr, "FATAL: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// Debug logs a debug message
func (l *DefaultLogger) Debug(args ...interface{}) {
	if l.level <= DEBUG {
		l.debug.Output(2, fmt.Sprint(args...))
	}
}

// Debugf logs a formatted debug message
func (l *DefaultLogger) Debugf(format string, args ...interface{}) {
	if l.level <= DEBUG {
		l.debug.Output(2, fmt.Sprintf(format, args...))
	}
}

// Info logs an info message
func (l *DefaultLogger) Info(args ...interface{}) {
	if l.level <= INFO {
		l.info.Output(2, fmt.Sprint(args...))
	}
}

// Infof logs a formatted info message
func (l *DefaultLogger) Infof(format string, args ...interface{}) {
	if l.level <= INFO {
		l.info.Output(2, fmt.Sprintf(format, args...))
	}
}

// Warn logs a warning message
func (l *DefaultLogger) Warn(args ...interface{}) {
	if l.level <= WARN {
		l.warn.Output(2, fmt.Sprint(args...))
	}
}

// Warnf logs a formatted warning message
func (l *DefaultLogger) Warnf(format string, args ...interface{}) {
	if l.level <= WARN {
		l.warn.Output(2, fmt.Sprintf(format, args...))
	}
}

// Error logs an error message
func (l *DefaultLogger) Error(args ...interface{}) {
	if l.level <= ERROR {
		l.error.Output(2, fmt.Sprint(args...))
	}
}

// Errorf logs a formatted error message
func (l *DefaultLogger) Errorf(format string, args ...interface{}) {
	if l.level <= ERROR {
		l.error.Output(2, fmt.Sprintf(format, args...))
	}
}

// Fatal logs a fatal message and exits
func (l *DefaultLogger) Fatal(args ...interface{}) {
	if l.level <= FATAL {
		l.fatal.Output(2, fmt.Sprint(args...))
		os.Exit(1)
	}
}

// Fatalf logs a formatted fatal message and exits
func (l *DefaultLogger) Fatalf(format string, args ...interface{}) {
	if l.level <= FATAL {
		l.fatal.Output(2, fmt.Sprintf(format, args...))
		os.Exit(1)
	}
}

// Global logger instance
var GlobalLogger Logger = NewDefaultLogger(INFO)

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(logger Logger) {
	GlobalLogger = logger
}

// SetLogLevel sets the log level for the global logger if it's a DefaultLogger
func SetLogLevel(level LogLevel) {
	if l, ok := GlobalLogger.(*DefaultLogger); ok {
		l.level = level
	}
}
