package gorm

import (
	"context"
	"strings"
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

	var printer func(string, ...interface{})

	if err != nil && logLevel >= oLogger.ErrorLevel {
		printer = l.l.Errorf
	} else if l.slowQueryTime > 0 && useTime > l.slowQueryTime && logLevel >= oLogger.WarnLevel {
		printer = l.l.Warningf
	} else if logLevel >= oLogger.DebugLevel {
		printer = l.l.Debugf
	} else {
		return
	}

	source := utils.FileWithLineNum()

	if dirs := strings.Split(source, "/"); len(dirs) >= 3 {
		source = strings.Join(dirs[len(dirs)-3:], "/")
	}

	sql, rows := fc()

	if err != nil {
		if rows == -1 {
			printer("query: <%s> | %4v | - | %s | Error: %s", source, useTime, sql, err)
		} else {
			printer("query: <%s> | %4v | %d rows | %s | Error: %s", source, useTime, rows, sql, err)
		}
	} else {
		if rows == -1 {
			printer("query: <%s> | %4v | - | %s", source, useTime, sql)
		} else {
			printer("query: <%s> | %4v | %d rows | %s", source, useTime, rows, sql)
		}
	}
}
