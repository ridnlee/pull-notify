package repo

import (
	"context"
	"pull-notify/pkg/pb"

	"github.com/go-redis/redis/v8"
)

type NotifyStorage struct {
	r *redis.Client
}

func NewNotifyStorage(r *redis.Client) *NotifyStorage {
	return &NotifyStorage{r}
}

func (st *NotifyStorage) GetLastMsgsByOffset(ctx context.Context, client_id int64, offset int64, lastMsgLimit int) ([]*pb.Msg, error) {
	return nil, nil
}
func (st *NotifyStorage) GetMsgList(ctx context.Context, client_id int64) ([]*pb.Msg, error) {
	return nil, nil
}
func (st *NotifyStorage) MarkRead(ctx context.Context, client_id int64, msgID string, offset int64) error {
	return nil
}
