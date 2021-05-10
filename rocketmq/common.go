package rocketmq

import "github.com/marsmay/golib/logger"

type ConnectConfig struct {
	Endpoints     []string `toml:"endpoints" json:"endpoints"`
	AccessKey     string   `toml:"access_key" json:"access_key"`
	SecretKey     string   `toml:"secret_key" json:"secret_key"`
	SecurityToken string   `toml:"security_token" json:"security_token"`
}

type Logger struct {
	l *logger.Logger
}

func (l *Logger) Debug(msg string, fields map[string]interface{}) {
	l.l.Debugf("message: %s | fields: %+v", msg, fields)
}

func (l *Logger) Info(msg string, fields map[string]interface{}) {
	l.l.Infof("message: %s | fields: %+v", msg, fields)
}

func (l *Logger) Warning(msg string, fields map[string]interface{}) {
	l.l.Warningf("message: %s | fields: %+v", msg, fields)
}

func (l *Logger) Error(msg string, fields map[string]interface{}) {
	l.l.Errorf("message: %s | fields: %+v", msg, fields)
}

func (l *Logger) Fatal(msg string, fields map[string]interface{}) {
	l.l.Fatalf("message: %s | fields: %+v", msg, fields)
}

func (l *Logger) Level(_ string) {}

func (l *Logger) OutputPath(path string) (err error) { return nil }
