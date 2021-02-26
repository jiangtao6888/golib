package mqtt_proto

import (
	"bufio"
	"strconv"
)

type PubrelPacket struct {
	FixedHeader
	MessageId uint16
}

func (p *PubrelPacket) String() string {
	return "{" + p.FixedHeader.String() + " MessageId:" + strconv.Itoa(int(p.MessageId)) + "}"
}

func (p *PubrelPacket) Encode() []byte {
	return []byte{98, 2, byte(p.MessageId >> 8), byte(p.MessageId)}
}

func (p *PubrelPacket) Decode(r *bufio.Reader) (err error) {
	p.MessageId, err = DecodeUint16(r)
	return
}
