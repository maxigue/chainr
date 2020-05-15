package worker

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v7"

	"github.com/Tyrame/chainr/notif/internal/notifier"
)

type RedisEventStore struct {
	client redis.Cmdable
}

func NewRedisEventStore() RedisEventStore {
	addrs := []string{"chainr-redis:6379"}
	masterName := ""
	password := ""
	db := 0
	if val, ok := os.LookupEnv("REDIS_ADDR"); ok {
		addrs = []string{val}
	}
	if val, ok := os.LookupEnv("REDIS_ADDRS"); ok {
		addrs = strings.Split(val, " ")
	}
	if val, ok := os.LookupEnv("REDIS_MASTER"); ok {
		masterName = val
	}
	if val, ok := os.LookupEnv("REDIS_PASSWORD"); ok {
		password = val
	}
	if val, ok := os.LookupEnv("REDIS_DB"); ok {
		d, err := strconv.Atoi(val)
		if err != nil {
			log.Println("Invalid REDIS_DB value " + val + ", using default 0")
			d = 0
		}
		db = d
	}

	client := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:      addrs,
		MasterName: masterName,
		Password:   password,
		DB:         db,
	})

	return RedisEventStore{client}
}

func (rs RedisEventStore) NextEvent() (notifier.Event, error) {
	eventQueue := "events:notif"

	vals, err := rs.client.BRPop(0, eventQueue).Result()
	if err != nil {
		return notifier.Event{}, err
	}

	eventKey := vals[1]
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
