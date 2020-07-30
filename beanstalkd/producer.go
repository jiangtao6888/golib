package beanstalkd

import (
	"context"
	"errors"
	"fmt"
	"github.com/kr/beanstalk"
	"github.com/Zivn/golib/logger"
	"sync"
	"time"
)

type ProducerConfig struct {
	*AddrList
	MaxJobSize int `toml:"max_job_size" json:"max_job_size"`
}

type Message struct {
	Tube     string
	Payload  []byte
	Priority uint32
	Delay    time.Duration
	Ttr      time.Duration
}

func (m *Message) String() string {
	return fmt.Sprintf("{Tube:%s Payload:%s Priority:%d Delay:%s Ttr:%s}", m.Tube, m.Payload, m.Priority, m.Delay, m.Ttr)
}

type Producer struct {
	c      *ProducerConfig
	queue  chan *Message
	conns  *Conns
	logger *logger.Logger
	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup
}

func (c *Producer) run() {
	c.wg.Add(len(c.c.Addrs))

	for _, addr := range c.c.Addrs {
		go c.send(addr)
	}
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
	if c.c.MaxJobSize > 0 && len(msg.Payload) > c.c.MaxJobSize {
		return errors.New("message payload is too big")
	}

	if c.closed() {
		return errors.New("producer is stoped")
	}

	c.queue <- msg
	return nil
}

func (c *Producer) put(addr string, msg *Message) (err error) {
	conn := c.conns.Get(addr)

	if conn == nil {
		if conn, err = reconnect(addr); err != nil {
			return
		}

		c.conns.Set(addr, conn)
	}

	tube := &beanstalk.Tube{Conn: conn, Name: msg.Tube}
	id, err := tube.Put(msg.Payload, msg.Priority, msg.Delay, msg.Ttr)

	if err != nil {
		return
	}

	c.logger.Debugf("create beanstalkd job | id: %d | msg: %s", id, msg)
	return
}

func (c *Producer) send(addr string) {
	defer c.wg.Done()

	for msg := range c.queue {
		if err := c.put(addr, msg); err != nil {
			c.logger.Errorf("create beanstalkd job failed | addr: %s | tube: %s | error: %s", addr, msg.Tube, err)

			if err := c.Send(msg); err != nil {
				c.logger.Errorf("create beanstalkd retry job failed | addr: %s | tube: %s | error: %s", addr, msg.Tube, err)
			}

			c.conns.Set(addr, nil)
			time.Sleep(time.Millisecond * 100)
			continue
		}
	}
}

func NewProducer(c *ProducerConfig, logger *logger.Logger) *Producer {
	producer := &Producer{
		c:      c,
		queue:  make(chan *Message, 32),
		conns:  &Conns{conns: make(map[string]*beanstalk.Conn, len(c.Addrs))},
		logger: logger,
		wg:     &sync.WaitGroup{},
	}

	producer.ctx, producer.cancel = context.WithCancel(context.Background())
	producer.run()
	return producer
}
