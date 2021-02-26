package mqtt_proto

import (
	"bufio"
	"strconv"
)

type PubackPacket struct {
	FixedHeader
	MessageId uint16
}

func (p *PubackPacket) String() string {
	return "{" + p.FixedHeader.String() + " MessageId:" + strconv.Itoa(int(p.MessageId)) + "}"
}

func (p *PubackPacket) Encode() []byte {
	return []byte{64, 2, byte(p.MessageId >> 8), byte(p.MessageId)}
}

func (p *PubackPacket) Decode(r *bufio.Reader) (err error) {
	p.MessageId, err = DecodeUint16(r)
	return
}
