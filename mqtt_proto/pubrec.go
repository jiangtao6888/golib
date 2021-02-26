package mqtt_proto

import (
	"bufio"
	"strconv"
)

type PubrecPacket struct {
	FixedHeader
	MessageId uint16
}

func (p *PubrecPacket) String() string {
	return "{" + p.FixedHeader.String() + " MessageId:" + strconv.Itoa(int(p.MessageId)) + "}"
}

func (p *PubrecPacket) Encode() []byte {
	return []byte{80, 2, byte(p.MessageId >> 8), byte(p.MessageId)}
}

func (p *PubrecPacket) Decode(r *bufio.Reader) (err error) {
	p.MessageId, err = DecodeUint16(r)
	return
}
