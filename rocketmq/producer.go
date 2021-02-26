package rocketmq

import (
	"context"
	"errors"
	"sync"
	"time"

	mq "github.com/aliyunmq/mq-http-go-sdk"
	"github.com/marsmay/golib/logger"
)

type ProducerConfig struct {
	*ConnectConfig
	Topic string `toml:"topic" json:"topic"`
}

type Producer struct {
	conf     *ProducerConfig
	queue    chan *Message
	producer mq.MQProducer
	logger   *logger.Logger
	ctx      context.Context
	cancel   context.CancelFunc
	wg       *sync.WaitGroup
}

func (c *Producer) run() {
	c.wg.Add(1)
	go c.send()
	return
}

func (c *Producer) Stop() {
	c.cancel()
	close(c.queue)
	c.wg.Wait()
}

func (c *Producer) closed() bool {
	select {
	case <-c.ctx.Done():
		return true
	default:
		return false
	}
}

func (c *Producer) Send(msg *Message) error {
	if c.closed() {
		return errors.New("producer is stoped")
	}

	c.queue <- msg
	return nil
}

func (c *Producer) send() {
	defer c.wg.Done()

	for msg := range c.queue {
		res, err := c.producer.PublishMessage(msg.Request())

		if err != nil {
			c.logger.Errorf("publish rocketmq message failed | topic: %s | message: %+v | error: %s", c.conf.Topic, msg, err)

			if err := c.Send(msg); err != nil {
				c.logger.Errorf("publish rocketmq retry message failed | topic: %s | message: %+v | error: %s", c.conf.Topic, msg, err)
			}

			time.Sleep(time.Millisecond * 100)
			continue
		}

		c.logger.Debugf("send rocketmq message | topic: %s | message: %+v | messageId: %s | messageMD5: %s", c.conf.Topic, msg, res.MessageId, res.MessageBodyMD5)
	}
}

func NewProducer(conf *ProducerConfig, logger *logger.Logger) *Producer {
	producer := &Producer{
		conf:   conf,
		queue:  make(chan *Message, 4096),
		logger: logger,
		wg:     &sync.WaitGroup{},
	}

	client := mq.NewAliyunMQClient(conf.Endpoint, conf.AccessKey, conf.SecretKey, conf.SecurityToken)
	producer.producer = client.GetProducer(conf.InstanceId, conf.Topic)

	producer.ctx, producer.cancel = context.WithCancel(context.Background())
	producer.run()

	return producer
}
