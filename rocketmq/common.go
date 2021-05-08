package rocketmq

type ConnectConfig struct {
	Endpoints     []string `toml:"endpoints" json:"endpoints"`
	AccessKey     string   `toml:"access_key" json:"access_key"`
	SecretKey     string   `toml:"secret_key" json:"secret_key"`
	SecurityToken string   `toml:"security_token" json:"security_token"`
}
