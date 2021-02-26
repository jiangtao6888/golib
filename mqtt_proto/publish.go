package mqtt_proto

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
)

type PublishPacket struct {
	FixedHeader
	TopicName string
	MessageId uint16
	Payload   []byte
}

// 超过这个长度启用gzip压缩
var gzipLen = 512

func (p *PublishPacket) String() string {
	return "{" + p.FixedHeader.String() + " TopicName:" + p.TopicName +
		" MessageId:" + strconv.Itoa(int(p.MessageId)) + " Payload:" + string(p.Payload) + "}"
}

func (p *PublishPacket) Encode() []byte {
	w := bytePool.Get().(*bytes.Buffer)
	defer ResetBytePool(w)

	payload := g(p.Payload)

	if len(p.Payload) > gzipLen {
		ss := strings.Split(p.TopicName, "/")
		ss[len(ss)-1] = "gzip"
		p.TopicName = strings.Join(ss, "/")
	}

	p.RemainLen = len(payload) + 2 + len(p.TopicName)

	if p.Qos > 0 {
		p.RemainLen += 2
	}

	p.FixedHeader.Encode(w)

	w.Write([]byte{
		byte(len(p.TopicName) >> 8),
		byte(len(p.TopicName)),
	})
	w.WriteString(p.TopicName)

	if p.Qos > 0 {
		w.Write([]byte{
			byte(p.MessageId >> 8),
			byte(p.MessageId),
		})
	}
	w.Write(payload)

	b := make([]byte, w.Len())
	copy(b, w.Bytes())
	return b
}

func (p *PublishPacket) Decode(r *bufio.Reader) (err error) {
	length := p.RemainLen

	if p.TopicName, err = DecodeString(r); err != nil {
		return
	}

	length -= 2
	length -= len(p.TopicName)

	if p.Qos > 0 {
		if p.MessageId, err = DecodeUint16(r); err != nil {
			return
		}
		length -= 2
	}

	p.Payload = make([]byte, length)
	_, err = io.ReadFull(r, p.Payload)

	if strings.HasSuffix(p.TopicName, "/gzip") {
		p.Payload = ug(p.Payload)
	}

	return
}

func g(b []byte) []byte {
	if len(b) <= gzipLen {
		return b
	}

	w := bytePool.Get().(*bytes.Buffer)
	defer ResetBytePool(w)
	gw := gzip.NewWriter(w)
	_, _ = gw.Write(b)
	_ = gw.Close()
	ret := make([]byte, w.Len())
	copy(ret, w.Bytes())
	return ret
}

func ug(b []byte) []byte {
	r, _ := gzip.NewReader(bytes.NewReader(b))
	defer func() { _ = r.Close() }()
	ret, _ := ioutil.ReadAll(r)
	return ret
}
