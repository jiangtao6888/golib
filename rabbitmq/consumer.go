package rabbitmq

import (
	"context"
	"github.com/streadway/amqp"
	"github.com/Zivn/golib/logger"
	"sync"
	"time"
)

type ConsumerConfig struct {
	*ConnectConfig
	Name         string `toml:"name" json:"name"`
	Queue        string `toml:"queue" json:"queue"`
	Exchange     string `toml:"exchange" json:"exchange"`
	ExchangeType string `toml:"exchange_type" json:"exchange_type"`
	RoutingKey   string `toml:"routing_key" json:"routing_key"`
	AutoDelete   bool   `toml:"auto_delete" json:"auto_delete"`
	AutoAck      bool   `toml:"auto_ack" json:"auto_ack"`
	Durable      bool   `toml:"durable" json:"durable"`
	Worker       int    `toml:"worker" json:"worker"`
}

type Consumer struct {
	conf          *ConsumerConfig
	handler       func(amqp.Delivery) error
	conn          *amqp.Connection
	channel       *amqp.Channel
	queue         <-chan amqp.Delivery
	connNotify    chan *amqp.Error
	channelNotify chan *amqp.Error
	logger        *logger.Logger
	ctx           context.Context
	cancel        context.CancelFunc
	wg            *sync.WaitGroup
}

func (c *Consumer) run() (err error) {
	if err = c.init(); err != nil {
		return
	}

	c.wg.Add(c.conf.Worker + 1)

	go c.checkConn()

	for i := 0; i < c.conf.Worker; i++ {
		go c.handle()
	}

	return
}

func (c *Consumer) close() {
	if !c.conn.IsClosed() {
		if err := c.channel.Cancel(c.conf.Name, true); err != nil {
			c.logger.Warningf("close rabbitmq consumer channel failed | error: %s", err)
		}

		if err := c.conn.Close(); err != nil {
			c.logger.Warningf("close rabbitmq consumer connection failed | error: %s", err)
		}
	}
}

func (c *Consumer) Stop() {
	c.cancel()
	c.close()
	c.wg.Wait()
}

func (c *Consumer) handle() {
	defer c.wg.Done()

	for msg := range c.queue {
		c.logger.Debugf("receive rabbitmq message | msg: %s", msg.Body)

		if err := c.handler(msg); err == nil {
			if err := msg.Ack(false); err != nil {
				c.logger.Errorf("ack rabbitmq message failed | message: %+v | error: %s", msg, err)
			}
		} else {
			if err := msg.Reject(true); err != nil {
				c.logger.Errorf("reject rabbitmq message failed | message: %+v | error: %s", msg, err)
			}
		}
	}
}

func (c *Consumer) init() (err error) {
	if c.conn, err = amqp.Dial(c.conf.Addr()); err != nil {
		c.logger.Error(c.conf.Addr(), err)
		return
	}

	if c.channel, err = c.conn.Channel(); err != nil {
		c.logger.Error(err)
		_ = c.conn.Close()
		return
	}

	if _, err = c.channel.QueueDeclare(c.conf.Queue, c.conf.Durable, c.conf.AutoDelete, false, false, nil); err != nil {
		c.logger.Error(err)
		_ = c.channel.Close()
		_ = c.conn.Close()
		return
	}

	if c.conf.Exchange != "" {
		if err := c.channel.ExchangeDeclare(c.conf.Exchange, c.conf.ExchangeType, c.conf.Durable, c.conf.AutoDelete, false, false, nil); err != nil {
			c.logger.Error(err)
			_ = c.channel.Close()
			_ = c.conn.Close()
		}

		if err = c.channel.QueueBind(c.conf.Queue, c.conf.RoutingKey, c.conf.Exchange, false, nil); err != nil {
			c.logger.Error(err)
			_ = c.channel.Close()
			_ = c.conn.Close()
			return
		}
	}

	if c.queue, err = c.channel.Consume(c.conf.Queue, c.conf.Name, c.conf.AutoAck, false, false, false, nil); err != nil {
		c.logger.Error(err)
		_ = c.channel.Close()
		_ = c.conn.Close()
		return
	}

	c.conn.NotifyClose(c.connNotify)
	c.channel.NotifyClose(c.channelNotify)
	return
}

func (c *Consumer) checkConn() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		case err := <-c.connNotify:
			c.logger.Warningf("notify rabbitmq connection closed | error: %s", err)
		case err := <-c.channelNotify:
			c.logger.Warningf("notify rabbitmq channel closed | error: %s", err)
		}

		c.close()

		for err := range c.connNotify {
			c.logger.Warningf("notify rabbitmq connection closed | error: %s", err)
		}

		for err := range c.channelNotify {
			c.logger.Warningf("notify rabbitmq channel closed | error: %s", err)
		}

		for {
			select {
			case <-c.ctx.Done():
				return
			default:
				if err := c.init(); err != nil {
					c.logger.Warningf("rabbitmq connection reconnect failed | error: %s", err)
					time.Sleep(time.Second * 3)
					continue
				}

				break
			}
		}
	}
}

func NewConsumer(conf *ConsumerConfig, handler func(amqp.Delivery) error, logger *logger.Logger) (consumer *Consumer, err error) {
	consumer = &Consumer{
		conf:          conf,
		handler:       handler,
		connNotify:    make(chan *amqp.Error),
		channelNotify: make(chan *amqp.Error),
		logger:        logger,
		wg:            &sync.WaitGroup{},
	}

	consumer.ctx, consumer.cancel = context.WithCancel(context.Background())
	err = consumer.run()
	return
}
