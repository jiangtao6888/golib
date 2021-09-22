package pool

import (
	"bufio"
	"io"
	"sync"
)

type bufReaderPool struct {
	*sync.Pool
}

func (p *bufReaderPool) Get(r io.Reader) *bufio.Reader {
	if v := p.Pool.Get(); v != nil {
		br := v.(*bufio.Reader)
		br.Reset(r)

		return br
	}

	return bufio.NewReader(r)
}

func (p *bufReaderPool) Put(br *bufio.Reader) {
	if br != nil {
		br.Reset(nil)
		p.Pool.Put(br)
	}
}

var BufReader = &bufReaderPool{&sync.Pool{}}
