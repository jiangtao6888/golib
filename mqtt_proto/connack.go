package mqtt_proto

import (
	"bufio"
	"bytes"
	"strconv"
)

type ConnackPacket struct {
	FixedHeader
	SessionPresent bool
	ReturnCode     byte
}

var ConnackBytes = []byte{32, 2, 0, 0}

func (p *ConnackPacket) String() string {
	return "{" + p.FixedHeader.String() + " SessionPresent:" +
		strconv.FormatBool(p.SessionPresent) + " ReturnCode:" + strconv.Itoa(int(p.ReturnCode)) + "}"
}

func (p *ConnackPacket) Encode() []byte {
	w := bytePool.Get().(*bytes.Buffer)
	defer ResetBytePool(w)

	p.RemainLen = 2
	p.FixedHeader.Encode(w)
	w.WriteByte(BoolToByte(p.SessionPresent))
	w.WriteByte(p.ReturnCode)

	b := make([]byte, w.Len())
	copy(b, w.Bytes())
	return b
}

func (p *ConnackPacket) Decode(r *bufio.Reader) error {
	i, err := r.ReadByte()
	if err != nil {
		return err
	}
	p.SessionPresent = i&0x01 > 0
	p.ReturnCode, err = r.ReadByte()
	return err
}
