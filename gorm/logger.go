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
	deadline time.Duration
	recorder *oLogger.Logger
}

// LogMode log mode
func (l *logger) LogMode(level gLogger.LogLevel) gLogger.Interface {
	return l
}

func (l *logger) Info(_ context.Context, format string, args ...interface{}) {
	l.recorder.Infof(format, args)
}

func (l *logger) Warn(_ context.Context, format string, args ...interface{}) {
	l.recorder.Warningf(format, args)
}

func (l *logger) Error(_ context.Context, format string, args ...interface{}) {
	l.recorder.Errorf(format, args)
}

func (l *logger) Trace(_ context.Context, begin time.Time, fc func() (string, int64), err error) {
	logLevel := l.recorder.Level()

	if logLevel == oLogger.DisableLevel {
		return
	}

	useTime := time.Since(begin)

	var printer func(string, ...interface{})

	if err != nil && logLevel >= oLogger.ErrorLevel {
		printer = l.recorder.Errorf
	} else if l.deadline > 0 && useTime > l.deadline && logLevel >= oLogger.WarnLevel {
		printer = l.recorder.Warningf
	} else if logLevel >= oLogger.DebugLevel {
		printer = l.recorder.Debugf
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
			printer("query: <%s> | %4v | - | %s | %s", source, useTime, sql, err)
		} else {
			printer("query: <%s> | %4v | %d rows | %s | %s", source, useTime, rows, sql, err)
		}
	} else {
		if rows == -1 {
			printer("query: <%s> | %4v | - | %s", source, useTime, sql)
		} else {
			printer("query: <%s> | %4v | %d rows | %s", source, useTime, rows, sql)
		}
	}
}
