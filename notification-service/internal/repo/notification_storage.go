package repo

import (
	"context"
	"fmt"
	"notification-service/internal/config"
	"notification-service/pkg/pb"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
	rangeMsgs := &redis.ZRangeBy{Min: fmt.Sprintf("(%d", offset), Max: "+Inf", Count: int64(lastMsgLimit)}

	return st.getMsgs(ctx, clientID, rangeMsgs)
}

func (st *NotifyStorage) GetMsgList(ctx context.Context, clientID int64) ([]*pb.Msg, error) {
	rangeMsgs := &redis.ZRangeBy{Max: "+Inf", Min: "-Inf"}

	return st.getMsgs(ctx, clientID, rangeMsgs)
}

func (st *NotifyStorage) getMsgs(ctx context.Context, clientID int64, r *redis.ZRangeBy) ([]*pb.Msg, error) {
	listKey := getListKey(clientID, getPeriodID(st.flushPeriod))
	var msgsCMD *redis.StringSliceCmd
	var existsCMD *redis.IntCmd
	_, err := st.r.TxPipelined(ctx, func(tx redis.Pipeliner) error {
		existsCMD = tx.Exists(ctx, listKey)
		msgsCMD = tx.ZRevRangeByScore(ctx, listKey, r)

		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "cannot get msgs")
	}

	msgs, err := msgsCMD.Result()
	if err != nil {
		return nil, errors.Wrap(err, "cannot get msgs result")
	}
	if len(msgs) == 0 {
		exists, err := existsCMD.Result()
		if err != nil {
			return nil, err
		}
		if exists == 0 {
			st.refreshList(ctx, clientID)
			return nil, nil
		}
	}

	protoMsgs, err := msgsToProto(msgs)
	if err != nil {
		return nil, errors.Wrap(err, "cannot convert msgs to proto")
	}

	return protoMsgs, nil
}

func (st *NotifyStorage) MarkRead(ctx context.Context, clientID int64, msgID string, offset int64) error {
	listKey := getListKey(clientID, getPeriodID(st.flushPeriod))

	var msgs []string
	var err error
	for ok := true; ok; ok = false {
		var msgsCMD *redis.StringSliceCmd
		var existsCMD *redis.IntCmd
		_, err := st.r.TxPipelined(ctx, func(tx redis.Pipeliner) error {
			existsCMD = tx.Exists(ctx, listKey)
			msgsCMD = tx.ZRangeByScore(ctx, listKey, &redis.ZRangeBy{Min: fmt.Sprintf("%d", offset), Max: fmt.Sprintf("%d", offset)})

			return nil
		})
		if err != nil {
			return errors.Wrap(err, "cannot get msgs")
		}

		msgs, err := msgsCMD.Result()
		if err != nil {
			return errors.Wrap(err, "cannot get msgs result")
		}
		if len(msgs) == 0 {
			exists, err := existsCMD.Result()
			if err != nil {
				return err
			}
			if exists == 0 {
				ok = true
				st.refreshList(ctx, clientID)
				return nil
			}
		}
	}

	msgsProto, err := msgsToProto(msgs)
	if err != nil {
		return errors.Wrap(err, "cannot unmarshal msgs")
	}

	zmsgs := make([]*redis.Z, 0, len(msgs))
	for _, mp := range msgsProto {
		if mp.Id == msgID {
			mp.IsRead = true
		}
		mpMarshaled, err := proto.Marshal(mp)
		if err != nil {
			return errors.Wrap(err, "cannot marshal msg")
		}
		zmsgs = append(zmsgs, &redis.Z{Score: float64(offset), Member: mpMarshaled})
	}

	_, err = st.r.TxPipelined(ctx, func(tx redis.Pipeliner) error {
		_, err = tx.ZRemRangeByScore(ctx, listKey, fmt.Sprintf("%d", offset), fmt.Sprintf("%d", offset)).Result()
		if err != nil {
			return errors.Wrap(err, "cannot del lists")
		}

		_, err := tx.ZAddNX(ctx, listKey, zmsgs...).Result()
		if err != nil {
			return errors.Wrap(err, "cannot save an updated message")
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (st *NotifyStorage) SaveMsg(ctx context.Context, msg *pb.Msg) error {
	listKey := getListKey(msg.ClientId, getPeriodID(st.flushPeriod))
	exists, err := st.r.Exists(ctx, listKey).Result()
	if err != nil {
		return errors.Wrap(err, "cannot check if a list exists")
	}

	if exists == 0 {
		st.refreshList(ctx, msg.ClientId)
	}

	mpMarshaled, err := proto.Marshal(msg)
	if err != nil {
		return errors.Wrap(err, "cannot marshal msg")
	}

	_, err = st.r.ZAddNX(ctx, listKey, &redis.Z{Score: float64(msg.Created), Member: mpMarshaled}).Result()
	if err != nil {
		return errors.Wrap(err, "cannot save a message")
	}

	return nil
}

func (st *NotifyStorage) refreshList(ctx context.Context, clientID int64) (string, error) {
	listnameKey := getListnameKey(clientID)

	prevListKey, err := st.r.Get(ctx, listnameKey).Result()
	if err != nil && err != redis.Nil {
		return "", errors.Wrap(err, "cannot get previous list name")
	}

	curListKey := getListKey(clientID, getPeriodID(st.flushPeriod))
	if prevListKey == "" {
		_, err := st.r.Set(ctx, listnameKey, curListKey, st.msgTTL).Result()
		if err != nil {
			return "", errors.Wrap(err, "cannot save a new list name")
		}
		logrus.Debugf("list created %s", curListKey)
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

		_, err = tx.ZRemRangeByScore(ctx, curListKey, "", fmt.Sprintf("%d", minScore)).Result()
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
	logrus.Debugf("list recreated %s->%s", prevListKey, curListKey)
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
