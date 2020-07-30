package redis

import (
	"errors"
	"github.com/go-redis/redis"
	"sync"
)

const (
	ReadStrategySlaveOnly = iota + 1
	ReadStrategyClosestNode
	ReadStrategyRandomNode
)

type ClusterConfig struct {
	Addrs        []string `toml:"addrs" json:"addrs"`
	Password     string   `toml:"password" json:"password"`
	ReadStrategy int      `toml:"read_strategy" json:"read_strategy"`
	PoolSize     int      `toml:"poolsize" json:"poolsize"`
}

type ClusterClient struct {
	*redis.ClusterClient
}

func (c *ClusterClient) MGet(keys ...string) (values map[string]string, err error) {
	pipe := c.Pipeline()

	for _, key := range keys {
		pipe.Get(key)
	}

	cmders, err := pipe.Exec()

	if err == redis.Nil {
		err = nil
	}

	if err != nil {
		return
	}

	values = make(map[string]string, len(keys))

	for index, cmder := range cmders {
		if v, e := cmder.(*redis.StringCmd).Result(); e == nil {
			values[keys[index]] = v
		}
	}

	return
}

func (c *ClusterClient) MSet(values map[string]string) (results map[string]bool, err error) {
	keys := make([]string, 0, len(values))
	pipe := c.Pipeline()

	for k, v := range values {
		keys = append(keys, k)
		pipe.Set(k, v, 0)
	}

	cmders, err := pipe.Exec()

	if err != nil {
		return
	}

	results = make(map[string]bool, len(values))

	for index, cmder := range cmders {
		if e := cmder.(*redis.StatusCmd).Err(); e == nil {
			results[keys[index]] = true
		}
	}

	return
}

type ClusterPool struct {
	locker  sync.RWMutex
	clients map[string]*ClusterClient
}

func (p *ClusterPool) Add(name string, conf *ClusterConfig) {
	p.locker.Lock()
	defer p.locker.Unlock()

	options := &redis.ClusterOptions{
		Addrs:    conf.Addrs,
		Password: conf.Password,
		PoolSize: conf.PoolSize,
	}

	switch conf.ReadStrategy {
	case ReadStrategyRandomNode:
		options.ReadOnly = true
		options.RouteRandomly = true
	case ReadStrategyClosestNode:
		options.ReadOnly = true
		options.RouteByLatency = true
	case ReadStrategySlaveOnly:
		options.ReadOnly = true
	}

	p.clients[name] = &ClusterClient{ClusterClient: redis.NewClusterClient(options)}
}

func (p *ClusterPool) Get(name string) (client *ClusterClient, err error) {
	p.locker.RLock()
	defer p.locker.RUnlock()

	client, ok := p.clients[name]

	if !ok {
		err = errors.New("no redis cluster client")
	}

	return
}

func NewClusterPool() *ClusterPool {
	return &ClusterPool{clients: make(map[string]*ClusterClient, 16)}
}
