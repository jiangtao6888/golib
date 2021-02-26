package rocketmq

import (
	"context"
	"strings"
	"sync"

	mq "github.com/aliyunmq/mq-http-go-sdk"
	"github.com/gogap/errors"
	"github.com/marsmay/golib/logger"
)

type ConsumerConfig struct {
	*ConnectConfig
	Topic          string `toml:"topic" json:"topic"`
	GroupId        string `toml:"group_id" json:"group_id"`
	Tag            string `toml:"tag" json:"tag"`
	Worker         int    `toml:"worker" json:"worker"`
	MsgNumOnce     int32  `toml:"msg_num_once" json:"msg_num_once"`
	MsgLoopSeconds int64  `toml:"msg_loop_seconds" json:"msg_loop_seconds"`
}

type Consumer struct {
	conf     *ConsumerConfig
	handler  func(mq.ConsumeMessageEntry) error
	consumer mq.MQConsumer
	queue    chan mq.ConsumeMessageResponse
	errChan  chan error
	logger   *logger.Logger
	ctx      context.Context
	cancel   context.CancelFunc
	wg       *sync.WaitGroup
}

func (c *Consumer) run() (err error) {
	c.wg.Add(c.conf.Worker + 1)

	go c.receive()
	go c.logError()

	for i := 0; i < c.conf.Worker; i++ {
		go c.handle()
	}

	return
}

func (c *Consumer) Stop() {
	c.cancel()
	c.wg.Wait()
}

func (c *Consumer) handle() {
	defer c.wg.Done()

	for resp := range c.queue {
		handles := make([]string, 0, len(resp.Messages))

		for _, entity := range resp.Messages {
			c.logger.Debugf("receive rocketmq message | msg: %+v", entity)

			if err := c.handler(entity); err != nil {
				c.logger.Errorf("handle rocketmq message failed | message: %+v | error: %s", entity, err)
			}

			handles = append(handles, entity.ReceiptHandle)
		}

		if err := c.consumer.AckMessage(handles); err != nil {
			c.logger.Errorf("ack rocketmq message failed | handles: %+v | error: %s", handles, err)
		}
	}
}

func (c *Consumer) logError() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		case err := <-c.errChan:
			if !strings.Contains(err.(errors.ErrCode).Error(), "MessageNotExist") {
				c.logger.Errorf("receive rocketmq error | error: %s", err)
			}
		}
	}
}

func (c *Consumer) receive() {
	defer close(c.queue)
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			c.consumer.ConsumeMessage(c.queue, c.errChan, c.conf.MsgNumOnce, c.conf.MsgLoopSeconds)
		}
	}
}

func NewConsumer(conf *ConsumerConfig, handler func(mq.ConsumeMessageEntry) error, logger *logger.Logger) (consumer *Consumer, err error) {
	consumer = &Consumer{
		conf:    conf,
		handler: handler,
		queue:   make(chan mq.ConsumeMessageResponse, 4096),
		errChan: make(chan error, 256),
		logger:  logger,
		wg:      &sync.WaitGroup{},
	}

	client := mq.NewAliyunMQClient(conf.Endpoint, conf.AccessKey, conf.SecretKey, conf.SecurityToken)
	consumer.consumer = client.GetConsumer(conf.InstanceId, conf.Topic, conf.GroupId, conf.Tag)

	consumer.ctx, consumer.cancel = context.WithCancel(context.Background())
	err = consumer.run()

	return
}
