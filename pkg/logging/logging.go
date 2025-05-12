package logging

import (
	"os"

	"github.com/rs/zerolog"
)

type Logger struct {
	logger *zerolog.Logger
}

func NewZeroLogger(level int8) *Logger {

	var zLevel zerolog.Level

	switch level {
	case 0:
		zLevel = zerolog.DebugLevel
	case 1:
		zLevel = zerolog.InfoLevel
	case 2:
		zLevel = zerolog.WarnLevel
	case 3:
		zLevel = zerolog.ErrorLevel
	case 4:
		zLevel = zerolog.FatalLevel
	default:
		zLevel = zerolog.InfoLevel
	}

	zl := zerolog.New(os.Stdout).Level(zLevel)

	return &Logger{
		logger: &zl,
	}
}

func (l *Logger) Debug(msg string, fields map[string]interface{}) {
	event := l.logger.Debug()
	for key, value := range fields {
		event.Interface(key, value)
	}
	event.Msg(msg)
}

func (l *Logger) Info(msg string, fields map[string]interface{}) {
	event := l.logger.Info()
	for key, value := range fields {
		event.Interface(key, value)
	}
	event.Msg(msg)
}

func (l *Logger) Warn(msg string, fields map[string]interface{}) {
	event := l.logger.Warn()
	for key, value := range fields {
		event.Interface(key, value)
	}
	event.Msg(msg)
}

func (l *Logger) Error(msg string, fields map[string]interface{}) {
	event := l.logger.Error()
	for key, value := range fields {
		event.Interface(key, value)
	}
	event.Msg(msg)
}

func (l *Logger) Fatal(msg string, fields map[string]interface{}) {
	event := l.logger.Fatal()
	for key, value := range fields {
		event.Interface(key, value)
	}
	event.Msg(msg)
}
