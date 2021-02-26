package mqtt_proto

import (
	"bufio"
	"strconv"
)

type UnsubackPacket struct {
	FixedHeader
	MessageId uint16
}

func (p *UnsubackPacket) String() string {
	return "{" + p.FixedHeader.String() + " MessageId:" + strconv.Itoa(int(p.MessageId)) + "}"
}

func (p *UnsubackPacket) Encode() []byte {
	return []byte{176, 2, byte(p.MessageId >> 8), byte(p.MessageId)}
}

func (p *UnsubackPacket) Decode(r *bufio.Reader) (err error) {
	p.MessageId, err = DecodeUint16(r)
	return
}
