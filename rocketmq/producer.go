package rocketmq

import (
	"context"
	"errors"
	"sync"
	"time"

	rocketmq "github.com/apache/rocketmq-client-go/core"
	"github.com/marsmay/golib/logger"
)

type ProducerConfig struct {
	*ConnectConfig
	Topic   string `toml:"topic" json:"topic"`
	GroupId string `toml:"group_id" json:"group_id"`
}

type Producer struct {
	conf     *ProducerConfig
	queue    chan *Message
	producer rocketmq.Producer
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

func (c *Producer) Send(msg *Message) (err error) {
	if c.closed() {
		err = errors.New("producer is stoped")
		return
	}

	c.queue <- msg
	return
}

func (c *Producer) send() {
	defer c.wg.Done()

	for msg := range c.queue {
		res, err := c.producer.SendMessageOrderlyByShardingKey(msg.Request(c.conf.Topic), msg.ShardingKey)

		if err != nil {
			c.logger.Errorf("publish rocketmq message failed | topic: %s | message: %+v | error: %s", c.conf.Topic, msg, err)

			if err := c.Send(msg); err != nil {
				c.logger.Errorf("publish rocketmq retry message failed | topic: %s | message: %+v | error: %s", c.conf.Topic, msg, err)
			}

			time.Sleep(time.Millisecond * 100)
			continue
		}

		c.logger.Debugf("send rocketmq message | topic: %s | message: %+v | result: %s", c.conf.Topic, msg, res.String())
	}

	if err := c.producer.Shutdown(); err != nil {
		c.logger.Errorf("stop rocketmq producer failed | error: %s", err)
	}
}

func (c *Producer) SyncSend(msg *Message) (res *rocketmq.SendResult, err error) {
	return c.producer.SendMessageOrderlyByShardingKey(msg.Request(c.conf.Topic), msg.ShardingKey)
}

func NewProducer(conf *ProducerConfig, logger *logger.Logger) (producer *Producer, err error) {
	p := &Producer{
		conf:   conf,
		queue:  make(chan *Message, 4096),
		logger: logger,
		wg:     &sync.WaitGroup{},
	}

	p.producer, err = rocketmq.NewProducer(&rocketmq.ProducerConfig{
		ClientConfig: rocketmq.ClientConfig{
			GroupID:    conf.GroupId,
			NameServer: conf.Endpoint,
			Credentials: &rocketmq.SessionCredentials{
				AccessKey: conf.AccessKey,
				SecretKey: conf.SecretKey,
				Channel:   conf.Channel,
			},
		},
		ProducerModel: rocketmq.OrderlyProducer,
	})

	if err != nil {
		return
	}

	if err = p.producer.Start(); err != nil {
		return
	}

	producer = p
	producer.ctx, producer.cancel = context.WithCancel(context.Background())
	producer.run()

	return
}
