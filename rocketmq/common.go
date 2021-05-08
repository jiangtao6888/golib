package rocketmq

import (
	"fmt"

	rocketmq "github.com/apache/rocketmq-client-go/core"
)

type ConnectConfig struct {
	Endpoint  string `toml:"endpoint" json:"endpoint"`
	AccessKey string `toml:"access_key" json:"access_key"`
	SecretKey string `toml:"secret_key" json:"secret_key"`
	Channel   string `toml:"channel" json:"channel"`
}

type Message struct {
	Tag         string `json:"tag"`
	Key         string `json:"key"`
	Payload     string `json:"payload"`
	ShardingKey string `json:"sharding_key"`
}

func (m *Message) String() string {
	return fmt.Sprintf("%+v", *m)
}

func (m *Message) Request(topic string) *rocketmq.Message {
	return &rocketmq.Message{
		Topic: topic,
		Body:  m.Payload,
		Keys:  m.Key,
		Tags:  m.Tag,
	}
}
