package main

import (
	"context"
	"fmt"
	"notification-service/pkg/pb"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/caarlos0/env"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

type Config struct {
	LogLevel  string `env:"LOG_LEVEL" envDefault:"debug"`
	RedisAddr string `env:"REDIS_ADDR" envDefault:"localhost:6379"`
	RedisDB   int    `env:"REDIS_DB" envDefault:"0"`
}

func main() {
	cfg := GetConfig()
	setLogger(cfg)
	logrus.Infof("%+v\n", cfg)

	r := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
		DB:   cfg.RedisDB,
	})

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := produce(ctx, r)
		if err != nil {
			logrus.Error(err)
		}
		cancel()
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	select {
	case <-sigs:
		cancel()
	case <-ctx.Done():
	}

	wg.Wait()
}

func produce(ctx context.Context, r *redis.Client) error {
	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-ctx.Done():
			logrus.Info("producer stopped")
			return nil
		case t := <-ticker.C:
			msg := &pb.Msg{
				Id:       uuid.New().String(),
				ClientId: 1,
				Created:  time.Now().Unix(),
				Title:    fmt.Sprintf("message %d", t.Unix()),
				Body:     fmt.Sprintf("message body"),
			}
			mMsg, err := proto.Marshal(msg)
			if err != nil {
				return errors.Wrap(err, "cannot marshal message")
			}
			r.Publish(ctx, "input", mMsg)
			logrus.Debugf("message sent %s", mMsg)
		}
	}
}

func setLogger(cfg *Config) {
	ll, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		panic(err)
	}
	logrus.SetLevel(ll)
}

func GetConfig() *Config {
	cfg := &Config{}

	if err := env.Parse(cfg); err != nil {
		panic(err)
	}

	return cfg
}
