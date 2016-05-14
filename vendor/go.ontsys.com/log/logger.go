package log

import (
	"io"
	"os"
	"strings"
	"sync"

	logrus "github.com/Sirupsen/logrus"
)

// Logger - is a logger that encapsulates all of the available logging
// functionality as well as instance specific settings for that logger.
type Logger struct {
	sync.Mutex
	globalFields  Fields
	doLongSplit   bool
	longSplitSize int
	logrusLogger  *logrus.Logger
}

// New - creates a new instance of a logger.  Without using a New() logger,
// you will by default be configuring/using a default global logger.
func New() *Logger {
	logger := &Logger{
		logrusLogger: &logrus.Logger{
			Out: os.Stdout,
			Formatter: &logrus.JSONFormatter{
				TimestampFormat: DefaultTimeFormat,
			},
			Hooks: make(logrus.LevelHooks),
			Level: logrus.InfoLevel,
		},
	}
	return logger
}

// UseColorFormatter - by default a JSON formatter is used.  This allows you to
// use the standard text formatter (with colors forced) instead.
func (l *Logger) UseColorFormatter() {
	l.Lock()
	l.logrusLogger.Formatter = &logrus.TextFormatter{ForceColors: true, TimestampFormat: DefaultTimeFormat}
	l.Unlock()
}

// SetSplitLongMessages - forces the given message to be split into chunks of the given size
// and then sent as multiple log statements.
// Only the message is split, and data fields related to the message are repeated for each
// resulting log message
func (l *Logger) SetSplitLongMessages(doSplit bool, splitSize int) {
	l.doLongSplit = doSplit
	l.longSplitSize = splitSize
}

// SetOutput - set the log output target
func (l *Logger) SetOutput(out io.Writer) {
	l.Lock()
	l.logrusLogger.Out = out
	l.Unlock()
}

// SetFormatter - sets the formatter on the logger
func (l *Logger) SetFormatter(formatter logrus.Formatter) {
	l.logrusLogger.Formatter = formatter
}

// GlobalFields - will add the given fields to every subsequent log message.
// Additional calls to WithGlobalFields will overwrite previous values
func (l *Logger) GlobalFields(fields Fields) {
	l.globalFields = fields
}

// MoreGlobalFields - will add the given fields to every subsequent log message.
// This will be appended to the currently active global fields
func (l *Logger) MoreGlobalFields(fields Fields) {
	if fields == nil {
		return
	}

	if l.globalFields == nil {
		l.globalFields = Fields{}
	}

	for k, v := range fields {
		l.globalFields[k] = v
	}
}

// MoreGlobalFlags will take a StringArrayFlags type and add the values as global
// fields.  The values in the StringArrayFlags should be `key:value`
func (l *Logger) MoreGlobalFlags(arrFlags StringArrayFlags) {
	globs := Fields{}
	for _, arrFlag := range arrFlags {
		fSplit := strings.Split(arrFlag, ":")
		if len(fSplit) == 2 {
			globs[fSplit[0]] = fSplit[1]
		}
	}
	l.MoreGlobalFields(globs)
}

func (l *Logger) getWrapper(callDepth int) *WithWrapper {
	return (&WithWrapper{
		logger:        l,
		doLongSplit:   l.doLongSplit,
		longSplitSize: l.longSplitSize,
		callDepth:     callDepth,
	}).WithFields(l.globalFields)
}

// WithField - Adds a field to the log entry.
// If you want multiple fields, you can use `WithFields`.
// This is chainable
func (l *Logger) WithField(key string, value interface{}) *WithWrapper {
	return l.getWrapper(4).WithField(key, value)
}

// WithFields - Adds a map of fields to the log entry. All it does is call `WithField` for
// each value in the map.
// This is chainable
func (l *Logger) WithFields(fields Fields) *WithWrapper {
	return l.getWrapper(4).WithFields(fields)
}

// WithError - Add an error as single field to the log entry.  All it does is call
// `WithField` for the given `err`.  It will also force logging of the caller location.
// This is chainable
func (l *Logger) WithError(err error) *WithWrapper {
	return l.getWrapper(4).WithError(err)
}

// WithCaller - will add the location of the caller.  This is automatically done
// for Warn, Error and Panic.
// This is chainable
func (l *Logger) WithCaller() *WithWrapper {
	return l.getWrapper(4).WithCaller().WithFields(l.globalFields)
}

// Debug - log a non-formatted debug message
// Multiple parameters will be concatenated
func (l *Logger) Debug(args ...interface{}) {
	l.getWrapper(4).WithFields(l.globalFields).Debug(args...)
}

// Info - log a non-formatted info message
// Multiple parameters will be concatenated
func (l *Logger) Info(args ...interface{}) {
	l.getWrapper(4).WithFields(l.globalFields).Info(args...)
}

// Warn - log a non-formatted warn message
// Multiple parameters will be concatenated
func (l *Logger) Warn(args ...interface{}) {
	l.getWrapper(4).WithFields(l.globalFields).Warn(args...)
}

// Error - log a non-formatted error message
// Multiple parameters will be concatenated
func (l *Logger) Error(args ...interface{}) {
	l.getWrapper(6).WithCaller().WithFields(l.globalFields).Error(args...)
}

// Panic - log a non-formatted panic message
// Multiple parameters will be concatenated
// Panic will call panic(message)
func (l *Logger) Panic(args ...interface{}) {
	l.getWrapper(6).WithCaller().WithFields(l.globalFields).Panic(args...)
}

// Debugf - log a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.getWrapper(4).WithFields(l.globalFields).Debugf(format, args...)
}

// Infof - log a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.getWrapper(4).WithFields(l.globalFields).Infof(format, args...)
}

// Warnf - log a formatted warn message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.getWrapper(4).WithFields(l.globalFields).Warnf(format, args...)
}

// Errorf - log a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.getWrapper(6).WithCaller().WithFields(l.globalFields).Errorf(format, args...)
}

// Panicf - log a formatted panic message
// Panic will call panic(message)
func (l *Logger) Panicf(format string, args ...interface{}) {
	l.getWrapper(6).WithCaller().WithFields(l.globalFields).Panicf(format, args...)
}
