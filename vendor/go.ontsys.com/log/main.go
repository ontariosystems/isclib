package log

import (
	"github.com/Sirupsen/logrus"
)

const (
	// DefaultTimeFormat is a time format with microseconds
	DefaultTimeFormat = "2006-01-02T15:04:05.000Z"
)

func init() {
	DefaultLogger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: DefaultTimeFormat,
	})
}

// DefaultLogger - is the default logger if you don't use New()
var DefaultLogger = &Logger{
	logrusLogger: logrus.StandardLogger(),
}

// SetLevel - set the level
func SetLevel(level Level) {
	DefaultLogger.SetLevel(level)
}

// SetLevelFromString - set the level from string
func SetLevelFromString(level string) {
	DefaultLogger.SetLevelFromString(level)
}

// UseColorFormatter - by default a JSON formatter is used.  This allows you to
// use the standard text formatter (with colors forced) instead.
func UseColorFormatter() {
	DefaultLogger.UseColorFormatter()
}

// SetSplitLongMessages - forces the given message to be split into chunks of the given size
// and then sent as multiple log statements.
// Only the message is split, and data fields related to the message are repeated for each
// resulting log message
func SetSplitLongMessages(doSplit bool, splitSize int) {
	DefaultLogger.SetSplitLongMessages(doSplit, splitSize)
}

// GlobalFields - will add the given fields to every subsequent log message.
// Additional calls to WithGlobalFields will overwrite previous values
func GlobalFields(fields Fields) {
	DefaultLogger.GlobalFields(fields)
}

// MoreGlobalFields - will add the given fields to every subsequent log message.
// This will be appended to the currently active global fields
func MoreGlobalFields(fields Fields) {
	DefaultLogger.MoreGlobalFields(fields)
}

// MoreGlobalFlags will take a StringArrayFlags type and add the values as global
// fields.  The values in the StringArrayFlags should be `key:value`
func MoreGlobalFlags(arrFlags StringArrayFlags) {
	DefaultLogger.MoreGlobalFlags(arrFlags)
}

// WithField - Adds a field to the log entry.
// If you want multiple fields, you can use `WithFields`.
// This is chainable
func WithField(key string, value interface{}) *WithWrapper {
	return DefaultLogger.WithField(key, value)
}

// WithFields - Adds a map of fields to the log entry. All it does is call `WithField` for
// each value in the map.
// This is chainable
func WithFields(fields Fields) *WithWrapper {
	return DefaultLogger.WithFields(fields)
}

// WithError - Add an error as single field to the log entry.  All it does is call
// `WithField` for the given `err`.  It will also force logging of the caller location.
// This is chainable
func WithError(err error) *WithWrapper {
	return DefaultLogger.WithError(err)
}

// WithCaller - will add the location of the caller.  This is automatically done
// for Warn, Error and Panic.
// This is chainable
func WithCaller() *WithWrapper {
	return DefaultLogger.WithCaller()
}

// Debug - log a non-formatted debug message
// Multiple parameters will be concatenated
func Debug(args ...interface{}) {
	DefaultLogger.Debug(args...)
}

// Info - log a non-formatted info message
// Multiple parameters will be concatenated
func Info(args ...interface{}) {
	DefaultLogger.Info(args...)
}

// Warn - log a non-formatted warn message
// Multiple parameters will be concatenated
func Warn(args ...interface{}) {
	DefaultLogger.Warn(args...)
}

// Error - log a non-formatted error message
// Multiple parameters will be concatenated
func Error(args ...interface{}) {
	DefaultLogger.Error(args...)
}

// Panic - log a non-formatted panic message
// Multiple parameters will be concatenated
// Panic will call panic(message)
func Panic(args ...interface{}) {
	DefaultLogger.Panic(args...)
}

// Debugf - log a formatted debug message
func Debugf(format string, args ...interface{}) {
	DefaultLogger.Debugf(format, args...)
}

// Infof - log a formatted info message
func Infof(format string, args ...interface{}) {
	DefaultLogger.Infof(format, args...)
}

// Warnf - log a formatted warn message
func Warnf(format string, args ...interface{}) {
	DefaultLogger.Warnf(format, args...)
}

// Errorf - log a formatted error message
func Errorf(format string, args ...interface{}) {
	DefaultLogger.Errorf(format, args...)
}

// Panicf - log a formatted panic message
// Panic will call panic(message)
func Panicf(format string, args ...interface{}) {
	DefaultLogger.Panicf(format, args...)
}
