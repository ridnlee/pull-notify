package config

import (
	"time"

	"github.com/caarlos0/env"
)

type Config struct {
	Port          int           `env:"PORT" envDefault:"3000"`
	LogLevel      string        `env:"LOG_LEVEL" envDefault:"debug"`
	MsgTTL        time.Duration `env:"MSG_TTL" envDefault:"720h"`
	FlushPeriod   time.Duration `env:"FLUSH_PERIOD" envDefault:"24h"`
	LastMsgPeriod time.Duration `env:"LAST_MSG_PERIOD" envDefault:"1m"`
	LastMsgLimit  int           `env:"LAST_MSG_LIMIT" envDefault:"3"`
	RedisAddr     string        `env:"REDIS_ADDR" envDefault:"localhost:6379"`
	RedisDB       int           `env:"REDIS_DB" envDefault:"0"`
}

func GetConfig() *Config {
	cfg := &Config{}

	if err := env.Parse(cfg); err != nil {
		panic(err)
	}

	return cfg
}
