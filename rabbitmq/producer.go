package rabbitmq

import (
	"context"
	"errors"
	"github.com/Zivn/golib/logger"
	"github.com/streadway/amqp"
	"sync"
	"time"
)

type ProducerConfig struct {
	*ConnectConfig
	Queue        string `toml:"queue" json:"queue"`
	Exchange     string `toml:"exchange" json:"exchange"`
	ExchangeType string `toml:"exchange_type" json:"exchange_type"`
	RoutingKey   string `toml:"routing_key" json:"routing_key"`
	AutoDelete   bool   `toml:"auto_delete" json:"auto_delete"`
	Durable      bool   `toml:"durable" json:"durable"`
}

type Producer struct {
	conf          *ProducerConfig
	conn          *amqp.Connection
	channel       *amqp.Channel
	queue         chan *amqp.Publishing
	connNotify    chan *amqp.Error
	channelNotify chan *amqp.Error
	logger        *logger.Logger
	ctx           context.Context
	cancel        context.CancelFunc
	wg            *sync.WaitGroup
}

func (c *Producer) run() (err error) {
	if err = c.init(); err != nil {
		return
	}

	c.wg.Add(2)

	go c.checkConn()
	go c.send()

	return
}

func (c *Producer) close() {
	if !c.conn.IsClosed() {
		if err := c.channel.Close(); err != nil {
			c.logger.Warningf("close rabbitmq consumer channel failed | error: %s", err)
		}

		if err := c.conn.Close(); err != nil {
			c.logger.Warningf("close rabbitmq consumer connection failed | error: %s", err)
		}
	}
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
	close(c.queue)
	c.wg.Wait()
	c.close()
}

func (c *Producer) Send(msg *amqp.Publishing) error {
	if c.closed() {
		return errors.New("producer is stoped")
	}

	c.queue <- msg
	return nil
}

func (c *Producer) send() {
	defer c.wg.Done()

	for msg := range c.queue {
		if err := c.channel.Publish(c.conf.Exchange, c.conf.RoutingKey, false, false, *msg); err != nil {
			c.logger.Errorf("publish rabbitmq message failed | queue: %s | message: %+v | error: %s", c.conf.Queue, msg, err)

			if err := c.Send(msg); err != nil {
				c.logger.Errorf("publish rabbitmq retry message failed | queue: %s | message: %+v | error: %s", c.conf.Queue, msg, err)
			}

			time.Sleep(time.Millisecond * 100)
			continue
		}

		c.logger.Debugf("send rabbitmq message | msg: %s", msg.Body)
	}
}

func (c *Producer) init() (err error) {
	if c.conn, err = amqp.Dial(c.conf.Addr()); err != nil {
		c.logger.Error(c.conf.Addr(), err)
		return
	}

	if c.channel, err = c.conn.Channel(); err != nil {
		c.logger.Error(err)
		_ = c.conn.Close()
		return
	}

	if c.conf.Exchange != "" {
		if err := c.channel.ExchangeDeclare(c.conf.Exchange, c.conf.ExchangeType, c.conf.Durable, c.conf.AutoDelete, false, false, nil); err != nil {
			c.logger.Error(err)
			_ = c.channel.Close()
			_ = c.conn.Close()
		}
	}

	c.conn.NotifyClose(c.connNotify)
	c.channel.NotifyClose(c.channelNotify)
	return
}

func (c *Producer) checkConn() {
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

func NewProducer(conf *ProducerConfig, logger *logger.Logger) (producer *Producer, err error) {
	producer = &Producer{
		conf:          conf,
		queue:         make(chan *amqp.Publishing, 4096),
		connNotify:    make(chan *amqp.Error),
		channelNotify: make(chan *amqp.Error),
		logger:        logger,
		wg:            &sync.WaitGroup{},
	}

	producer.ctx, producer.cancel = context.WithCancel(context.Background())
	err = producer.run()
	return
}
