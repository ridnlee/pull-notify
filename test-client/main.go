package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"test-client/pkg/pb"
	"time"

	"github.com/caarlos0/env"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

type Config struct {
	LogLevel            string `env:"LOG_LEVEL" envDefault:"debug"`
	NotificationService string `env:"NOTIFICATION_SERVICE" envDefault:"notification-service:3000"`
}

func main() {
	cfg := GetConfig()
	setLogger(cfg)
	logrus.Infof("%+v\n", cfg)

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}
	conn, err := grpc.Dial(cfg.NotificationService, opts...)

	if err != nil {
		grpclog.Fatalf("fail to dial: %v", err)
	}

	defer conn.Close()

	client := pb.NewNotifyerClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := run(ctx, client)
		if err != nil {
			logrus.Panic(err)
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

func run(ctx context.Context, client pb.NotifyerClient) error {
	tickerNotification := time.NewTicker(2 * time.Second)
	tickerHistory := time.NewTicker(10 * time.Second)
	offset := int64(0)
	for {
		select {
		case <-ctx.Done():
			logrus.Info("producer stopped")
			return nil
		case _ = <-tickerNotification.C:
			resp, err := client.GetLastMsgs(ctx, &pb.GetLastMsgsRequest{ClientId: 1, LastOffset: offset})
			if err != nil {
				return errors.Wrap(err, "cannot get last messages")
			}
			offset = resp.LastOffset
			logrus.Debugf("NOTIFICATIONS %d: got %d new messages: %+v", offset, len(resp.Msgs), resp.Msgs)
		case _ = <-tickerHistory.C:
			resp, err := client.GetMsgList(ctx, &pb.GetMsgListRequest{ClientId: 1})
			if err != nil {
				return errors.Wrap(err, "cannot get last messages")
			}
			logrus.Debugf("HISTORY: got %d messages", len(resp.Msgs))
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
