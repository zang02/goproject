package woodlog

import (
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

// Note: If you want, you could extend this code to support additional severity levels such as DEBUG and WARNING.

type Level int8

// Initialize constants which represent a specific severity level. We use the iota
// keyword as a shortcut to assign successive integer values to the constants.
const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarning
	LevelError
	LevelFatal
	LevelOff
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarning:
		return "WARNING"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	case LevelOff:
		return "OFF"
	default:
		return ""
	}
}

// Define a custom Logger type. This holds the output destination that the log entries
// will be written to, the minimum severity level that log entries will be written for,
// plus a mutex for coordinating the writes.
type Logger struct {
	out      io.Writer
	minLevel Level
	mu       sync.Mutex
}

// Return a new Logger instance which writes log entries at or above a minimum severity
// level to a specific output destination.
func New(out io.Writer, minLevel Level) *Logger {
	return &Logger{
		out:      out,
		minLevel: minLevel,
	}
}

// Print is an internal method for writing the log entry.
func (l *Logger) print(level Level, message string, properties string) (int, error) {
	// If the severity level of the log entry is below the minimum severity for the
	// logger, then return with no further action.

	if level < l.minLevel {
		return 0, nil
	}

	aux := struct {
		Level      string `json:"level"`
		Message    string `json:"message"`
		Properties string `json:"properties,omitempty"`
		Trace      string `json:"trace,omitempty"`
		Time       string `json:"time,omitempty"`
	}{
		Level:      level.String(),
		Time:       time.Now().UTC().Format("2006-01-02 15:04:05"),
		Message:    message,
		Properties: properties,
	}

	// Include a stack trace for entries at the ERROR and FATAL levels.
	if level >= LevelFatal {
		aux.Trace = string(debug.Stack())
	}

	line := []byte(fmt.Sprintf("%s | %s | %s | %s | %s", aux.Time, aux.Level, aux.Message, aux.Properties, aux.Trace))

	// Lock the mutex so that no two writes to the output destination cannot happen
	// concurrently. If we don't do this, it's possible that the text for two or more
	// log entries will be intermingled in the output.
	l.mu.Lock()
	defer l.mu.Unlock()

	// Write the log entry followed by a newline.
	return l.out.Write(append(line, '\n'))
}

// Declare some helper methods for writing log entries at the different levels. Notice
// that these all accept a map as the second parameter which can contain any arbitrary
// 'properties' that you want to appear in the log entry.
func (l *Logger) PrintInfo(message string, properties string) {
	l.print(LevelInfo, message, properties)
}
func (l *Logger) PrintWarning(message string, properties string) {
	l.print(LevelWarning, message, properties)
}
func (l *Logger) PrintError(err string, properties string) {
	l.print(LevelError, err, properties)
}
func (l *Logger) PrintFatal(err string, properties string) {
	l.print(LevelFatal, err, properties)
	os.Exit(1) // For entries at the FATAL level, we also terminate the application.
}
func (l *Logger) PrintDebug(message string, properties string) {
	l.print(LevelDebug, message, properties)
}

// We also implement a Write() method on our Logger type so that it satisfies the
// io.Writer interface. This writes a log entry at the ERROR level with no additional // properties.
func (l *Logger) Write(message []byte) (n int, err error) {
	return l.print(LevelError, string(message), "")
}
