package mqtt_proto

import (
	"bufio"
)

type PingrespPacket struct {
	FixedHeader
}

func (p *PingrespPacket) String() string { return p.FixedHeader.String() }

func (p *PingrespPacket) Encode() []byte { return []byte{208, 0} }

func (p *PingrespPacket) Decode(_ *bufio.Reader) error { return nil }
