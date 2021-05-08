package rocketmq

import (
	rocketmq "github.com/apache/rocketmq-client-go/core"
	"github.com/marsmay/golib/logger"
)

type ConsumerConfig struct {
	*ConnectConfig
	Topic   string `toml:"topic" json:"topic"`
	GroupId string `toml:"group_id" json:"group_id"`
	Tag     string `toml:"tag" json:"tag"`
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

func (c *Consumer) receive(msg *rocketmq.MessageExt) rocketmq.ConsumeStatus {
	c.logger.Debugf("rocketmq Consumer message received, MessageID:%s, Tags:%s, Body:%s", msg.MessageID, msg.Tags, msg.Body)

	if err := c.handler([]byte(msg.Body)); err != nil {
		c.logger.Errorf("handle rocketmq message failed | message: %+v | error: %s", msg, err)
		return rocketmq.ReConsumeLater
	}

	return rocketmq.ConsumeSuccess
}

func NewConsumer(conf *ConsumerConfig, handler func([]byte) error, logger *logger.Logger) (consumer *Consumer, err error) {
	c := &Consumer{
		conf:    conf,
		handler: handler,
		logger:  logger,
	}

	c.consumer, err = rocketmq.NewPushConsumer(&rocketmq.PushConsumerConfig{
		ClientConfig: rocketmq.ClientConfig{
			GroupID:    conf.GroupId,
			NameServer: conf.Endpoint,
			Credentials: &rocketmq.SessionCredentials{
				AccessKey: conf.AccessKey,
				SecretKey: conf.SecretKey,
				Channel:   conf.Channel,
			},
		},
		Model:         rocketmq.Clustering,
		ConsumerModel: rocketmq.Orderly,
	})

	if err != nil {
		return
	}

	if err = c.consumer.Subscribe(conf.Topic, conf.Tag, c.receive); err != nil {
		return
	}

	if err = c.consumer.Start(); err != nil {
		return
	}

	consumer = c
	return
}
