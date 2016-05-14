package log

import (
	"fmt"
	"strconv"

	"github.com/Sirupsen/logrus"
)

// WithWrapper - A wrapper of one log call that allows for message configuration chaining
type WithWrapper struct {
	logger        *Logger
	withFields    Fields
	caller        bool
	doLongSplit   bool
	longSplitSize int
	callDepth     int
}

// WithField - Adds a field to the log entry.
// If you want multiple fields, you can use `WithFields`.
// This is chainable
func (l *WithWrapper) WithField(key string, value interface{}) *WithWrapper {
	if l.withFields == nil {
		l.withFields = Fields{}
	}
	l.withFields[key] = value
	return l
}

// WithFields - Adds a map of fields to the log entry. All it does is call `WithField` for
// each value in the map.
// This is chainable
func (l *WithWrapper) WithFields(fields Fields) *WithWrapper {
	if fields == nil {
		return l
	}
	if l.withFields == nil {
		l.withFields = Fields{}
	}
	for k, v := range fields {
		l.withFields[k] = v
	}
	return l
}

// WithError - Add an error as single field to the log entry.  All it does is call
// `WithField` for the given `err`.
// This is chainable
func (l *WithWrapper) WithError(err error) *WithWrapper {
	l.WithCaller()
	l.WithField("error", err.Error())
	return l
}

// WithCaller - will add the location of the caller.  This is automatically done
// for Warn, Error and Panic.
// This is chainable
func (l *WithWrapper) WithCaller() *WithWrapper {
	l.caller = true
	return l
}

func (l *WithWrapper) finalize(level Level) *logrus.Entry {
	// always log caller for Warn...Panic, otherwise only if the user requested
	if l.caller && l.logger.WillLog(level) {
		// if the calldepth was never set then we don't know what to look for, just don't
		if l.callDepth > 0 {
			callerInfo := CallerInfo(l.callDepth)
			l.WithField("file", callerInfo.File)
			l.WithField("line", strconv.Itoa(callerInfo.Line))
			l.WithField("func", callerInfo.FunctionName)
		}
	}

	logrusFields := logrus.Fields{}
	for k, v := range l.withFields {
		logrusFields[k] = v
	}

	return l.logger.logrusLogger.WithFields(logrusFields)
}

func (l *WithWrapper) getChunks(args ...interface{}) []string {
	fullMessage := fmt.Sprint(args...)

	if !l.doLongSplit || l.longSplitSize == 0 || len(fullMessage) <= l.longSplitSize {
		return []string{fullMessage}
	}

	var chunks []string

	mlen := len(fullMessage)
	for x := 0; x < mlen; x += l.longSplitSize {
		left := x
		// right of slice is exclusive
		right := x + l.longSplitSize
		if right > (mlen) {
			right = mlen
		}
		chunks = append(chunks, fullMessage[left:right])
	}
	return chunks
}

// Debug - log a non-formatted debug message
// Multiple parameters will be concatenated
func (l *WithWrapper) Debug(args ...interface{}) {
	entry := l.finalize(DebugLevel)
	for _, chunk := range l.getChunks(args...) {
		entry.Debug(chunk)
	}
}

// Info - log a non-formatted info message
// Multiple parameters will be concatenated
func (l *WithWrapper) Info(args ...interface{}) {
	l.finalize(InfoLevel).Info(args...)
}

// Warn - log a non-formatted warn message
// Multiple parameters will be concatenated
func (l *WithWrapper) Warn(args ...interface{}) {
	l.finalize(WarnLevel).Warn(args...)
}

// Error - log a non-formatted error message
// Multiple parameters will be concatenated
func (l *WithWrapper) Error(args ...interface{}) {
	l.finalize(ErrorLevel).Error(args...)
}

// Panic - log a non-formatted panic message
// Multiple parameters will be concatenated
// Panic will call panic(message)
func (l *WithWrapper) Panic(args ...interface{}) {
	l.finalize(PanicLevel).Panic(args...)
}

// Debugf - log a formatted debug message
func (l *WithWrapper) Debugf(format string, args ...interface{}) {
	l.finalize(DebugLevel).Debugf(format, args...)
}

// Infof - log a formatted info message
func (l *WithWrapper) Infof(format string, args ...interface{}) {
	l.finalize(InfoLevel).Infof(format, args...)
}

// Warnf - log a formatted warn message
func (l *WithWrapper) Warnf(format string, args ...interface{}) {
	l.finalize(WarnLevel).Warnf(format, args...)
}

// Errorf - log a formatted error message
func (l *WithWrapper) Errorf(format string, args ...interface{}) {
	l.finalize(ErrorLevel).Errorf(format, args...)
}

// Panicf - log a formatted panic message
// Panic will call panic(message)
func (l *WithWrapper) Panicf(format string, args ...interface{}) {
	l.finalize(PanicLevel).Panicf(format, args...)
}
