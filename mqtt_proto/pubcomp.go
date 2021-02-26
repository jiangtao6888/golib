package mqtt_proto

import (
	"bufio"
	"strconv"
)

type PubcompPacket struct {
	FixedHeader
	MessageId uint16
}

func (p *PubcompPacket) String() string {
	return "{" + p.FixedHeader.String() + " MessageId:" + strconv.Itoa(int(p.MessageId)) + "}"
}

func (p *PubcompPacket) Encode() []byte {
	return []byte{112, 2, byte(p.MessageId >> 8), byte(p.MessageId)}
}

func (p *PubcompPacket) Decode(r *bufio.Reader) (err error) {
	p.MessageId, err = DecodeUint16(r)
	return
}
