package input

import (
	"context"
	"pull-notify/pkg/pb"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

type NotificationRepo interface {
	SaveMsg(ctx context.Context, msg *pb.Msg) error
}

type InputQueue struct {
	r     *redis.Client
	repo  NotificationRepo
	queue string
}

func NewInputQueue(r *redis.Client, repo NotificationRepo, inputQueue string) *InputQueue {
	return &InputQueue{r, repo, inputQueue}
}

func (q *InputQueue) Run(ctx context.Context) {
	logrus.Info("starting redis suscriber")
	sub := q.r.Subscribe(ctx, q.queue)
	defer func() {
		if err := sub.Unsubscribe(ctx, q.queue); err != nil {
			logrus.WithError(err).Errorf("cannot unsubscribe %v", err)
		}
	}()

	for {
		select {
		case msg := <-sub.Channel():
			if err := q.handle(ctx, msg); err != nil {
				logrus.WithError(err).Errorf("cannot handle message %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (q *InputQueue) handle(ctx context.Context, msg *redis.Message) error {
	logrus.Debugf("got message %s", msg)
	msgProto := &pb.Msg{}
	err := proto.Unmarshal([]byte(msg.Payload), msgProto)
	if err != nil {
		return errors.Wrap(err, "cannot unmarshal message")
	}

	return q.repo.SaveMsg(ctx, msgProto)
}
