package bytes_pool

import (
	"bytes"
	"sync"
)

var pool = &sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func Get() *bytes.Buffer {
	return pool.Get().(*bytes.Buffer)
}

func Return(bf *bytes.Buffer) {
	bf.Reset()
	pool.Put(bf)
}
