package config

import (
	"fmt"
	"os"
	"time"
)

// LogLevel represents the logging level
type LogLevel int

const (
	LogLevelInfo LogLevel = iota
	LogLevelWarn
	LogLevelError
	LogLevelDebug
)

// Logger provides logging functionality
type Logger struct {
	level LogLevel
}

// NewLogger creates a new logger
func NewLogger(debug bool) *Logger {
	level := LogLevelInfo
	if debug {
		level = LogLevelDebug
	}
	return &Logger{level: level}
}

// LogInfo logs an info message
func (l *Logger) LogInfo(format string, args ...interface{}) {
	if l.level <= LogLevelInfo {
		l.log("INFO", format, args...)
	}
}

// LogWarn logs a warning message
func (l *Logger) LogWarn(format string, args ...interface{}) {
	if l.level <= LogLevelWarn {
		l.log("WARN", format, args...)
	}
}

// LogError logs an error message
func (l *Logger) LogError(format string, args ...interface{}) {
	l.log("ERROR", format, args...)
}

// LogDebug logs a debug message
func (l *Logger) LogDebug(format string, args ...interface{}) {
	if l.level <= LogLevelDebug {
		l.log("DEBUG", format, args...)
	}
}

// log outputs a log message with timestamp
func (l *Logger) log(level, format string, args ...interface{}) {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(os.Stderr, "[%s] %s: %s\n", timestamp, level, msg)
}

// Global logger instance
var defaultLogger *Logger

// InitLogger initializes the global logger
func InitLogger(debug bool) {
	defaultLogger = NewLogger(debug)
}

// GetLogger returns the global logger
func GetLogger() *Logger {
	if defaultLogger == nil {
		defaultLogger = NewLogger(false)
	}
	return defaultLogger
}

// Info logs an info message using the global logger
func Info(format string, args ...interface{}) {
	GetLogger().LogInfo(format, args...)
}

// Warn logs a warning message using the global logger
func Warn(format string, args ...interface{}) {
	GetLogger().LogWarn(format, args...)
}

// Error logs an error message using the global logger
func Error(format string, args ...interface{}) {
	GetLogger().LogError(format, args...)
}

// Debug logs a debug message using the global logger
func Debug(format string, args ...interface{}) {
	GetLogger().LogDebug(format, args...)
}
