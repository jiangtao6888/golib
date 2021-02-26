package mqtt_proto

import (
	"bufio"
	"bytes"
)

type DisconnectPacket struct {
	FixedHeader
}

func (p *DisconnectPacket) String() string {
	return p.FixedHeader.String()
}

func (p *DisconnectPacket) Encode() []byte {
	w := bytePool.Get().(*bytes.Buffer)
	defer ResetBytePool(w)

	p.FixedHeader.Encode(w)
	b := make([]byte, w.Len())
	copy(b, w.Bytes())

	return b
}

func (p *DisconnectPacket) Decode(_ *bufio.Reader) error {
	return nil
}
