package rabbitmq

import (
	"fmt"
)

type ConnectConfig struct {
	Host     string `toml:"host" json:"host"`
	Port     int    `toml:"port" json:"port"`
	Username string `toml:"username" json:"username"`
	Password string `toml:"password" json:"password"`
	Vhost    string `toml:"vhost" json:"vhost"`
}

func (c *ConnectConfig) Addr() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/%s", c.Username, c.Password, c.Host, c.Port, c.Vhost)
}
