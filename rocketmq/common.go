package rocketmq

import (
	"fmt"
	"time"

	mq "github.com/aliyunmq/mq-http-go-sdk"
	"github.com/marsmay/golib/math2"
)

type ConnectConfig struct {
	Endpoint      string `toml:"endpoint" json:"endpoint"`
	AccessKey     string `toml:"access_key" json:"access_key"`
	SecretKey     string `toml:"secret_key" json:"secret_key"`
	SecurityToken string `toml:"security_token" json:"security_token"`
	InstanceId    string `toml:"instance_id" json:"instance_id"`
}

type Message struct {
	Tag     string
	Key     string
	Payload string
	Props   map[string]string
	Delay   time.Duration
}

func (m *Message) String() string {
	return fmt.Sprintf("%+v", *m)
}

func (m *Message) Request() mq.PublishMessageRequest {
	return mq.PublishMessageRequest{
		MessageBody:      m.Payload,
		MessageTag:       m.Tag,
		MessageKey:       m.Key,
		Properties:       m.Props,
		StartDeliverTime: math2.IIfInt64(m.Delay > 0, time.Now().Add(m.Delay).Unix()*1000, 0),
	}
}
