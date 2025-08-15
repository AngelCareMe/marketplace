package logger

import (
	"marketplace/pkg/config"
	"os"

	"github.com/sirupsen/logrus"
)

type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
}

type LogrusLogger struct {
	logger *logrus.Logger
}

func NewLogger(cfg *config.LoggerConfig) Logger {
	log := logrus.New()

	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	log.SetOutput(os.Stdout)

	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		log.Warnf("Invalid log level %s, using info level", cfg.Level)
		level = logrus.InfoLevel
	}
	log.SetLevel(level)

	return &LogrusLogger{logger: log}
}

func (l *LogrusLogger) Debug(args ...interface{}) {
	l.logger.Debug(args...)
}

func (l *LogrusLogger) Info(args ...interface{}) {
	l.logger.Info(args...)
}

func (l *LogrusLogger) Warn(args ...interface{}) {
	l.logger.Warn(args...)
}

func (l *LogrusLogger) Error(args ...interface{}) {
	l.logger.Error(args...)
}

func (l *LogrusLogger) Fatal(args ...interface{}) {
	l.logger.Fatal(args...)
}

func (l *LogrusLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

func (l *LogrusLogger) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

func (l *LogrusLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

func (l *LogrusLogger) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}

func (l *LogrusLogger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}
