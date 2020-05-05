package worker

import (
	"testing"

	"errors"
	"os"
	"time"

	"github.com/go-redis/redis/v7"
)

func TestNewRedisEventStore(t *testing.T) {
	es := NewRedisEventStore()
	client := es.client.(*redis.Client)
	expected := "Redis<chainr-redis:6379 db:0>"
	if client.String() != expected {
		t.Errorf("client = %v, expected %v", client, expected)
	}
}

func TestNewRedisEventStoreWithEnv(t *testing.T) {
	if err := os.Setenv("REDIS_ADDR", "test:1234"); err != nil {
		t.Fatal(err)
	}
	defer os.Unsetenv("REDIS_ADDR")
	if err := os.Setenv("REDIS_PASSWORD", "passw0rd"); err != nil {
		t.Fatal(err)
	}
	defer os.Unsetenv("REDIS_PASSWORD")
	if err := os.Setenv("REDIS_DB", "1"); err != nil {
		t.Fatal(err)
	}
	defer os.Unsetenv("REDIS_DB")
	es := NewRedisEventStore()
	client := es.client.(*redis.Client)

	expected := "Redis<test:1234 db:1>"
	if client.String() != expected {
		t.Errorf("client = %v, expected %v", client, expected)
	}
}

// In case of error reading the redis database, the default database is used.
func TestNewRedisEventStoreWithEnvError(t *testing.T) {
	if err := os.Setenv("REDIS_DB", "test"); err != nil {
		t.Fatal(err)
	}
	defer os.Unsetenv("REDIS_DB")
	es := NewRedisEventStore()
	client := es.client.(*redis.Client)

	expected := "Redis<chainr-redis:6379 db:0>"
	if client.String() != expected {
		t.Errorf("client = %v, expected %v", client, expected)
	}
}

type redisClientMock struct {
	t *testing.T
	*redis.Client
}

type redisClientStub struct {
	*redis.Client
}

type nextEventClientMock redisClientMock

func (c nextEventClientMock) BRPop(timeout time.Duration, keys ...string) *redis.StringSliceCmd {
	if timeout != 0 {
		c.t.Errorf("BRPop should block indefinitely")
	}
	if len(keys) != 1 || keys[0] != "events:notif" {
		c.t.Errorf("BRPop listens on %v, expected events:notif", keys[0])
	}

	vals := []string{
		"events:notif",
		"event:abc",
	}
	return redis.NewStringSliceResult(vals, nil)
}

func (c nextEventClientMock) HGetAll(key string) *redis.StringStringMapCmd {
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

func TestNextEvent(t *testing.T) {
	es := RedisEventStore{&nextEventClientMock{t: t}}
	event, err := es.NextEvent()
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

type nextEventClientErrorBRPopStub redisClientStub

func (c nextEventClientErrorBRPopStub) BRPop(timeout time.Duration, keys ...string) *redis.StringSliceCmd {
	return redis.NewStringSliceResult([]string{}, errors.New("BRPop failed"))
}

func (c nextEventClientErrorBRPopStub) HGetAll(key string) *redis.StringStringMapCmd {
	vals := map[string]string{
		"type":    "SUCCESS",
		"title":   "t",
		"message": "m",
	}
	return redis.NewStringStringMapResult(vals, nil)
}

func TestNextEventErrorBRPop(t *testing.T) {
	es := RedisEventStore{&nextEventClientErrorBRPopStub{}}
	_, err := es.NextEvent()
	if err.Error() != "BRPop failed" {
		t.Errorf("redis error was not forwarded")
	}
}

type nextEventClientErrorHGetAllStub redisClientStub

func (c nextEventClientErrorHGetAllStub) BRPop(timeout time.Duration, keys ...string) *redis.StringSliceCmd {
	vals := []string{
		"events:notif",
		"event:abc",
	}
	return redis.NewStringSliceResult(vals, nil)
}

func (c nextEventClientErrorHGetAllStub) HGetAll(key string) *redis.StringStringMapCmd {
	return redis.NewStringStringMapResult(make(map[string]string), errors.New("HGetAll failed"))
}

func TestNextEventErrorHGetAll(t *testing.T) {
	es := RedisEventStore{&nextEventClientErrorHGetAllStub{}}
	_, err := es.NextEvent()
	if err.Error() != "HGetAll failed" {
		t.Errorf("redis error was not forwarded")
	}
}
