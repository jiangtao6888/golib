package coder

import (
	"bytes"
	"encoding/json"
	"github.com/kataras/iris/v12/context"
)

var JsonCoder = &jsonCoder{escapeHTML: true}

type jsonCoder struct {
	escapeHTML bool
}

func (c *jsonCoder) Unmarshal(data []byte, v interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	return decoder.Decode(v)
}

func (c *jsonCoder) Marshal(v interface{}) ([]byte, error) {
	if c.escapeHTML {
		return json.Marshal(v)
	}

	bf := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(bf)
	jsonEncoder.SetEscapeHTML(c.escapeHTML)

	if err := jsonEncoder.Encode(v); err != nil {
		return nil, err
	}

	return bf.Bytes(), nil
}

func (c *jsonCoder) DecodeIrisReq(ctx context.Context, v interface{}) error {
	return ctx.UnmarshalBody(v, c)
}

func (c *jsonCoder) SendIrisReply(ctx context.Context, v interface{}) error {
	option := context.JSON{UnescapeHTML: !c.escapeHTML}

	ctx.ContentType(context.ContentJSONHeaderValue)
	ctx.Header(EncodingHeader, EncodingJson)

	_, err := ctx.JSON(v, option)
	return err
}
