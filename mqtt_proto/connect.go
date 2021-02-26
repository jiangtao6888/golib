package mqtt_proto

import (
	"bufio"
	"bytes"
	"strconv"
)

type ConnectPacket struct {
	FixedHeader
	ProtocolName    string
	ProtocolVersion byte
	CleanSession    bool
	WillFlag        bool
	WillQos         byte
	WillRetain      bool
	UsernameFlag    bool
	PasswordFlag    bool
	ReservedBit     byte
	Keepalive       uint16
	Key             string
	WillTopic       string
	WillMessage     []byte
	Username        string
	Password        []byte
}

func (p *ConnectPacket) String() string {
	s := "{" + p.FixedHeader.String()
	s += " ProtocolName:" + p.ProtocolName
	s += " ProtocolVersion:" + strconv.Itoa(int(p.ProtocolVersion))
	s += " CleanSession:" + strconv.FormatBool(p.CleanSession)
	s += " WillFlag:" + strconv.FormatBool(p.WillFlag)
	s += " WillQos:" + strconv.Itoa(int(p.WillQos))
	s += " WillRetain:" + strconv.FormatBool(p.WillRetain)
	s += " UsernameFlag:" + strconv.FormatBool(p.UsernameFlag)
	s += " PasswordFlag:" + strconv.FormatBool(p.PasswordFlag)
	s += " ReservedBit:" + strconv.Itoa(int(p.ReservedBit))
	s += " Keepalive:" + strconv.Itoa(int(p.Keepalive))
	s += " Key:" + p.Key
	s += " WillTopic:" + p.WillTopic
	s += " WillMessage:" + string(p.WillMessage)
	s += " Username:" + p.Username
	s += " Password:" + string(p.Password) + "}"
	return s
}

func (p *ConnectPacket) Encode() []byte {
	w := bytePool.Get().(*bytes.Buffer)
	defer ResetBytePool(w)

	length := 8 + len(p.ProtocolName) + len(p.Key)

	if p.WillFlag {
		length += 4 + len(p.WillTopic) + len(p.WillMessage)
	}

	if p.UsernameFlag {
		length += 2 + len(p.Username)
	}

	if p.PasswordFlag {
		length += 2 + len(p.Password)
	}

	p.RemainLen = length

	p.FixedHeader.Encode(w)

	w.Write([]byte{
		byte(len(p.ProtocolName) >> 8),
		byte(len(p.ProtocolName)),
	})
	w.WriteString(p.ProtocolName)

	w.WriteByte(p.ProtocolVersion)

	w.WriteByte(BoolToByte(p.CleanSession)<<1 |
		BoolToByte(p.WillFlag)<<2 |
		p.WillQos<<3 |
		BoolToByte(p.WillRetain)<<5 |
		BoolToByte(p.PasswordFlag)<<6 |
		BoolToByte(p.UsernameFlag)<<7)

	w.Write([]byte{
		byte(p.Keepalive >> 8),
		byte(p.Keepalive),
	})

	w.Write([]byte{
		byte(len(p.Key) >> 8),
		byte(len(p.Key)),
	})

	w.WriteString(p.Key)

	if p.WillFlag {
		w.Write([]byte{
			byte(len(p.WillTopic) >> 8),
			byte(len(p.WillTopic)),
		})
		w.WriteString(p.WillTopic)

		w.Write([]byte{
			byte(len(p.WillMessage) >> 8),
			byte(len(p.WillMessage)),
		})
		w.Write(p.WillMessage)
	}

	if p.UsernameFlag {
		w.Write([]byte{
			byte(len(p.Username) >> 8),
			byte(len(p.Username)),
		})
		w.WriteString(p.Username)
	}

	if p.PasswordFlag {
		w.Write([]byte{
			byte(len(p.Password) >> 8),
			byte(len(p.Password)),
		})
		w.Write(p.Password)
	}

	b := make([]byte, w.Len())
	copy(b, w.Bytes())
	return b
}

func (p *ConnectPacket) Decode(r *bufio.Reader) (err error) {
	if p.ProtocolName, err = DecodeString(r); err != nil {
		return
	}

	if p.ProtocolVersion, err = r.ReadByte(); err != nil {
		return
	}

	options, err := r.ReadByte()
	if err != nil {
		return
	}

	p.ReservedBit = options & 0x01
	p.CleanSession = (options>>1)&0x01 > 0
	p.WillFlag = (options>>2)&0x01 > 0
	p.WillQos = (options >> 3) & 0x03
	p.WillRetain = (options>>5)&0x01 > 0
	p.PasswordFlag = (options>>6)&0x01 > 0
	p.UsernameFlag = (options>>7)&0x01 > 0

	if p.Keepalive, err = DecodeUint16(r); err != nil {
		return
	}

	if p.Key, err = DecodeString(r); err != nil {
		return
	}

	if p.WillFlag {
		if p.WillTopic, err = DecodeString(r); err != nil {
			return
		}
		if p.WillMessage, err = DecodeBytes(r); err != nil {
			return
		}
	}

	if p.UsernameFlag {
		if p.Username, err = DecodeString(r); err != nil {
			return
		}
	}

	if p.PasswordFlag {
		if p.Password, err = DecodeBytes(r); err != nil {
			return
		}
	}
	return nil
}
