package pool

import (
	"bytes"
	"sync"
)

type bufBytesPool struct {
	*sync.Pool
}

func (p *bufBytesPool) Get() *bytes.Buffer {
	if v := p.Pool.Get(); v != nil {
		br := v.(*bytes.Buffer)
		br.Reset()

		return br
	}

	return new(bytes.Buffer)
}

func (p *bufBytesPool) Put(bf *bytes.Buffer) {
	if bf != nil {
		bf.Reset()
		p.Pool.Put(bf)
	}
}

var BufBytes = &bufBytesPool{&sync.Pool{}}
