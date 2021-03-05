package sync_pool

import (
	"bytes"
	"sync"
)

var BytePool *sync.Pool

func init() {
	BytePool = &sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
}

func ResetBytePool(w *bytes.Buffer) {
	w.Reset()
	BytePool.Put(w)
}
