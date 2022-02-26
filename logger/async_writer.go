package logger

import (
	"bufio"
	"context"
	"os"
	"time"
)

type asyncWriter struct {
	dir      string
	file     string
	fd       *os.File
	writer   *bufio.Writer
	msgQueue chan []byte
	timer    *time.Ticker
	getFile  func() string
	ctx      context.Context
	cancel   context.CancelFunc
	end      chan bool
}

func newAsyncWriter(dir string, getFile func() string) (writer *asyncWriter, err error) {
	writer = &asyncWriter{
		dir:      dir,
		getFile:  getFile,
		msgQueue: make(chan []byte, 8192),
		timer:    time.NewTicker(time.Second),
		end:      make(chan bool, 1),
	}
	writer.ctx, writer.cancel = context.WithCancel(context.Background())

	if err = os.MkdirAll(writer.dir, 0755); err != nil {
		return
	}

	writer.refresh()
	writer.fd, err = os.OpenFile(writer.file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return
	}

	writer.writer = bufio.NewWriter(writer.fd)

	go writer.flush()
	go writer.start()

	return
}

func (l *asyncWriter) refresh() bool {
	oldFile := l.file
	l.file = l.getFile()
	return l.file != oldFile
}

func (l *asyncWriter) start() {
	for msg := range l.msgQueue {
		if msg == nil {
			_ = l.writer.Flush()

			if l.refresh() {
				_ = l.fd.Close()
				l.fd, _ = os.OpenFile(l.file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				l.writer.Reset(l.fd)
			}
		} else {
			_, _ = l.writer.Write(msg)
		}
	}

	l.end <- true
}

func (l *asyncWriter) flush() {
	for range l.timer.C {
		l.msgQueue <- nil
	}
}

func (l *asyncWriter) Write(p []byte) (n int, err error) {
	select {
	case <-l.ctx.Done():
	default:
		l.msgQueue <- p
	}

	return len(p), nil
}

func (l *asyncWriter) Close() error {
	l.cancel()
	l.timer.Stop()
	l.msgQueue <- nil
	close(l.msgQueue)
	<-l.end
	return nil
}
