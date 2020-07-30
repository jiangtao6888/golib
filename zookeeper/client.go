package zookeeper

import (
	"context"
	"github.com/samuel/go-zookeeper/zk"
	"github.com/Zivn/golib/logger"
	"strings"
	"sync"
	"time"
)

const (
	EventTypeAll = 0
)

const (
	FlagPersistent           = 0
	FlagEphemeralAndSequence = zk.FlagEphemeral + zk.FlagSequence
)

const (
	WatchTypeNode = iota
	WatchTypeChildren
)

var PermWorldAll = zk.WorldACL(zk.PermAll)

type zkLogger struct {
	*logger.Logger
}

func (l *zkLogger) Printf(format string, args ...interface{}) {
	l.Errorf(format, args...)
}

type Config struct {
	Addrs   []string      `toml:"addrs" json:"addrs"`
	Timeout time.Duration `toml:"timeout" json:"timeout"`
}

type Client struct {
	c      *Config
	conn   *zk.Conn
	logger *logger.Logger
	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup
}

func (c *Client) WatchNode(path string, eventType zk.EventType, callback func(zk.Event)) {
	c.wg.Add(1)

	go func() {
		defer c.wg.Done()

		for {
			select {
			case <-c.ctx.Done():
				return
			default:
				_, _, eventCh, err := c.conn.ExistsW(path)

				if err != nil {
					c.logger.Warningf("zookeeper watch failed | path: %s | error: %s", path, err)
					continue
				}

				event := <-eventCh

				if event.Err != nil {
					if event.Err != zk.ErrClosing {
						c.logger.Warningf("zookeeper watch failed | path: %s | error: %s", path, event.Err)
					}

					continue
				}

				if eventType == EventTypeAll || eventType == event.Type {
					callback(event)
				}
			}
		}
	}()
}

func (c *Client) WatchChildren(path string, eventType zk.EventType, callback func(zk.Event)) {
	c.wg.Add(1)

	go func() {
		defer c.wg.Done()

		for {
			select {
			case <-c.ctx.Done():
				return
			default:
				_, _, eventCh, err := c.conn.ChildrenW(path)

				if err != nil {
					if err != zk.ErrNoNode {
						c.logger.Warningf("zookeeper watch failed | path: %s | error: %s", path, err)
					}

					continue
				}

				event := <-eventCh

				if event.Err != nil {
					if event.Err != zk.ErrClosing {
						c.logger.Warningf("zookeeper watch failed | path: %s | error: %s", path, event.Err)
					}

					continue
				}

				if eventType == EventTypeAll || eventType == event.Type {
					callback(event)
				}
			}
		}
	}()
}

func (c *Client) connect() (err error) {
	conn, _, err := zk.Connect(c.c.Addrs, c.c.Timeout*time.Second, zk.WithLogger(&zkLogger{c.logger}), zk.WithLogInfo(false))

	if err == nil {
		c.conn = conn
	}

	return
}

func (c *Client) Create(path string, data []byte, flags int32, acl []zk.ACL) (err error) {
	items := strings.Split(path, "/")

	if len(items) > 2 {
		parentPath := strings.Join(items[0:len(items)-1], "/")
		ok, _, e := c.conn.Exists(parentPath)

		if e != nil {
			err = e
			return
		}

		if !ok {
			if err = c.Create(parentPath, nil, FlagPersistent, acl); err != nil {
				return
			}
		}
	}

	_, err = c.conn.Create(path, data, flags, acl)
	return
}

func (c *Client) Update(path string, data []byte) (err error) {
	ok, stat, err := c.conn.Exists(path)

	if err != nil {
		return
	}

	if !ok {
		return zk.ErrNoNode
	}

	_, err = c.conn.Set(path, data, stat.Version)
	return
}

func (c *Client) Delete(path string) (err error) {
	ok, stat, err := c.conn.Exists(path)

	if err != nil {
		return
	}

	if !ok {
		return zk.ErrNoNode
	}

	err = c.conn.Delete(path, stat.Version)
	return
}

func (c *Client) Children(path string) (children []string, err error) {
	children, _, err = c.conn.Children(path)
	return
}

func (c *Client) GetNode(path string) (data []byte, err error) {
	data, _, err = c.conn.Get(path)
	return
}

func (c *Client) GetNodes(path string) (datas map[string][]byte, err error) {
	nodes, _, err := c.conn.Children(path)

	if err != nil {
		if err == zk.ErrNoNode {
			err = nil
		}

		return
	}

	datas = make(map[string][]byte, len(nodes))

	for _, node := range nodes {
		data, _, e := c.conn.Get(path + "/" + node)

		if e != nil {
			if e == zk.ErrNoNode {
				continue
			}

			err = e
			return
		}

		datas[node] = data
	}

	return
}

func (c *Client) Conn() *zk.Conn {
	return c.conn
}

func (c *Client) Close() {
	c.cancel()
	c.conn.Close()
	c.wg.Wait()
}

func New(c *Config, logger *logger.Logger) (client *Client, err error) {
	client = &Client{
		c:      c,
		logger: logger,
		wg:     &sync.WaitGroup{},
	}
	client.ctx, client.cancel = context.WithCancel(context.Background())

	if err = client.connect(); err != nil {
		return
	}

	return
}
