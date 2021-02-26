package mqtt

import (
	"context"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/marsmay/golib/logger"
)

type ConsumerConfig struct {
	*ConnectConfig
	ClientId string        `toml:"client_id" json:"client_id"`
	Timeout  time.Duration `toml:"timeout" json:"timeout"`
	Worker   int           `toml:"worker" json:"worker"`
}

type Consumer struct {
	c        *ConsumerConfig
	client   mqtt.Client
	msgQueue chan mqtt.Message
	handlers map[string]func([]byte)
	logger   *logger.Logger
	ctx      context.Context
	cancel   context.CancelFunc
	wg       *sync.WaitGroup
}

func (c *Consumer) GetOptions() *mqtt.ClientOptions {
	return c.c.GetOptions()
}

func (c *Consumer) GetClientID() string {
	return c.c.ClientId
}

func (c *Consumer) ConnectHandler(client mqtt.Client) {
	c.logger.Debugf("mqtt connected | addr: %s", c.c.ConnectConfig.GetAddr())

	filters := make(map[string]byte, len(c.handlers))

	for t := range c.handlers {
		filters[t] = c.c.QoS
	}

	token := client.SubscribeMultiple(filters, func(client mqtt.Client, msg mqtt.Message) {
		c.msgQueue <- msg
	})

	if ok := token.WaitTimeout(c.c.Timeout * time.Millisecond); !ok {
		c.logger.Errorf("mqtt subscribe timeout | filters: %+v", filters)
	}

	if err := token.Error(); err != nil {
		c.logger.Errorf("mqtt subscribe failed | filters: %+v | error: %s", filters, err)
	}
}

func (c *Consumer) DisconnectHandler(_ mqtt.Client, err error) {
	c.logger.Debugf("mqtt lost connection | addr: %s | error: %s", c.c.ConnectConfig.GetAddr(), err)
}

func (c *Consumer) run() (err error) {
	if c.client, err = connect(c); err != nil {
		return
	}

	c.wg.Add(c.c.Worker)

	for i := 0; i < c.c.Worker; i++ {
		go c.handle()
	}

	return
}

func (c *Consumer) Stop() {
	topics := make([]string, 0, len(c.handlers))

	for t := range c.handlers {
		topics = append(topics, t)
	}

	token := c.client.Unsubscribe(topics...)

	if ok := token.WaitTimeout(c.c.Timeout * time.Millisecond); !ok {
		c.logger.Errorf("unsubscribe topics timeout | addr: %s | topics: %+v", c.c.GetAddr(), topics)
	}

	if err := token.Error(); err != nil {
		c.logger.Errorf("can't unsubscribe topics | addr: %s | topics: %+v | error: %s", c.c.GetAddr(), topics, err)
	}

	c.client.Disconnect(c.c.DisconnectTimeout)

	c.cancel()
	c.wg.Wait()
}

func (c *Consumer) handle() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		case msg := <-c.msgQueue:
			if handler, ok := c.handlers[msg.Topic()]; ok {
				handler(msg.Payload())
			}
		}
	}
}

func NewConsumer(c *ConsumerConfig, handlers map[string]func([]byte), logger *logger.Logger) (consumer *Consumer, err error) {
	consumer = &Consumer{
		c:        c,
		msgQueue: make(chan mqtt.Message, 4096),
		handlers: handlers,
		logger:   logger,
		wg:       &sync.WaitGroup{},
	}

	consumer.ctx, consumer.cancel = context.WithCancel(context.Background())
	err = consumer.run()
	return
}
