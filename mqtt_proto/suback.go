package mqtt_proto

import (
	"bufio"
	"bytes"
	"io"
	"strconv"
)

type SubackPacket struct {
	FixedHeader
	MessageId   uint16
	GrantedQoss []byte
}

func (p *SubackPacket) String() string {
	return "{" + p.FixedHeader.String() + " MessageId:" + strconv.Itoa(int(p.MessageId)) + "}"
}

func (p *SubackPacket) Encode() []byte {
	w := bytePool.Get().(*bytes.Buffer)
	defer ResetBytePool(w)

	p.RemainLen = 2 + len(p.GrantedQoss)
	p.FixedHeader.Encode(w)

	w.Write([]byte{
		byte(p.MessageId >> 8),
		byte(p.MessageId),
	})
	w.Write(p.GrantedQoss)

	b := make([]byte, w.Len())
	copy(b, w.Bytes())
	return b
}

func (p *SubackPacket) Decode(r *bufio.Reader) (err error) {
	length := p.RemainLen - 2

	if p.MessageId, err = DecodeUint16(r); err != nil {
		return
	}

	p.GrantedQoss = make([]byte, length)
	_, err = io.ReadFull(r, p.GrantedQoss)
	return
}
