package gorm

import (
	"context"
	"time"

	oLogger "github.com/marsmay/golib/logger"
	gLogger "gorm.io/gorm/logger"
)

type logger struct {
	slowQueryTime time.Duration
	l             *oLogger.Logger
}

// LogMode log mode
func (l *logger) LogMode(level gLogger.LogLevel) gLogger.Interface {
	return l
}

func (l *logger) Info(_ context.Context, format string, args ...interface{}) {
	l.l.Log(oLogger.InfoLevel, format, args)
}

func (l *logger) Warn(_ context.Context, format string, args ...interface{}) {
	l.l.Log(oLogger.WarnLevel, format, args)
}

func (l *logger) Error(_ context.Context, format string, args ...interface{}) {
	l.l.Log(oLogger.ErrorLevel, format, args)
}

func (l *logger) Trace(_ context.Context, begin time.Time, fc func() (string, int64), err error) {
	logLevel := l.l.Level()

	if logLevel == oLogger.DisableLevel {
		return
	}

	useTime := time.Since(begin)

	if err != nil && logLevel >= oLogger.ErrorLevel {
		sql, rows := fc()

		if rows == -1 {
			l.l.Log(oLogger.ErrorLevel, "query: %4v | - | %s", useTime, sql)
		} else {
			l.l.Log(oLogger.ErrorLevel, "query: %4v | %d rows | %s", useTime, rows, sql)
		}

		return
	}

	if l.slowQueryTime > 0 && useTime > l.slowQueryTime && logLevel >= oLogger.WarnLevel {
		sql, rows := fc()

		if rows == -1 {
			l.l.Log(oLogger.WarnLevel, "query: %4v | - | %s", useTime, sql)
		} else {
			l.l.Log(oLogger.WarnLevel, "query: %4v | %d rows | %s", useTime, rows, sql)
		}

		return
	}

	if logLevel >= oLogger.InfoLevel {
		sql, rows := fc()

		if rows == -1 {
			l.l.Log(oLogger.InfoLevel, "query: %4v | - | %s", useTime, sql)
		} else {
			l.l.Log(oLogger.InfoLevel, "query: %4v | %d rows | %s", useTime, rows, sql)
		}

		return
	}
}
