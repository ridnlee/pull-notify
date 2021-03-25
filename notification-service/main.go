package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"notification-service/internal/config"
	"notification-service/internal/input"
	"notification-service/internal/repo"
	"sync"

	"notification-service/pkg/pb"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.GetConfig()
	setLogger(cfg)
	logrus.Infof("%+v\n", cfg)

	r := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
		DB:   cfg.RedisDB,
	})

	notifyStorage := repo.NewNotifyStorage(r, cfg)

	wg := &sync.WaitGroup{}

	ctx, cancel := context.WithCancel(context.Background())
	inputQueue := input.NewInputQueue(r, notifyStorage, "input")

	grpcServer := grpc.NewServer([]grpc.ServerOption{}...)
	pb.RegisterNotifyerServer(grpcServer, input.NewNotifyServer(notifyStorage, cfg))
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", cfg.Port))
	if err != nil {
		log.Panicf("failed to listen: %v", err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		inputQueue.Run(ctx)
		grpcServer.GracefulStop()
	}()

	grpcServer.Serve(lis)

	cancel()
	wg.Wait()
}

func setLogger(cfg *config.Config) {
	ll, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		panic(err)
	}
	logrus.SetLevel(ll)
}
