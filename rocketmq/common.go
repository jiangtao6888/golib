package rocketmq

import "github.com/marsmay/golib/logger"

type ConnectConfig struct {
	Endpoints     []string `toml:"endpoints" json:"endpoints"`
	AccessKey     string   `toml:"access_key" json:"access_key"`
	SecretKey     string   `toml:"secret_key" json:"secret_key"`
	SecurityToken string   `toml:"security_token" json:"security_token"`
}

type Logger struct {
	logger *logger.Logger
	quiet  bool
}

func (l *Logger) Debug(msg string, fields map[string]interface{}) {
	if !l.quiet {
		l.logger.Debug(msg, fields)
	}
}

func (l *Logger) Info(msg string, fields map[string]interface{}) {
	if !l.quiet {
		l.logger.Info(msg, fields)
	}
}

func (l *Logger) Warning(msg string, fields map[string]interface{}) {
	l.logger.Warning(msg, fields)
}

func (l *Logger) Error(msg string, fields map[string]interface{}) {
	l.logger.Error(msg, fields)
}

func (l *Logger) Fatal(msg string, fields map[string]interface{}) {
	l.logger.Fatal(msg, fields)
}

func (l *Logger) Level(_ string) {}

func (l *Logger) OutputPath(_ string) (err error) { return nil }
