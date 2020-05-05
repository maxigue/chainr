package worker

import (
	"testing"

	"errors"
	"strings"

	"github.com/go-redis/redis/v7"
)

func TestNewRedisEventStore(t *testing.T) {
	// Test that NewRedisEventStore does not panic.
	_ = NewRedisEventStore()
}

type createEventClientMock redisClientMock

func (c createEventClientMock) HSet(key string, values ...interface{}) *redis.IntCmd {
	if strings.HasPrefix(key, "event:") == false {
		c.t.Errorf("HSet: key = %v, expected string starting with event:", key)
	}

	failure := false
	expectedValues := []string{"type", "SUCCESS", "title", "t", "message", "m"}
	if len(values) != len(expectedValues) {
		failure = true
	} else {
		for i := range values {
			if values[i] != expectedValues[i] {
				failure = true
			}
		}
	}
	if failure {
		c.t.Errorf("HSet: values = %v, expected %v", values, expectedValues)
	}

	return redis.NewIntResult(1, nil)
}

func (c createEventClientMock) LPush(key string, values ...interface{}) *redis.IntCmd {
	return redis.NewIntResult(0, nil)
}

func TestCreateEvent(t *testing.T) {
	es := RedisEventStore{&createEventClientMock{t: t}}
	if err := es.CreateEvent(Event{"SUCCESS", "t", "m"}); err != nil {
		t.Errorf("err = %v, expected nil", err)
	}
}

type createEventClientErrorHSetStub redisClientStub

func (c createEventClientErrorHSetStub) HSet(key string, values ...interface{}) *redis.IntCmd {
	return redis.NewIntResult(0, errors.New("HSet failed"))
}

func (c createEventClientErrorHSetStub) LPush(key string, values ...interface{}) *redis.IntCmd {
	return redis.NewIntResult(0, nil)
}

func TestCreateEventErrorHSet(t *testing.T) {
	es := RedisEventStore{&createEventClientErrorHSetStub{}}
	err := es.CreateEvent(Event{"SUCCESS", "t", "m"})
	if err.Error() != "HSet failed" {
		t.Errorf("err.Error() = %v, expected HSet failed", err.Error())
	}
}

type createEventClientErrorLPushStub redisClientStub

func (c createEventClientErrorLPushStub) HSet(key string, values ...interface{}) *redis.IntCmd {
	return redis.NewIntResult(1, nil)
}

func (c createEventClientErrorLPushStub) LPush(key string, values ...interface{}) *redis.IntCmd {
	return redis.NewIntResult(0, errors.New("LPush failed"))
}

func TestCreateEventErrorLPush(t *testing.T) {
	es := RedisEventStore{&createEventClientErrorLPushStub{}}
	err := es.CreateEvent(Event{"SUCCESS", "t", "m"})
	if err.Error() != "LPush failed" {
		t.Errorf("err.Error() = %v, expected LPush failed", err.Error())
	}
}
