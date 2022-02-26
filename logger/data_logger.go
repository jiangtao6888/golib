package logger

import (
	"fmt"
	"io"
	"path"
	"time"
)

const (
	PartitionDay = iota
	PartitionHour
)

type DataConfig struct {
	Dir       string         `toml:"dir" json:"dir"`
	Prefix    string         `toml:"prefix" json:"prefix"`
	Partition int            `toml:"partition" json:"partition"`
	Timezone  string         `toml:"timezone" json:"timezone"`
	Location  *time.Location `toml:"-" json:"-"`
}

func (c *DataConfig) LoadLoc() {
	c.Location = time.Local

	if c.Timezone != "" {
		if loc, err := time.LoadLocation(c.Timezone); err == nil {
			c.Location = loc
		}
	}
}

func DefaultDataConfig() *DataConfig {
	return &DataConfig{
		Dir:       "./data",
		Partition: PartitionDay,
	}
}

type DataLogger struct {
	conf   *DataConfig
	writer io.WriteCloser
}

func NewDataLogger(conf *DataConfig) (l *DataLogger, err error) {
	conf.LoadLoc()

	l = &DataLogger{conf: conf}
	l.writer, err = newAsyncWriter(conf.Dir, l.getFile)
	return
}

func (l *DataLogger) getFile() string {
	nowTime := time.Now().In(l.conf.Location)
	year, month, day := nowTime.Date()
	hour := nowTime.Hour()

	switch l.conf.Partition {
	case PartitionHour:
		return path.Join(l.conf.Dir, fmt.Sprintf("%s%04d%02d%02d.%02d.log", l.conf.Prefix, year, month, day, hour))
	default:
		return path.Join(l.conf.Dir, fmt.Sprintf("%s%04d%02d%02d.log", l.conf.Prefix, year, month, day))
	}
}

func (l *DataLogger) Log(args ...interface{}) {
	_, _ = fmt.Fprintln(l.writer, args...)
}

func (l *DataLogger) Logf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(l.writer, format+"\n", args...)
}

func (l *DataLogger) Config() *DataConfig {
	return l.conf
}

func (l *DataLogger) Close() {
	_ = l.writer.Close()
}
