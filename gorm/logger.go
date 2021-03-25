package gorm

import (
	"context"
	"time"

	oLogger "github.com/marsmay/golib/logger"
	gLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
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
	l.l.Infof(format, args)
}

func (l *logger) Warn(_ context.Context, format string, args ...interface{}) {
	l.l.Warningf(format, args)
}

func (l *logger) Error(_ context.Context, format string, args ...interface{}) {
	l.l.Errorf(format, args)
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
			l.l.Errorf("query: <%s> | %4v | - | %s", utils.FileWithLineNum(), useTime, sql)
		} else {
			l.l.Errorf("query: <%s> | %4v | %d rows | %s", utils.FileWithLineNum(), useTime, rows, sql)
		}

		return
	}

	if l.slowQueryTime > 0 && useTime > l.slowQueryTime && logLevel >= oLogger.WarnLevel {
		sql, rows := fc()

		if rows == -1 {
			l.l.Warningf("query: <%s> | %4v | - | %s", utils.FileWithLineNum(), useTime, sql)
		} else {
			l.l.Warningf("query: <%s> | %4v | %d rows | %s", utils.FileWithLineNum(), useTime, rows, sql)
		}

		return
	}

	if logLevel >= oLogger.InfoLevel {
		sql, rows := fc()

		if rows == -1 {
			l.l.Infof("query: <%s> | %4v | - | %s", utils.FileWithLineNum(), useTime, sql)
		} else {
			l.l.Infof("query: <%s> | %4v | %d rows | %s", utils.FileWithLineNum(), useTime, rows, sql)
		}

		return
	}
}
