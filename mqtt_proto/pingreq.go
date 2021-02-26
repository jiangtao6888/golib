package mqtt_proto

import (
	"bufio"
	"bytes"
)

type PingreqPacket struct {
	FixedHeader
}

func (p *PingreqPacket) String() string {
	return p.FixedHeader.String()
}

func (p *PingreqPacket) Encode() []byte {
	w := bytePool.Get().(*bytes.Buffer)
	defer ResetBytePool(w)

	p.FixedHeader.Encode(w)
	b := make([]byte, w.Len())
	copy(b, w.Bytes())

	return b
}

func (p *PingreqPacket) Decode(_ *bufio.Reader) error {
	return nil
}
