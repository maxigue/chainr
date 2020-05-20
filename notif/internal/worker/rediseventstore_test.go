package worker

import (
	"testing"

	"errors"
	"time"

	"github.com/go-redis/redis/v7"
)

func TestNewRedisEventStore(t *testing.T) {
	// Test that NewRedisEventStore does not panic.
	_ = NewRedisEventStore(testInfo)
}

type nextEventClientMock redisClientMock

func (c nextEventClientMock) BRPopLPush(source, destination string, timeout time.Duration) *redis.StringCmd {
	if timeout != 0 {
		c.t.Errorf("BRPopLPush should block indefinitely")
	}
	if source != "events:notif" {
		c.t.Errorf("BRPopLPush listens on %v, expected events:notif", source)
	}
	if destination != "events:notifier:xyz" {
		c.t.Errorf("BRPopLPush listens on %v, expected events:notifier:xyz", destination)
	}

	return redis.NewStringResult("event:abc", nil)
}

func TestNextEvent(t *testing.T) {
	es := RedisEventStore{testInfo, &nextEventClientMock{t: t}}
	eventID, err := es.NextEvent()
	if err != nil {
		t.Fatal(err)
	}
	if eventID != "event:abc" {
		t.Errorf("eventID = %v, expected event:abc", eventID)
	}
}

type nextEventClientErrorStub redisClientStub

func (c nextEventClientErrorStub) BRPopLPush(source, destination string, timeout time.Duration) *redis.StringCmd {
	return redis.NewStringResult("", errors.New("BRPopLPush failed"))
}

func TestNextEventError(t *testing.T) {
	es := RedisEventStore{testInfo, &nextEventClientErrorStub{}}
	_, err := es.NextEvent()
	if err.Error() != "BRPopLPush failed" {
		t.Errorf("redis error was not forwarded")
	}
}

type getEventClientMock redisClientMock

func (c getEventClientMock) HGetAll(key string) *redis.StringStringMapCmd {
	if key != "event:abc" {
		c.t.Errorf("key = %v, expected event:abc", key)
	}

	vals := map[string]string{
		"type":    "SUCCESS",
		"title":   "t",
		"message": "m",
	}
	return redis.NewStringStringMapResult(vals, nil)
}

func TestGetEvent(t *testing.T) {
	es := RedisEventStore{testInfo, &getEventClientMock{t: t}}
	event, err := es.GetEvent("event:abc")
	if err != nil {
		t.Fatal(err)
	}

	if event.Type != "SUCCESS" {
		t.Errorf("event.Type = %v, expected SUCCESS", event.Type)
	}
	if event.Title != "t" {
		t.Errorf("event.Title = %v, expected t", event.Title)
	}
	if event.Message != "m" {
		t.Errorf("event.Message = %v, expected m", event.Message)
	}
}

type getEventClientErrorStub redisClientStub

func (c getEventClientErrorStub) HGetAll(key string) *redis.StringStringMapCmd {
	return redis.NewStringStringMapResult(make(map[string]string), errors.New("HGetAll failed"))
}

func TestGetEventError(t *testing.T) {
	es := RedisEventStore{testInfo, &getEventClientErrorStub{}}
	_, err := es.GetEvent("event:abc")
	if err.Error() != "HGetAll failed" {
		t.Errorf("redis error was not forwarded")
	}
}

type closeClientMock redisClientMock

func (c closeClientMock) LRem(key string, count int64, value interface{}) *redis.IntCmd {
	if key != "events:notifier:xyz" {
		c.t.Errorf("key = %v, expected events:notifier:xyz", key)
	}
	if count != -1 {
		c.t.Errorf("count = %v, expected -1 to remove only oldest element", count)
	}
	if value != "event:abc" {
		c.t.Errorf("value = %v, expected event:abc", value)
	}

	return redis.NewIntResult(1, nil)
}

func TestClose(t *testing.T) {
	es := RedisEventStore{testInfo, &closeClientMock{t: t}}
	err := es.Close("event:abc")
	if err != nil {
		t.Fatal(err)
	}
}

type closeClientErrorStub redisClientStub

func (c closeClientErrorStub) LRem(key string, count int64, value interface{}) *redis.IntCmd {
	return redis.NewIntResult(0, errors.New("LRem failed"))
}

func TestCloseError(t *testing.T) {
	es := RedisEventStore{testInfo, &closeClientErrorStub{}}
	err := es.Close("event:abc")
	if err.Error() != "LRem failed" {
		t.Errorf("redis error was not forwarded")
	}
}
