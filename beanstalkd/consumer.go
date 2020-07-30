package beanstalkd

import (
	"context"
	"github.com/kr/beanstalk"
	"github.com/Zivn/golib/logger"
	"sync"
	"time"
)

type ConsumerConfig struct {
	*AddrList
	Tube   string `toml:"tube" json:"tube"`
	Worker int    `toml:"worker" json:"worker"`
}

type Consumer struct {
	c       *ConsumerConfig
	queue   chan []byte
	handler func([]byte)
	conns   *Conns
	logger  *logger.Logger
	ctx     context.Context
	cancel  context.CancelFunc
	wg      *sync.WaitGroup
}

func (c *Consumer) run() {
	c.wg.Add(len(c.c.Addrs) + c.c.Worker)

	for _, addr := range c.c.Addrs {
		go c.receive(addr)
	}

	for i := 0; i < c.c.Worker; i++ {
		go c.handle()
	}
}

func (c *Consumer) Stop() {
	c.cancel()
	c.wg.Wait()
}

func (c *Consumer) handle() {
	defer c.wg.Done()

	for msg := range c.queue {
		c.handler(msg)
	}
}

func (c *Consumer) consume(addr string) (err error) {
	conn := c.conns.Get(addr)

	if conn == nil {
		if conn, err = reconnect(addr); err != nil {
			return
		}

		c.conns.Set(addr, conn)
	}

	tubeSet := beanstalk.NewTubeSet(conn, c.c.Tube)
	id, body, err := tubeSet.Reserve(3 * time.Second)

	if err != nil {
		if e, ok := err.(beanstalk.ConnError); ok && e.Err == beanstalk.ErrTimeout {
			err = nil
		}

		return
	}

	if err = conn.Delete(id); err != nil {
		return
	}

	c.queue <- body
	c.logger.Debugf("receive beanstalkd job | id: %d | msg: %s", id, body)
	return
}

func (c *Consumer) receive(addr string) {
	defer close(c.queue)
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			if err := c.consume(addr); err != nil {
				c.logger.Errorf("consume beanstalkd queue failed | addr: %s | tube: %s | error: %s", addr, c.c.Tube, err)

				c.conns.Set(addr, nil)
				time.Sleep(time.Second * 3)
				continue
			}
		}
	}

}

func NewConsumer(c *ConsumerConfig, handler func([]byte), logger *logger.Logger) *Consumer {
	consumer := &Consumer{
		c:       c,
		queue:   make(chan []byte, 32),
		handler: handler,
		conns:   &Conns{conns: make(map[string]*beanstalk.Conn, len(c.Addrs))},
		logger:  logger,
		wg:      &sync.WaitGroup{},
	}

	consumer.ctx, consumer.cancel = context.WithCancel(context.Background())
	consumer.run()
	return consumer
}
