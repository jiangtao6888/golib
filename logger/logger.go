package logger

import (
	"fmt"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/marsmay/golib/net2"
)

type Config struct {
	Dir        string `toml:"dir" json:"dir"`
	Prefix     string `toml:"prefix" json:"prefix"`
	Level      string `toml:"level" json:"level"`
	Color      bool   `toml:"color" json:"color"`
	Terminal   bool   `toml:"terminal" json:"terminal"`
	ShowIp     bool   `toml:"show_ip" json:"show_ip"`
	TimeFormat string `toml:"time_format" json:"time_format"`
}

func DefaultConfig() *Config {
	return &Config{
		Dir:        "./logs",
		Level:      "debug",
		Color:      true,
		Terminal:   true,
		ShowIp:     true,
		TimeFormat: "2006-01-02 15:04:05",
	}
}

type Logger struct {
	conf   *Config
	writer IWriter
	level  Level
	ip     string
}

func NewLogger(conf *Config) (l *Logger, err error) {
	l = &Logger{conf: conf, level: GetLevel(conf.Level)}

	if l.ip, err = net2.GetLocalIp(); err != nil {
		return
	}

	if l.conf.Terminal {
		l.writer = &stdWriter{}
	} else {
		l.writer, err = newAsyncWriter(conf.Dir, l.getFile)
	}

	return
}

func (l *Logger) getFile() string {
	return path.Join(l.conf.Dir, l.conf.Prefix+time.Now().Format("20060102.log"))
}

func (l *Logger) prefix(level Level, file string, line int) string {
	nowTime := time.Now().Format(l.conf.TimeFormat)
	levelText := GetLevelText(level, l.conf.Color)
	loc := fmt.Sprintf("<%s:%d>", file, line)

	if l.conf.Color {
		loc = Blue(loc)
	}

	if l.conf.ShowIp {
		return fmt.Sprintf("%s (%s) %s %s ", levelText, l.ip, nowTime, loc)
	} else {
		return fmt.Sprintf("%s %s %s ", levelText, nowTime, loc)
	}
}

func (l *Logger) getFileInfo() (file string, line int) {
	for i := 1; ; i++ {
		_, f, l, ok := runtime.Caller(i)

		if !ok {
			return "???", 1
		}

		if strings.HasSuffix(f, "logger/logger.go") {
			continue
		}

		if dirs := strings.Split(f, "/"); len(dirs) >= 2 {
			return strings.Join(dirs[len(dirs)-2:], "/"), l
		}
	}
}

func (l *Logger) Write(p []byte) (n int, err error) {
	msg := &message{args: []interface{}{string(p)}, ignoreLF: true}
	l.writer.write(msg)
	n = len(p)
	return
}

func (l *Logger) Log(level Level, format string, args ...interface{}) {
	file, line := l.getFileInfo()
	prefix := l.prefix(level, file, line)
	msg := &message{prefix: prefix, format: format, args: args}

	if level <= l.level {
		l.writer.write(msg)
	}
}

func (l *Logger) Debug(args ...interface{}) {
	l.Log(DebugLevel, "", args...)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Log(DebugLevel, format, args...)
}

func (l *Logger) Info(args ...interface{}) {
	l.Log(InfoLevel, "", args...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.Log(InfoLevel, format, args...)
}

func (l *Logger) Warning(args ...interface{}) {
	l.Log(WarnLevel, "", args...)
}

func (l *Logger) Warningf(format string, args ...interface{}) {
	l.Log(WarnLevel, format, args...)
}

func (l *Logger) Error(args ...interface{}) {
	l.Log(ErrorLevel, "", args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Log(ErrorLevel, format, args...)
}

func (l *Logger) Fatal(args ...interface{}) {
	l.Log(FatalLevel, "", args...)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.Log(FatalLevel, format, args...)
}

func (l *Logger) Config() *Config {
	return l.conf
}

func (l *Logger) Level() Level {
	return l.level
}

func (l *Logger) Close() {
	l.writer.close()
}
