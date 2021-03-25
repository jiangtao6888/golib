package gorm

import (
	"errors"
	"fmt"
	"sync"
	"time"

	oLogger "github.com/marsmay/golib/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Config struct {
	Host          string        `toml:"host" json:"host"`
	Port          uint          `toml:"port" json:"port"`
	User          string        `toml:"user" json:"user"`
	Password      string        `toml:"password" json:"password"`
	Charset       string        `toml:"charset" json:"charset"`
	Database      string        `toml:"database" json:"database"`
	Timeout       int           `toml:"timeout" json:"timeout"`
	MaxOpenConns  int           `toml:"max_open_conns" json:"max_open_conns"`
	MaxIdleConns  int           `toml:"max_idle_conns" json:"max_idle_conns"`
	MaxConnTtl    int           `toml:"max_conn_ttl" json:"max_conn_ttl"`
	SlowQueryTime time.Duration `toml:"slow_query_time" json:"slow_query_time"`
}

func (c *Config) GetDsn() string {
	if c.Timeout <= 0 {
		c.Timeout = 3
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local&timeout=%ds",
		c.User, c.Password, c.Host, c.Port, c.Database, c.Charset, c.Timeout)
}

type Pool struct {
	locker  sync.RWMutex
	clients map[string]*gorm.DB
	logger  *oLogger.Logger
}

func (p *Pool) Add(name string, c *Config) error {
	p.locker.Lock()
	defer p.locker.Unlock()

	orm, err := gorm.Open(mysql.Open(c.GetDsn()), &gorm.Config{
		Logger: &logger{deadline: c.SlowQueryTime * time.Millisecond, recorder: p.logger},
	})

	if err != nil {
		return err
	}

	db, err := orm.DB()

	if err != nil {
		return err
	}

	if c.MaxIdleConns > 0 {
		db.SetMaxIdleConns(c.MaxIdleConns)
	}

	if c.MaxOpenConns > 0 {
		db.SetMaxOpenConns(c.MaxOpenConns)
	}

	if c.MaxConnTtl > 0 {
		db.SetConnMaxLifetime(time.Duration(c.MaxConnTtl) * time.Second)
	}

	p.clients[name] = orm

	return nil
}

func (p *Pool) Get(name string) (*gorm.DB, error) {
	p.locker.RLock()
	defer p.locker.RUnlock()

	client, ok := p.clients[name]

	if ok {
		return client, nil
	}

	return nil, errors.New("no mysql gorm client")
}

func NewPool(logger *oLogger.Logger) *Pool {
	return &Pool{clients: make(map[string]*gorm.DB, 64), logger: logger}
}
