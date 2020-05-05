package worker

import (
	"github.com/go-redis/redis/v7"
	"github.com/google/uuid"
)

type RedisEventStore struct {
	client redis.Cmdable
}

func NewRedisEventStore() RedisEventStore {
	return RedisEventStore{NewRedisClient()}
}

func (es RedisEventStore) CreateEvent(event Event) error {
	eventKey := "event:" + uuid.New().String()

	fields := []interface{}{
		"type", event.Type,
		"title", event.Title,
		"message", event.Message,
	}
	if err := es.client.HSet(eventKey, fields...).Err(); err != nil {
		return err
	}

	eventQueue := "events:notif"
	if err := es.client.LPush(eventQueue, eventKey).Err(); err != nil {
		return err
	}

	return nil
}
