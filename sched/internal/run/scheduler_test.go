package run

import (
	"testing"

	"os"
	"time"

	"github.com/go-redis/redis/v7"
)

func TestNewScheduler(t *testing.T) {
	s := NewScheduler().(*RedisScheduler)
	client := s.client.(*redis.Client)
	expected := "Redis<redis:6379 db:0>"
	if client.String() != expected {
		t.Errorf("client = %v, expected %v", client, expected)
	}
}

func TestNewSchedulerWithEnv(t *testing.T) {
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
	s := NewScheduler().(*RedisScheduler)
	client := s.client.(*redis.Client)

	expected := "Redis<test:1234 db:1>"
	if client.String() != expected {
		t.Errorf("client = %v, expected %v", client, expected)
	}
}

// In case of error reading the redis database, the default database is used.
func TestNewSchedulerWithEnvError(t *testing.T) {
	if err := os.Setenv("REDIS_DB", "test"); err != nil {
		t.Fatal(err)
	}
	defer os.Unsetenv("REDIS_DB")
	s := NewScheduler().(*RedisScheduler)
	client := s.client.(*redis.Client)

	expected := "Redis<redis:6379 db:0>"
	if client.String() != expected {
		t.Errorf("client = %v, expected %v", client, expected)
	}
}

func indexOf(a []string, s string) (int, bool) {
	for i, v := range a {
		if v == s {
			return i, true
		}
	}
	return 0, false
}
func remove(a []string, i int) []string {
	a[len(a)-1], a[i] = a[i], a[len(a)-1]
	return a[:len(a)-1]
}
func equals(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, _ := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

type redisClientMock struct {
	t            *testing.T
	expectSetK   []string
	expectHSetK  []string
	expectLPushV []string
	*redis.Client
}

func newRedisClientMock(t *testing.T) redis.Cmdable {
	return &redisClientMock{
		t:          t,
		expectSetK: []string{"run:abc"},
		expectHSetK: []string{
			"job:job1:run:abc",
			"job:job2:run:abc",
			"dependency:job1:job:job2:run:abc",
			"dependency:job42:job:job2:run:abc",
		},
		expectLPushV: []string{"job:job1:run:abc", "job:job2:run:abc"},
	}
}
func (c *redisClientMock) Set(key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	if i, ok := indexOf(c.expectSetK, key); ok {
		if value != "abc" {
			c.t.Errorf("value = %v, expected abc", value)
		}
		if expiration != 0 {
			c.t.Errorf("expiration = %v, expected 0", expiration)
		}
		c.expectSetK = remove(c.expectSetK, i)
	} else {
		c.t.Errorf("unexpected key %v", key)
	}
	return redis.NewStatusCmd()
}
func (c *redisClientMock) HSet(key string, values ...interface{}) *redis.IntCmd {
	if i, ok := indexOf(c.expectHSetK, key); ok {
		expectedValues := []string{}
		switch c.expectHSetK[i] {
		case "job:job1:run:abc":
			expectedValues = []string{"name", "job1", "image", "busybox", "run", "exit 0", "status", "PENDING"}
		case "job:job2:run:abc":
			expectedValues = []string{"name", "job2", "image", "busybox", "run", "exit 1", "status", "PENDING"}
		case "dependency:job1:job:job2:run:abc":
			expectedValues = []string{"failure", "true"}
		case "dependency:job42:job:job2:run:abc":
			expectedValues = []string{"failure", "false"}
		}
		vals := make([]string, len(values))
		for i, v := range values {
			vals[i] = v.(string)
		}
		if !equals(vals, expectedValues) {
			c.t.Errorf("values = %v, expected %v", vals, expectedValues)
		}
		c.expectHSetK = remove(c.expectHSetK, i)
	} else {
		c.t.Errorf("unexpected key %v", key)
	}
	return redis.NewIntCmd()
}
func (c *redisClientMock) LPush(key string, values ...interface{}) *redis.IntCmd {
	if key != "work:jobs" {
		c.t.Errorf("unexpected key %v", key)
	}
	for _, val := range values {
		v := val.(string)
		if i, ok := indexOf(c.expectLPushV, v); ok {
			c.expectLPushV = remove(c.expectLPushV, i)
		} else {
			c.t.Errorf("unexpected value %v", v)
		}
	}
	return redis.NewIntCmd()
}
func (c *redisClientMock) Keys(pattern string) *redis.StringSliceCmd {
	vals := make([]string, 0)
	switch pattern {
	case "run:*":
		vals = append(vals, "run:abc")
	case "job:*:run:abc":
		vals = append(vals, "job:job1:run:abc")
		vals = append(vals, "job:job2:run:abc")
	case "job:*:run:notfound":
	default:
		c.t.Errorf("unexpected pattern %v", pattern)
	}
	return redis.NewStringSliceResult(vals, nil)
}
func (c *redisClientMock) Get(key string) *redis.StringCmd {
	if key != "run:abc" {
		c.t.Errorf("unexpected key %v, expected run:abc", key)
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
		vals["status"] = "RUNNING"
	case "job:job2:run:abc":
		vals["name"] = "job2"
		vals["image"] = "busybox"
		vals["run"] = "exit 1"
		vals["status"] = "WAITING"
	default:
		c.t.Errorf("unexpected key %v", key)
	}

	return redis.NewStringStringMapResult(vals, nil)
}

func TestSchedule(t *testing.T) {
	s := RedisScheduler{newRedisClientMock(t)}

	spec := []byte(`{
		"kind": "Pipeline",
		"jobs": {
			"job1": {
				"image": "busybox",
				"run": "exit 0"
			},
			"job2": {
				"image": "busybox",
				"run": "exit 1",
				"dependsOn": [{
					"job": "job1",
					"conditions": {
						"failure": true
					}
				}, {
					"job": "job42"
				}]
			}
		}
	}`)
	run, err := New(spec)
	run.Metadata.UID = "abc"
	run.Metadata.SelfLink = "/api/runs/" + run.Metadata.UID
	if err != nil {
		t.Fatal(err)
	}

	status, err := s.Schedule(run)
	if err != nil {
		t.Fatal(err)
	}

	if status["job1"] != "PENDING" {
		t.Errorf("status[job1] = %v, expected PENDING", status["job1"])
	}
	if status["job2"] != "PENDING" {
		t.Errorf("status[job2] = %v, expected PENDING", status["job1"])
	}

	client := s.client.(*redisClientMock)
	if len(client.expectSetK) > 0 {
		t.Errorf("missing calls to Set: %v", client.expectSetK)
	}
	if len(client.expectHSetK) > 0 {
		t.Errorf("missing calls to HSet: %v", client.expectHSetK)
	}
	if len(client.expectLPushV) > 0 {
		t.Errorf("missing calls to LPush: %v", client.expectLPushV)
	}
}

func TestStatus(t *testing.T) {
	s := RedisScheduler{newRedisClientMock(t)}
	status, err := s.Status("abc")
	if err != nil {
		t.Fatal(err)
	}

	if status["job1"] != "RUNNING" {
		t.Errorf("status[job1] = %v, expected RUNNING", status["job1"])
	}
	if status["job2"] != "WAITING" {
		t.Errorf("status[job2] = %v, expected WAITING", status["job2"])
	}
}

func TestStatusNotFound(t *testing.T) {
	s := RedisScheduler{newRedisClientMock(t)}
	_, err := s.Status("notfound")
	e := err.(*NotFoundError)
	if e.RunUID != "notfound" {
		t.Errorf("e.RunID = %v, expected notfound", e.RunUID)
	}
}

func TestStatusMap(t *testing.T) {
	s := RedisScheduler{newRedisClientMock(t)}
	statusMap, err := s.StatusMap()
	if err != nil {
		t.Fatal(err)
	}

	if statusMap["abc"]["job1"] != "RUNNING" {
		t.Errorf("statusMap[abc][job1] = %v, expected RUNNING", statusMap["abc"]["job1"])
	}
	if statusMap["abc"]["job2"] != "WAITING" {
		t.Errorf("statusMap[abc][job2] = %v, expected WAITING", statusMap["abc"]["job2"])
	}
}
