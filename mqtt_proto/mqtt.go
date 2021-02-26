package mqtt_proto

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strconv"
)

// mqtt的协议解析和封装
// https://github.com/mcxiaoke/mqtt
// https://mcxiaoke.gitbooks.io/mqtt-cn/content/mqtt/02-ControlPacketFormat.html
var ErrBadData = errors.New("bad data")

type ControlPacket interface {
	String() string
	Encode() []byte
	Decode(*bufio.Reader) error
}

//根据类型
func NewControlPacket(t byte) ControlPacket {
	fh := FixedHeader{Type: t}
	if t == PUBREL || t == SUBSCRIBE || t == UNSUBSCRIBE {
		fh.Qos = 1
	}
	return NewControlPacketWithHeader(fh)
}

//根据公共消息头
func NewControlPacketWithHeader(fh FixedHeader) ControlPacket {
	switch fh.Type {
	case CONNECT:
		return &ConnectPacket{FixedHeader: fh}
	case CONNACK:
		return &ConnackPacket{FixedHeader: fh}
	case PUBLISH:
		return &PublishPacket{FixedHeader: fh}
	case PUBACK:
		return &PubackPacket{FixedHeader: fh}
	case PUBREC:
		return &PubrecPacket{FixedHeader: fh}
	case PUBREL:
		return &PubrelPacket{FixedHeader: fh}
	case PUBCOMP:
		return &PubcompPacket{FixedHeader: fh}
	case SUBSCRIBE:
		return &SubscribePacket{FixedHeader: fh}
	case SUBACK:
		return &SubackPacket{FixedHeader: fh}
	case UNSUBSCRIBE:
		return &UnsubscribePacket{FixedHeader: fh}
	case UNSUBACK:
		return &UnsubackPacket{FixedHeader: fh}
	case PINGREQ:
		return &PingreqPacket{FixedHeader: fh}
	case PINGRESP:
		return &PingrespPacket{FixedHeader: fh}
	case DISCONNECT:
		return &DisconnectPacket{FixedHeader: fh}
	default:
		return nil
	}
}

const (
	CONNECT     = 1
	CONNACK     = 2
	PUBLISH     = 3
	PUBACK      = 4
	PUBREC      = 5
	PUBREL      = 6
	PUBCOMP     = 7
	SUBSCRIBE   = 8
	SUBACK      = 9
	UNSUBSCRIBE = 10
	UNSUBACK    = 11
	PINGREQ     = 12
	PINGRESP    = 13
	DISCONNECT  = 14
)

var PacketNames = map[uint8]string{
	1:  "CONNECT",
	2:  "CONNACK",
	3:  "PUBLISH",
	4:  "PUBACK",
	5:  "PUBREC",
	6:  "PUBREL",
	7:  "PUBCOMP",
	8:  "SUBSCRIBE",
	9:  "SUBACK",
	10: "UNSUBSCRIBE",
	11: "UNSUBACK",
	12: "PINGREQ",
	13: "PINGRESP",
	14: "DISCONNECT",
}

func BoolToByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}

//MQTT的消息头
type FixedHeader struct {
	Type      byte
	Dup       bool
	Qos       byte
	Retain    bool
	RemainLen int
}

func (fh *FixedHeader) String() string {
	return "{Type:" + PacketNames[fh.Type] + " Qos:" + strconv.Itoa(int(fh.Qos)) +
		" RemainLen:" + strconv.Itoa(fh.RemainLen) + "}"
}

func (fh *FixedHeader) First() byte {
	return fh.Type<<4 | BoolToByte(fh.Dup)<<3 | fh.Qos<<1 | BoolToByte(fh.Retain)
}

func (fh *FixedHeader) Encode(w *bytes.Buffer) {
	w.WriteByte(fh.First())
	length := fh.RemainLen

	for {
		digit := byte(length % 128)
		length /= 128
		if length > 0 {
			digit |= 0x80
		}
		w.WriteByte(digit)
		if length == 0 {
			break
		}
	}
}

//解析剩余长度
func (fh *FixedHeader) DecodeRemainLen(r *bufio.Reader) error {
	var length, idx uint32

	for {
		digit, err := r.ReadByte()
		if err != nil {
			return err
		}

		length += uint32(digit&0x7f) << idx
		if (digit & 0x80) == 0 {
			break
		}
		idx += 7
	}

	fh.RemainLen = int(length)

	return nil
}

func ReadPacket(r *bufio.Reader) (cp ControlPacket, err error) {
	first, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	fh := FixedHeader{
		Type:   first >> 4,
		Dup:    (first>>3)&0x01 > 0,
		Qos:    (first >> 1) & 0x03,
		Retain: first&0x01 > 0,
	}

	if err := fh.DecodeRemainLen(r); err != nil {
		return nil, err
	}

	cp = NewControlPacketWithHeader(fh)

	if cp == nil {
		return nil, ErrBadData
	}

	if fh.RemainLen == 0 {
		return cp, nil
	}

	if err := cp.Decode(r); err != nil {
		return nil, err
	}
	return cp, nil
}

func DecodeString(r *bufio.Reader) (string, error) {
	l, err := DecodeUint16(r)
	if err != nil {
		return "", err
	}

	b := make([]byte, l)
	_, err = io.ReadFull(r, b)
	return string(b), err
}

func DecodeBytes(r *bufio.Reader) ([]byte, error) {
	l, err := DecodeUint16(r)
	if err != nil {
		return nil, err
	}
	b := make([]byte, l)
	_, err = io.ReadFull(r, b)
	return b, err
}

func DecodeUint16(r *bufio.Reader) (uint16, error) {
	a, err := r.ReadByte()
	if err != nil {
		return 0, err
	}

	b, err := r.ReadByte()
	if err != nil {
		return 0, err
	}
	return uint16(a)<<8 | uint16(b), nil
}
