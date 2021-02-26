package mqtt_proto

import (
	"bytes"
	"sync"
)

var bytePool = &sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	}}

func ResetBytePool(w *bytes.Buffer) {
	w.Reset()
	bytePool.Put(w)
}
