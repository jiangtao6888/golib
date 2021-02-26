package coder

import (
	"bytes"
	"encoding/json"

	"github.com/gin-gonic/gin"
)

const (
	EncodingJson    = "json"
	ContentTypeJSON = "application/json; charset=utf-8"
)

var JsonCoder = &jsonCoder{EscapeHTML: true, UseNumber: true}

type jsonCoder struct {
	EscapeHTML bool
	UseNumber  bool
}

func (c *jsonCoder) Marshal(v interface{}) (data []byte, err error) {
	if c.EscapeHTML {
		return json.Marshal(v)
	}

	bf := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(bf)
	jsonEncoder.SetEscapeHTML(c.EscapeHTML)

	if err = jsonEncoder.Encode(v); err != nil {
		return
	}

	data = bf.Bytes()
	return
}

func (c *jsonCoder) Unmarshal(data []byte, v interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	return decoder.Decode(v)
}

func (c *jsonCoder) DecodeRequest(ctx *gin.Context, v interface{}) (err error) {
	data, err := GetBody(ctx)

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
