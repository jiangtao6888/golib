package coder

import (
	"bytes"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/marsmay/golib/sync_pool"
)

const (
	EncodingJson    = "json"
	ContentTypeJSON = "application/json; charset=utf-8"
)

var JsonCoder = &jsonCoder{EscapeHTML: true, UseNumber: true, DisallowUnknownFields: false}

type jsonCoder struct {
	EscapeHTML            bool
	UseNumber             bool
	DisallowUnknownFields bool
}

func (c *jsonCoder) Marshal(v interface{}) (data []byte, err error) {
	if c.EscapeHTML {
		return json.Marshal(v)
	}

	w := sync_pool.BytePool.Get().(*bytes.Buffer)
	defer sync_pool.ResetBytePool(w)

	jsonEncoder := json.NewEncoder(w)
	jsonEncoder.SetEscapeHTML(c.EscapeHTML)

	if err = jsonEncoder.Encode(v); err != nil {
		return
	}

	data = make([]byte, w.Len())
	copy(data, w.Bytes())
	return
}

func (c *jsonCoder) Unmarshal(data []byte, v interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(data))

	if c.UseNumber {
		decoder.UseNumber()
	}

	if c.DisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}

	return decoder.Decode(v)
}

func (c *jsonCoder) DecodeRequest(ctx *gin.Context, v interface{}) (err error) {
	data, err := GetRequestBody(ctx)

	if err != nil {
		return
	}

	return c.Unmarshal(data, v)
}

func (c *jsonCoder) SendResponse(ctx *gin.Context, v interface{}) (err error) {
	ctx.Header(EncodingHeader, EncodingJson)
	ctx.Header(ContentTypeHeader, ContentTypeJSON)

	data, err := c.Marshal(v)

	if err != nil {
		return
	}

	_, err = ctx.Writer.Write(data)
	return
}
