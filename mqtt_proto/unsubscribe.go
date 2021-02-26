package mqtt_proto

import (
	"bufio"
	"bytes"
	"strconv"
	"strings"
)

type UnsubscribePacket struct {
	FixedHeader
	MessageId uint16
	Topics    []string
}

func (p *UnsubscribePacket) String() string {
	return "{" + p.FixedHeader.String() + " MessageId:" + strconv.Itoa(int(p.MessageId)) +
		" Topics:" + strings.Join(p.Topics, "/") + "}"
}

func (p *UnsubscribePacket) Encode() []byte {
	w := bytePool.Get().(*bytes.Buffer)
	defer ResetBytePool(w)

	p.RemainLen = 2
	for _, topic := range p.Topics {
		p.RemainLen += 2
		p.RemainLen += len(topic)
	}

	p.FixedHeader.Encode(w)
	w.Write([]byte{
		byte(p.MessageId >> 8),
		byte(p.MessageId),
	})

	for _, topic := range p.Topics {
		w.Write([]byte{
			byte(len(topic) >> 8),
			byte(len(topic)),
		})
		w.WriteString(topic)
	}

	b := make([]byte, w.Len())
	copy(b, w.Bytes())
	return b
}

func (p *UnsubscribePacket) Decode(r *bufio.Reader) (err error) {
	if p.MessageId, err = DecodeUint16(r); err != nil {
		return
	}

	length := p.RemainLen - 2
	for length > 2 {
		topic, err := DecodeString(r)
		if err != nil {
			return err
		}
		p.Topics = append(p.Topics, topic)
		length -= len(topic)
		length -= 2
	}

	if length != 0 {
		return ErrBadData
	}

	return nil
}
