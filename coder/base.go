package coder

import (
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

func GetBody(ctx *gin.Context) (body []byte, err error) {
	if cb, ok := ctx.Get(gin.BodyBytesKey); ok {
		if cbb, ok := cb.([]byte); ok {
			body = cbb
		}
	}

	if body == nil {
		body, err = ioutil.ReadAll(ctx.Request.Body)

		if err != nil {
			return
		}

		ctx.Set(gin.BodyBytesKey, body)
	}

	return
}
