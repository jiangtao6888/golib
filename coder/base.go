package coder

import (
	"bytes"
	"io/ioutil"

	"github.com/gin-gonic/gin"
)

const (
	EncodingHeader    = "Protocol-Encoding"
	ContentTypeHeader = "Content-Type"
)

type ICoder interface {
	Unmarshal(data []byte, v interface{}) error
	Marshal(v interface{}) ([]byte, error)
	DecodeRequest(ctx *gin.Context, v interface{}) error
	SendResponse(ctx *gin.Context, v interface{}) error
}

func GetRequestBody(ctx *gin.Context) (body []byte, err error) {
	if b, ok := ctx.Get(gin.BodyBytesKey); ok {
		if bs, ok := b.([]byte); ok {
			body = bs
			return
		}
	}

	if ctx.Request.Body == nil {
		return
	}

	body, err = ioutil.ReadAll(ctx.Request.Body)

	if err != nil {
		return
	}

	ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	ctx.Set(gin.BodyBytesKey, body)
	return
}
