package mqtt_proto

import (
	"bufio"
	"bytes"
	"strconv"
	"strings"
)

type SubscribePacket struct {
	FixedHeader
	MessageId uint16
	Topics    []string
	Qoss      []byte
}

func (p *SubscribePacket) String() string {
	return "{" + p.FixedHeader.String() + " MessageId:" + strconv.Itoa(int(p.MessageId)) +
		" Topics:" + strings.Join(p.Topics, "/") + "}"
}

func (p *SubscribePacket) Encode() []byte {
	w := bytePool.Get().(*bytes.Buffer)
	defer ResetBytePool(w)

	p.RemainLen = 2 + len(p.Qoss)
	for _, topic := range p.Topics {
		p.RemainLen += 2
		p.RemainLen += len(topic)
	}

	p.FixedHeader.Encode(w)
	w.Write([]byte{
		byte(p.MessageId >> 8),
		byte(p.MessageId),
	})

	for i, topic := range p.Topics {
		w.Write([]byte{
			byte(len(topic) >> 8),
			byte(len(topic)),
		})
		w.WriteString(topic)
		w.WriteByte(p.Qoss[i])
	}

	b := make([]byte, w.Len())
	copy(b, w.Bytes())
	return b
}

func (p *SubscribePacket) Decode(r *bufio.Reader) (err error) {
	length := p.RemainLen

	if p.MessageId, err = DecodeUint16(r); err != nil {
		return
	}

	length -= 2
	for length > 0 {
		topic, err := DecodeString(r)
		if err != nil {
			return err
		}
		p.Topics = append(p.Topics, topic)

		qos, err := r.ReadByte()
		if err != nil {
			return err
		}
		p.Qoss = append(p.Qoss, qos)
		length -= 3
		length -= len(topic)
	}
	return nil
}
