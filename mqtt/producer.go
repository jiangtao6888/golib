package mqtt

import (
	"context"
	"errors"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/Zivn/golib/logger"
	"sync"
	"time"
)

type Message struct {
	Topic    string
	Retained bool
	Payload  []byte
}

func (m *Message) String() string {
	return fmt.Sprintf("{Topic:%s Retained:%+v Payload:%s}", m.Topic, m.Retained, m.Payload)
}

type ProducerConfig struct {
	*ConnectConfig
	ClientId      string        `toml:"client_id"`
	Timeout       time.Duration `toml:"timeout"`
	RetryInterval time.Duration `toml:"retry_interval"`
}

type Producer struct {
	c        *ProducerConfig
	client   mqtt.Client
	msgQueue chan *Message
	logger   *logger.Logger
	ctx      context.Context
	cancel   context.CancelFunc
	wg       *sync.WaitGroup
}

func (c *Producer) GetOptions() *mqtt.ClientOptions {
	return c.c.GetOptions()
}

func (c *Producer) GetClientID() string {
	return c.c.ClientId
}

func (c *Producer) ConnectHandler(_ mqtt.Client) {
	c.logger.Debugf("mqtt connected | addr: %s", c.c.ConnectConfig.GetAddr())
}

func (c *Producer) DisconnectHandler(_ mqtt.Client, err error) {
	c.logger.Debugf("mqtt lost connection | addr: %s | error: %s", c.c.ConnectConfig.GetAddr(), err)
}

func (c *Producer) run() (err error) {
	if c.client, err = connect(c); err != nil {
		return
	}

	c.wg.Add(1)
	go c.publish()

	return
}

func (c *Producer) Stop() {
	c.cancel()
	c.wg.Wait()
	c.client.Disconnect(c.c.DisconnectTimeout)
}

func (c *Producer) Send(topic string, retained bool, payload []byte) error {
	select {
	case <-c.ctx.Done():
		return errors.New("producer is stoped")
	default:
		c.msgQueue <- &Message{topic, retained, payload}
		return nil
	}
}

func (c *Producer) retry(msg *Message) {
	c.msgQueue <- msg
	time.Sleep(c.c.RetryInterval * time.Millisecond)
}

func (c *Producer) publish() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		case msg := <-c.msgQueue:
			token := c.client.Publish(msg.Topic, c.c.QoS, msg.Retained, msg.Payload)

			if ok := token.WaitTimeout(c.c.Timeout * time.Millisecond); !ok {
				c.logger.Errorf("publish message timeout | addr: %s | message: %s", c.c.GetAddr(), msg)
				c.retry(msg)
				continue
			}

			if err := token.Error(); err != nil {
				c.logger.Errorf("can't publish message | addr: %s | message: %s | error: %s", c.c.GetAddr(), msg, err)
				c.retry(msg)
				continue
			}

			c.logger.Debugf("publish message | addr: %s | message: %s", c.c.GetAddr(), msg)
		}
	}
}

func NewProducer(c *ProducerConfig, logger *logger.Logger) (producer *Producer, err error) {
	producer = &Producer{
		c:        c,
		msgQueue: make(chan *Message, 4096),
		logger:   logger,
		wg:       &sync.WaitGroup{},
	}

	producer.ctx, producer.cancel = context.WithCancel(context.Background())
	err = producer.run()
	return
}
