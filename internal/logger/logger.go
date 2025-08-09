package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// LogLevel represents the logging level
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// String returns the string representation of LogLevel
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ParseLogLevel parses a string into a LogLevel
func ParseLogLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN", "WARNING":
		return WARN
	case "ERROR":
		return ERROR
	default:
		return INFO // Default to INFO
	}
}

// Logger represents our application logger
type Logger struct {
	level  LogLevel
	logger *log.Logger
}

// Global logger instance
var globalLogger *Logger

// init initializes the global logger with INFO level by default
func init() {
	globalLogger = &Logger{
		level:  INFO,
		logger: log.New(os.Stderr, "", 0), // No timestamp prefix, we'll add our own
	}
}

// SetLevel sets the global logging level
func SetLevel(level LogLevel) {
	globalLogger.level = level
}

// SetLevelFromString sets the global logging level from string
func SetLevelFromString(level string) {
	SetLevel(ParseLogLevel(level))
}

// GetLevel returns the current logging level
func GetLevel() LogLevel {
	return globalLogger.level
}

// shouldLog checks if a message at the given level should be logged
func (l *Logger) shouldLog(level LogLevel) bool {
	return level >= l.level
}

// logf formats and logs a message at the specified level
func (l *Logger) logf(level LogLevel, format string, args ...interface{}) {
	if !l.shouldLog(level) {
		return
	}
	
	prefix := fmt.Sprintf("[%s] ", level.String())
	message := fmt.Sprintf(format, args...)
	l.logger.Printf("%s%s", prefix, message)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.logf(DEBUG, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.logf(INFO, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.logf(WARN, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.logf(ERROR, format, args...)
}

// Global convenience functions
func Debug(format string, args ...interface{}) {
	globalLogger.Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	globalLogger.Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	globalLogger.Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	globalLogger.Error(format, args...)
}

// PrintDebug prints debug information if debug level is enabled
func PrintDebug(format string, args ...interface{}) {
	if globalLogger.shouldLog(DEBUG) {
		Debug(format, args...)
	}
}

// PrintInfo prints informational messages 
func PrintInfo(format string, args ...interface{}) {
	if globalLogger.shouldLog(INFO) {
		Info(format, args...)
	}
}

// IsDebugEnabled returns true if debug logging is enabled
func IsDebugEnabled() bool {
	return globalLogger.shouldLog(DEBUG)
}

// IsLevelEnabled returns true if the given level is enabled
func IsLevelEnabled(level LogLevel) bool {
	return globalLogger.shouldLog(level)
}
