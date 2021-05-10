package rocketmq

import (
	"context"
	"errors"
	"fmt"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	oProducer "github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/apache/rocketmq-client-go/v2/rlog"
	"github.com/marsmay/golib/logger"
)

type Message struct {
	Topic       string   `json:"topic"`
	Payload     []byte   `json:"payload"`
	Tag         string   `json:"tag"`
	Keys        []string `json:"keys"`
	ShardingKey string   `json:"sharding_key"`
	DelayLevel  int      `json:"delay_level"`
}

func (m *Message) String() string {
	return fmt.Sprintf("%+v", *m)
}

func (m *Message) Request() *primitive.Message {
	msg := primitive.NewMessage(m.Topic, m.Payload)

	if m.Tag != "" {
		msg.WithTag(m.Tag)
	}

	if len(m.Keys) > 0 {
		msg.WithKeys(m.Keys)
	}

	if m.ShardingKey != "" {
		msg.WithShardingKey(m.ShardingKey)
	}

	if m.DelayLevel > 0 {
		msg.WithDelayTimeLevel(m.DelayLevel)
	}

	return msg
}

type ProducerConfig struct {
	*ConnectConfig
	Group      string `toml:"group" json:"group"`
	RetryTimes int    `toml:"retry_times" json:"retry_times"`
}

type Producer struct {
	conf     *ProducerConfig
	producer rocketmq.Producer
	logger   *logger.Logger
	ctx      context.Context
	cancel   context.CancelFunc
}

func (c *Producer) closed() bool {
	select {
	case <-c.ctx.Done():
		return true
	default:
		return false
	}
}

func (c *Producer) Stop() {
	c.cancel()

	if err := c.producer.Shutdown(); err != nil {
		c.logger.Errorf("stop rocketmq producer failed | error: %s", err)
	}
}

func (c *Producer) SendSync(msg *Message) (res *primitive.SendResult, err error) {
	if c.closed() {
		err = errors.New("producer is stoped")
		return
	}

	return c.producer.SendSync(c.ctx, msg.Request())
}

func (c *Producer) SendAsync(msg *Message) (err error) {
	if c.closed() {
		err = errors.New("producer is stoped")
		return
	}

	err = c.producer.SendAsync(c.ctx, func(_ context.Context, res *primitive.SendResult, e error) {
		if e != nil {
			c.logger.Errorf("send rocketmq message failed | message: %+v | result: %s | error: %s", msg, res, e)
		}

		c.logger.Debugf("send rocketmq message | message: %+v | result: %s", msg, res)
	}, msg.Request())
	return
}

func NewProducer(conf *ProducerConfig, logger *logger.Logger) (producer *Producer, err error) {
	opts := []oProducer.Option{
		oProducer.WithNameServer(conf.Endpoints),
	}

	if conf.Group != "" {
		opts = append(opts, oProducer.WithGroupName(conf.Group))
	}

	if conf.AccessKey != "" && conf.SecretKey != "" {
		opts = append(opts, oProducer.WithCredentials(primitive.Credentials{
			AccessKey:     conf.AccessKey,
			SecretKey:     conf.SecretKey,
			SecurityToken: conf.SecurityToken,
		}))
	}

	if conf.RetryTimes > 0 {
		opts = append(opts, oProducer.WithRetry(conf.RetryTimes))
	}

	rlog.SetLogger(&Logger{l: logger})

	p, err := rocketmq.NewProducer(opts...)

	if err != nil {
		return
	}

	if err = p.Start(); err != nil {
		return
	}

	producer = &Producer{conf: conf, logger: logger, producer: p}
	producer.ctx, producer.cancel = context.WithCancel(context.Background())
	return
}
