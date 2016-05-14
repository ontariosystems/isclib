package log

import "github.com/Sirupsen/logrus"

// Level type
type Level uint8

// These are the different logging levels. You can set the logging level to log
// on your instance of logger, obtained with `logrus.New()`.
const (
	// PanicLevel level, highest level of severity. Logs and then calls panic with the
	// message passed to Debug, Info, ...
	PanicLevel Level = iota
	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	ErrorLevel
	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel
	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	InfoLevel
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel
)

// SetLevel - set the level
func (l *Logger) SetLevel(level Level) {
	switch level {
	case DebugLevel:
		l.logrusLogger.Level = logrus.DebugLevel
	case InfoLevel:
		l.logrusLogger.Level = logrus.InfoLevel
	case WarnLevel:
		l.logrusLogger.Level = logrus.WarnLevel
	case ErrorLevel:
		l.logrusLogger.Level = logrus.ErrorLevel
	case PanicLevel:
		l.logrusLogger.Level = logrus.PanicLevel
	}
}

// SetLevelFromString - set the level from string
func (l *Logger) SetLevelFromString(level string) {
	logrusLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		logrus.SetLevel(logrusLevel)
	}
}

// WillLog - return true if the current log level is set so that the given
// level will be logged.
func (l *Logger) WillLog(level Level) bool {
	return int(level) <= int(logrus.GetLevel())
}
