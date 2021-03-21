package main

import (
	"fmt"
	"log"
	"net"
	"pull-notify/internal/config"
	"pull-notify/internal/input"
	"pull-notify/internal/repo"

	"pull-notify/pkg/pb"

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

	notifyStorage := repo.NewNotifyStorage(r)

	grpcServer := grpc.NewServer([]grpc.ServerOption{}...)
	pb.RegisterNotifyerServer(grpcServer, input.NewNotifyServer(notifyStorage, cfg))
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", cfg.Port))
	if err != nil {
		log.Panicf("failed to listen: %v", err)
	}
	grpcServer.Serve(lis)
}

func setLogger(cfg *config.Config) {
	ll, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		panic(err)
	}
	logrus.SetLevel(ll)
}
