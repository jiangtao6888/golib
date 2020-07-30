package mqtt

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"time"
)

type ConnectConfig struct {
	Host              string        `toml:"host" json:"host"`
	Port              int           `toml:"port" json:"port"`
	Username          string        `toml:"username" json:"username"`
	Password          string        `toml:"password" json:"password"`
	QoS               byte          `toml:"qos" json:"qos"`
	CleanSession      bool          `toml:"clean_session" json:"clean_session"`
	ConnectTimeout    time.Duration `toml:"connect_timeout" json:"connect_timeout"`
	DisconnectTimeout uint          `toml:"disconnect_timeout" json:"disconnect_timeout"`
}

func (c *ConnectConfig) GetAddr() string {
	return fmt.Sprintf("tcp://%s:%d", c.Host, c.Port)
}

func (c *ConnectConfig) GetOptions() *mqtt.ClientOptions {
	return mqtt.NewClientOptions().
		AddBroker(c.GetAddr()).
		SetUsername(c.Username).
		SetPassword(c.Password).
		SetAutoReconnect(true).
		SetCleanSession(c.CleanSession).
		SetConnectTimeout(c.ConnectTimeout * time.Millisecond)
}

type IConfig interface {
	GetOptions() *mqtt.ClientOptions
	GetClientID() string
	ConnectHandler(mqtt.Client)
	DisconnectHandler(mqtt.Client, error)
}

func connect(c IConfig) (mqtt.Client, error) {
	options := c.GetOptions().
		SetClientID(c.GetClientID()).
		SetOnConnectHandler(c.ConnectHandler).
		SetConnectionLostHandler(c.DisconnectHandler)

	client := mqtt.NewClient(options)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return client, nil
}
