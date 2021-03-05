package logger

import (
	"fmt"

	"github.com/marsmay/golib/bytes_pool"
)

type message struct {
	prefix   string
	format   string
	args     []interface{}
	ignoreLF bool
}

func (msg *message) bytes() []byte {
	w := bytes_pool.Get()

	defer func() {
		_ = recover()
		w.Reset()
		bytes_pool.Return(w)
	}()

	if len(msg.prefix) > 0 {
		_, _ = fmt.Fprintf(w, msg.prefix)
	}

	if len(msg.format) == 0 {
		for i := 0; i < len(msg.args); i++ {
			if i > 0 {
				w.Write([]byte{' '})
			}

			_, _ = fmt.Fprint(w, msg.args[i])
		}
	} else {
		_, _ = fmt.Fprintf(w, msg.format, msg.args...)
	}

	if !msg.ignoreLF {
		_, _ = fmt.Fprintf(w, "\n")
	}

	b := make([]byte, w.Len())
	copy(b, w.Bytes())

	return b
}
