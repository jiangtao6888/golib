package logger

import (
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/marsmay/golib/net2"
)

var localIp string
var sourceDir string

func init() {
	_, file, _, _ := runtime.Caller(0)
	sourceDir, _ = path.Split(file)

	if ip, e := net2.GetLocalIp(); e == nil {
		localIp = ip.String()
	}
}

type Config struct {
	Dir        string `toml:"dir" json:"dir"`
	Prefix     string `toml:"prefix" json:"prefix"`
	Level      string `toml:"level" json:"level"`
	Color      bool   `toml:"color" json:"color"`
	Terminal   bool   `toml:"terminal" json:"terminal"`
	ShowIp     bool   `toml:"show_ip" json:"show_ip"`
	UseUtc     bool   `toml:"use_utc" json:"use_utc"`
	TimeFormat string `toml:"time_format" json:"time_format"`
}

func DefaultConfig() *Config {
	return &Config{
		Dir:        "./logs",
		Level:      "debug",
		Color:      true,
		Terminal:   true,
		ShowIp:     false,
		UseUtc:     false,
		TimeFormat: time.RFC3339,
	}
}

type Logger struct {
	conf   *Config
	writer io.WriteCloser
	level  Level
}

func NewLogger(conf *Config) (l *Logger, err error) {
	l = &Logger{conf: conf, level: GetLevel(conf.Level)}

	if l.conf.Terminal {
		l.writer = os.Stdout
	} else {
		l.writer, err = newAsyncWriter(conf.Dir, l.getFile)
	}

	return
}

func (l *Logger) now() time.Time {
	t := time.Now()

	if l.conf.UseUtc {
		t = t.UTC()
	}

	return t
}

func (l *Logger) getFile() string {
	return path.Join(l.conf.Dir, l.conf.Prefix+l.now().Format("20060102.log"))
}

func (l *Logger) prefix(level Level) string {
	var (
		formaters []string
		params    []interface{}
	)

	if l.conf.ShowIp && localIp != "" {
		formaters = append(formaters, "(%s)")
		params = append(params, localIp)
	}

	nowTime := time.Now().Format(l.conf.TimeFormat)
	levelText := GetLevelText(level, l.conf.Color)
	formaters = append(formaters, "%s %s")
	params = append(params, levelText, nowTime)

	if f := "<%s:%d>"; l.conf.Color {
		formaters = append(formaters, Blue(f))
	} else {
		formaters = append(formaters, f)
	}

	file, line := l.getFileInfo()
	params = append(params, file, line)

	return fmt.Sprintf(strings.Join(formaters, " "), params...)
}

func (l *Logger) getFileInfo() (file string, line int) {
	for i := 2; i < 15; i++ {
		_, f, n, ok := runtime.Caller(i)

		if ok && (!strings.HasPrefix(f, sourceDir) || strings.HasSuffix(f, "_test.go")) {
			if items := strings.Split(f, "/"); len(items) >= 2 {
				return items[len(items)-2] + "/" + items[len(items)-1], n
			}

			return f, n
		}
	}

	return "???", 0
}

func (l *Logger) Write(p []byte) (n int, err error) {
	return l.writer.Write(p)
}

func (l *Logger) Log(level Level, message string) {
	if level <= l.level {
		_, _ = fmt.Fprintln(l.writer, l.prefix(level), message)
	}
}

func (l *Logger) Debug(args ...interface{}) {
	l.Log(DebugLevel, fmt.Sprint(args...))
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Log(DebugLevel, fmt.Sprintf(format, args...))
}

func (l *Logger) Info(args ...interface{}) {
	l.Log(InfoLevel, fmt.Sprint(args...))
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.Log(InfoLevel, fmt.Sprintf(format, args...))
}

func (l *Logger) Warning(args ...interface{}) {
	l.Log(WarnLevel, fmt.Sprint(args...))
}

func (l *Logger) Warningf(format string, args ...interface{}) {
	l.Log(WarnLevel, fmt.Sprintf(format, args...))
}

func (l *Logger) Error(args ...interface{}) {
	l.Log(ErrorLevel, fmt.Sprint(args...))
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Log(ErrorLevel, fmt.Sprintf(format, args...))
}

func (l *Logger) Fatal(args ...interface{}) {
	l.Log(FatalLevel, fmt.Sprint(args...))
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.Log(FatalLevel, fmt.Sprintf(format, args...))
}

func (l *Logger) Config() *Config {
	return l.conf
}

func (l *Logger) Level() Level {
	return l.level
}

func (l *Logger) Close() {
	_ = l.writer.Close()
}
