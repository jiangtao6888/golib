package rocketmq

import (
	"context"
	"strings"

	"github.com/apache/rocketmq-client-go/v2"
	oConsumer "github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/marsmay/golib/logger"
)

type ConsumerConfig struct {
	*ConnectConfig
	Topic   string   `toml:"topic" json:"topic"`
	GroupId string   `toml:"group_id" json:"group_id"`
	Tags    []string `toml:"tags" json:"tags"`
}

type Consumer struct {
	conf     *ConsumerConfig
	handler  func([]byte) error
	consumer rocketmq.PushConsumer
	logger   *logger.Logger
}

func (c *Consumer) Stop() {
	if err := c.consumer.Shutdown(); err != nil {
		c.logger.Errorf("rocketmq consumer shutdown failed | error: %s", err)
		return
	}
}

func (c *Consumer) receive(_ context.Context, msgs ...*primitive.MessageExt) (res oConsumer.ConsumeResult, err error) {
	for _, msg := range msgs {
		c.logger.Debugf("receive rocketmq message | message: %s", msg)

		if e := c.handler(msg.Body); e != nil {
			c.logger.Errorf("handle rocketmq message failed | message: %s | error: %s", msg, err)
		}
	}

	return oConsumer.ConsumeSuccess, nil
}

func NewConsumer(conf *ConsumerConfig, handler func([]byte) error, logger *logger.Logger) (consumer *Consumer, err error) {
	opts := []oConsumer.Option{
		oConsumer.WithNameServer(conf.Endpoints),
		oConsumer.WithConsumerModel(oConsumer.Clustering),
		oConsumer.WithConsumerOrder(true),
	}

	if conf.GroupId != "" {
		opts = append(opts, oConsumer.WithGroupName(conf.GroupId))
	}

	if conf.AccessKey != "" && conf.SecretKey != "" {
		opts = append(opts, oConsumer.WithCredentials(primitive.Credentials{
			AccessKey:     conf.AccessKey,
			SecretKey:     conf.SecretKey,
			SecurityToken: conf.SecurityToken,
		}))
	}

	selector := oConsumer.MessageSelector{}

	if len(conf.Tags) > 0 {
		selector = oConsumer.MessageSelector{
			Type:       oConsumer.TAG,
			Expression: strings.Join(conf.Tags, " || "),
		}
	}

	c := &Consumer{
		conf:    conf,
		handler: handler,
		logger:  logger,
	}

	if c.consumer, err = rocketmq.NewPushConsumer(opts...); err != nil {
		return
	}

	if err = c.consumer.Subscribe(conf.Topic, selector, c.receive); err != nil {
		return
	}

	if err = c.consumer.Start(); err != nil {
		return
	}

	consumer = c
	return
}
