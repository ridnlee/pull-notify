package input

import (
	"context"
	"notification-service/internal/config"
	"notification-service/pkg/pb"
	"time"
)

type NotifyRepo interface {
	GetLastMsgsByOffset(ctx context.Context, client_id int64, offset int64, lastMsgLimit int) ([]*pb.Msg, error)
	GetMsgList(ctx context.Context, client_id int64) ([]*pb.Msg, error)
	MarkRead(ctx context.Context, client_id int64, msgID string, offset int64) error
}

type NotifyServer struct {
	pb.UnimplementedNotifyerServer

	repo          NotifyRepo
	MsgTTL        time.Duration
	FlushPeriod   time.Duration
	LastMsgPeriod time.Duration
	LastMsgLimit  int
}

func NewNotifyServer(repo NotifyRepo, cfg *config.Config) *NotifyServer {
	return &NotifyServer{
		repo:          repo,
		MsgTTL:        cfg.MsgTTL,
		FlushPeriod:   cfg.FlushPeriod,
		LastMsgPeriod: cfg.LastMsgPeriod,
		LastMsgLimit:  cfg.LastMsgLimit,
	}
}

func (ns *NotifyServer) GetLastMsgs(ctx context.Context, req *pb.GetLastMsgsRequest) (*pb.GetLastMsgsResponse, error) {
	offset := req.GetLastOffset()
	if offset == 0 {
		offset = time.Now().Add(ns.LastMsgPeriod * -1).Unix()
	}

	msgs, err := ns.repo.GetLastMsgsByOffset(ctx, req.GetClientId(), offset, ns.LastMsgLimit)
	if err != nil {
		return nil, err
	}

	return &pb.GetLastMsgsResponse{Msgs: msgs, LastOffset: ns.getLastOffset(msgs)}, nil
}

func (ns *NotifyServer) GetMsgList(ctx context.Context, req *pb.GetMsgListRequest) (*pb.GetMsgListResponse, error) {
	msgs, err := ns.repo.GetMsgList(ctx, req.GetClientId())
	if err != nil {
		return nil, err
	}
	return &pb.GetMsgListResponse{Msgs: msgs}, nil
}

func (ns *NotifyServer) MarkRead(ctx context.Context, req *pb.MarkReadRequest) (*pb.MarkReadResponse, error) {
	err := ns.repo.MarkRead(ctx, req.GetClientId(), req.GetMsgId(), req.GetOffset())
	if err != nil {
		return nil, err
	}
	return &pb.MarkReadResponse{}, nil
}

func (ns *NotifyServer) getLastOffset(msgs []*pb.Msg) int64 {
	if msgs == nil {
		return 0
	}
	var offset int64
	for _, msg := range msgs {
		if msg.GetCreated() > offset {
			offset = msg.GetCreated()
		}
	}
	return offset
}
