package logx

import (
	"io"
	"log"
	"os"
	"sync"
)

// Level indicates log severity.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger is a lightweight structured logger built on the stdlib log package.
type Logger struct {
	mu     sync.Mutex
	level  Level
	logger *log.Logger
	fields map[string]string
}

// New creates a logger writing to stderr at info level.
func New() *Logger {
	return &Logger{
		level:  LevelInfo,
		logger: log.New(os.Stderr, "", log.LstdFlags|log.Lmicroseconds),
		fields: make(map[string]string),
	}
}

// WithOutput redirects log output.
func (l *Logger) WithOutput(w io.Writer) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger = log.New(w, "", log.LstdFlags|log.Lmicroseconds)
	return l
}

// WithLevel sets the minimum log level.
func (l *Logger) WithLevel(level Level) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
	return l
}

// WithField returns a child logger carrying an extra field.
func (l *Logger) WithField(key, value string) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	child := &Logger{
		level:  l.level,
		logger: l.logger,
		fields: make(map[string]string, len(l.fields)+1),
	}
	for k, v := range l.fields {
		child.fields[k] = v
	}
	child.fields[key] = value
	return child
}

func (l *Logger) logf(level Level, format string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if level < l.level {
		return
	}
	prefix := level.String()
	if len(l.fields) > 0 {
		prefix += " "
		for k, v := range l.fields {
			prefix += k + "=" + v + " "
		}
	}
	l.logger.Printf(prefix+format, args...)
}

// Debug logs a debug message.
func (l *Logger) Debug(format string, args ...any) { l.logf(LevelDebug, format, args...) }

// Info logs an info message.
func (l *Logger) Info(format string, args ...any) { l.logf(LevelInfo, format, args...) }

// Warn logs a warning message.
func (l *Logger) Warn(format string, args ...any) { l.logf(LevelWarn, format, args...) }

// Error logs an error message.
func (l *Logger) Error(format string, args ...any) { l.logf(LevelError, format, args...) }
