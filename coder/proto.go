package coder

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/gogo/protobuf/proto"
)

const (
	EncodingProtobuf    = "protobuf"
	ContentTypeProtobuf = "application/x-protobuf"
)

var ProtoCoder = &protoCoder{}

type protoCoder struct{}

func (c *protoCoder) Unmarshal(data []byte, v interface{}) error {
	pb, ok := v.(proto.Message)

	if !ok {
		return errors.New("invalid protobuf message")
	}

	return proto.Unmarshal(data, pb)
}

func (c *protoCoder) Marshal(v interface{}) ([]byte, error) {
	pb, ok := v.(proto.Message)

	if !ok {
		return nil, errors.New("invalid protobuf message")
	}

	return proto.Marshal(pb)
}

func (c *protoCoder) DecodeRequest(ctx *gin.Context, v interface{}) (err error) {
	data, err := GetBody(ctx)

	if err != nil {
		return
	}

	return c.Unmarshal(data, v)
}

func (c *protoCoder) SendResponse(ctx *gin.Context, v interface{}) (err error) {
	ctx.Header(EncodingHeader, EncodingProtobuf)
	ctx.Header(ContentTypeHeader, ContentTypeProtobuf)

	data, err := c.Marshal(v)

	if err != nil {
		return
	}

	_, err = ctx.Writer.Write(data)
	return
}
