package worker

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"errors"
	"os"
	"time"

	"github.com/go-redis/redis/v7"
)

func TestNew(t *testing.T) {
	s := New().(*RedisWorker)
	client := s.client.(*redis.Client)
	expected := "Redis<redis:6379 db:0>"
	if client.String() != expected {
		t.Errorf("client = %v, expected %v", client, expected)
	}
}

func TestNewWithEnv(t *testing.T) {
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
	s := New().(*RedisWorker)
	client := s.client.(*redis.Client)

	expected := "Redis<test:1234 db:1>"
	if client.String() != expected {
		t.Errorf("client = %v, expected %v", client, expected)
	}
}

// In case of error reading the redis database, the default database is used.
func TestNewWithEnvError(t *testing.T) {
	if err := os.Setenv("REDIS_DB", "test"); err != nil {
		t.Fatal(err)
	}
	defer os.Unsetenv("REDIS_DB")
	s := New().(*RedisWorker)
	client := s.client.(*redis.Client)

	expected := "Redis<redis:6379 db:0>"
	if client.String() != expected {
		t.Errorf("client = %v, expected %v", client, expected)
	}
}

type redisClientMock struct {
	t           *testing.T
	popI        int
	expectHSetK []string
	*redis.Client
}

func newRedisClientMock(t *testing.T) redis.Cmdable {
	return &redisClientMock{
		t:    t,
		popI: 0,
		expectHSetK: []string{
			"run:abc",
			"job:job1:run:abc",
			"job:job2:run:abc",
		},
	}
}
func (c *redisClientMock) BRPop(timeout time.Duration, keys ...string) *redis.StringSliceCmd {
	// Fail on second call, to avoid infinite loop.
	if c.popI > 0 {
		return redis.NewStringSliceResult(nil, errors.New("fail"))
	}
	c.popI++

	if timeout != 0 {
		c.t.Errorf("BRPop should block indefinitely")
	}
	if len(keys) != 1 || keys[0] != "runs:work" {
		c.t.Errorf("BRPop listens on %v, expected runs:work", keys[0])
	}

	vals := []string{
		"runs:work",
		"run:abc",
	}
	return redis.NewStringSliceResult(vals, nil)
}
func (c *redisClientMock) HSet(key string, values ...interface{}) *redis.IntCmd {
	// TODO: implement mock
	return redis.NewIntCmd()
}
func (c *redisClientMock) HGet(key, field string) *redis.StringCmd {
	if key != "run:abc" {
		c.t.Errorf("HGet: key = %v, expected run:abc", key)
	}
	if field != "uid" {
		c.t.Errorf("HGet: field = %v, expected uid", field)
	}
	return redis.NewStringResult("abc", nil)
}
func (c *redisClientMock) HGetAll(key string) *redis.StringStringMapCmd {
	vals := make(map[string]string)
	switch key {
	case "job:job1:run:abc":
		vals["name"] = "job1"
		vals["image"] = "busybox"
		vals["run"] = "exit 0"
		vals["status"] = "PENDING"
	case "job:job2:run:abc":
		vals["name"] = "job2"
		vals["image"] = "busybox"
		vals["run"] = "exit 1"
		vals["status"] = "PENDING"
	default:
		c.t.Errorf("HGetAll: unexpected key %v", key)
	}

	return redis.NewStringStringMapResult(vals, nil)
}
func (c *redisClientMock) SMembers(key string) *redis.StringSliceCmd {
	vals := make([]string, 0)
	switch key {
	case "jobs:run:abc":
		vals = append(vals, "job:job1:run:abc", "job:job2:run:abc")
	default:
		c.t.Errorf("SMembers: unexpected key %v", key)
	}
	return redis.NewStringSliceResult(vals, nil)
}

func TestStart(t *testing.T) {
	w := RedisWorker{newRedisClientMock(t)}
	_ = w.Start()

	Convey("Scenario: process a run", t, func() {
		Convey("Given a run is scheduled", func() {
			Convey("The run should be set to status RUNNING", nil)
		})

		Convey("And there are no dependencies", func() {
			Convey("All jobs should be set to status RUNNING", nil)
			Convey("All jobs should be started on Kubernetes", nil)

			Convey("When the run succeeds", func() {
				Convey("All jobs should be set to status SUCCESSFUL", nil)
				Convey("The run should be set to status SUCCESSFUL", nil)
			})
		})
	})
}
