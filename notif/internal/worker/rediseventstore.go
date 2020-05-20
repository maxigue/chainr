package worker

import (
	"github.com/go-redis/redis/v7"

	"github.com/Tyrame/chainr/notif/internal/notifier"
)

type RedisEventStore struct {
	info   Info
	client redis.Cmdable
}

func NewRedisEventStore(info Info) RedisEventStore {
	return RedisEventStore{info, NewRedisClient()}
}

func (rs RedisEventStore) NextEvent() (string, error) {
	return rs.client.BRPopLPush(rs.info.Queue, rs.info.ProcessQueue, 0).Result()
}

func (rs RedisEventStore) GetEvent(eventKey string) (notifier.Event, error) {
	event, err := rs.client.HGetAll(eventKey).Result()
	if err != nil {
		return notifier.Event{}, err
	}

	return notifier.Event{
		Type:    event["type"],
		Title:   event["title"],
		Message: event["message"],
	}, nil
}

func (rs RedisEventStore) Close(eventKey string) error {
	return rs.client.LRem(rs.info.ProcessQueue, -1, eventKey).Err()
}
