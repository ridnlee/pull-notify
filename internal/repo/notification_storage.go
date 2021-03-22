package repo

import (
	"context"
	"fmt"
	"pull-notify/internal/config"
	"pull-notify/pkg/pb"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

const (
	listnameKeyPref = "nf:listname"
	listKeyPref     = "nf:list"
)

type NotifyStorage struct {
	r           *redis.Client
	msgTTL      time.Duration
	flushPeriod time.Duration
}

func NewNotifyStorage(r *redis.Client, cfg *config.Config) *NotifyStorage {
	return &NotifyStorage{r, cfg.MsgTTL, cfg.FlushPeriod}
}

func (st *NotifyStorage) GetLastMsgsByOffset(ctx context.Context, clientID int64, offset int64, lastMsgLimit int) ([]*pb.Msg, error) {
	listKey := getListKey(clientID, getPeriodID(st.flushPeriod))
	msgs, err := st.r.ZRevRangeByScore(ctx, listKey, &redis.ZRangeBy{Min: fmt.Sprintf("%d", offset), Count: int64(lastMsgLimit)}).Result()
	if err != nil {
		return nil, errors.Wrap(err, "cannot get msgs")
	}

	protoMsgs, err := msgsToProto(msgs)
	if err != nil {
		return nil, errors.Wrap(err, "cannot convert msgs to proto")
	}

	return protoMsgs, nil
}

func (st *NotifyStorage) GetMsgList(ctx context.Context, clientID int64) ([]*pb.Msg, error) {
	listKey, err := st.checkAndGetListKey(ctx, clientID)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get list key")
	}

	msgs, err := st.r.ZRevRangeByScore(ctx, listKey, &redis.ZRangeBy{}).Result()
	if err != nil {
		return nil, err
	}

	protoMsgs, err := msgsToProto(msgs)
	if err != nil {
		return nil, errors.Wrap(err, "cannot convert msgs to proto")
	}

	return protoMsgs, nil
}

func (st *NotifyStorage) MarkRead(ctx context.Context, clientID int64, msgID string, offset int64) error {
	return nil
}

func (st *NotifyStorage) checkAndGetListKey(ctx context.Context, clientID int64) (string, error) {
	curListKey := getListKey(clientID, getPeriodID(st.flushPeriod))
	exists, err := st.r.Exists(ctx, curListKey).Result()
	if err != nil {
		return "", errors.Wrap(err, "cannot check if a list exists")
	}

	if exists != 0 {
		return curListKey, nil
	}

	listnameKey := getListnameKey(clientID)
	prevListKey, err := st.r.Get(ctx, listnameKey).Result()
	if err != nil {
		return "", errors.Wrap(err, "cannot get previous list name")
	}

	if prevListKey == "" {
		return curListKey, nil
	}

	minScore := time.Now().Add(st.msgTTL * -1).Unix()
	_, err = st.r.TxPipelined(ctx, func(tx redis.Pipeliner) error {
		_, err = tx.Rename(ctx, prevListKey, curListKey).Result()
		if err != nil {
			return errors.Wrap(err, "cannot create a new list")
		}
		_, err = tx.Expire(ctx, curListKey, st.msgTTL).Result()
		if err != nil {
			return errors.Wrap(err, "cannot set expire for a list")
		}

		_, err = tx.ZRemRangeByScore(ctx, curListKey, fmt.Sprintf("%d", minScore), "").Result()
		if err != nil {
			return errors.Wrap(err, "cannot del lists")
		}

		_, err := tx.Set(ctx, listnameKey, curListKey, st.msgTTL).Result()
		if err != nil {
			return errors.Wrap(err, "cannot save a new list name")
		}
		return nil
	})

	if err != nil {
		return "", err
	}
	return curListKey, nil
}

func getPeriodID(flushPeriod time.Duration) int64 {
	return time.Now().Unix() / int64(flushPeriod.Seconds())
}

func getListKey(clientID int64, periodID int64) string {
	return fmt.Sprintf("%s:%d:%d", listKeyPref, clientID, periodID)
}

func getListnameKey(clientID int64) string {
	return fmt.Sprintf("%s:%d", listKeyPref, clientID)
}

func msgsToProto(msgs []string) ([]*pb.Msg, error) {
	msgsProto := make([]*pb.Msg, 0, len(msgs))
	for _, msg := range msgs {
		msgP := &pb.Msg{}
		err := proto.Unmarshal([]byte(msg), msgP)
		if err != nil {
			return nil, err
		}
		msgsProto = append(msgsProto, msgP)
	}

	return msgsProto, nil
}
