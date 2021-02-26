package csv

import (
	"encoding/csv"
	"os"
)

type Writer struct {
	filename string
	fd       *os.File
	writer   *csv.Writer
}

func (w *Writer) WriteLine(line []string) (err error) {
	return w.writer.Write(line)
}

func (w *Writer) WriteLines(lines [][]string) (err error) {
	return w.writer.WriteAll(lines)
}

func (w *Writer) Flush() {
	w.writer.Flush()
}

func (w *Writer) Close() {
	w.writer.Flush()
	_ = w.fd.Close()
}

func NewWriter(filename string) (w *Writer, err error) {
	fd, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)

	if err != nil {
		return
	}

	w = &Writer{
		filename: filename,
		fd:       fd,
		writer:   csv.NewWriter(fd),
	}
	return
}
