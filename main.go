package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env"
	"github.com/sirupsen/logrus"
)

type config struct {
	Port        int           `env:"PORT" envDefault:"3000"`
	LogLevel    string        `env:"LOG_LEVEL" envDefault:"debug"`
	MsgTTL      time.Duration `env:"MSG_TTL" envDefault:"30d"`
	FlushPeriod time.Duration `env:"FLUSH_PERIOD" envDefault:"1d"`
}

func main() {
	cfg := getConfig()
	setLogger(cfg)
	logrus.Infof("%+v\n", cfg)

	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	select {
	case <-sigs:
		cancel()
	case <-ctx.Done():
	}

}

func getConfig() *config {
	cfg := &config{}

	if err := env.Parse(cfg); err != nil {
		panic(err)
	}

	return cfg
}

func setLogger(cfg *config) {
	ll, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		panic(err)
	}
	logrus.SetLevel(ll)
}
