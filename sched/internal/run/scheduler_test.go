package run

import (
	"testing"

	"os"

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
func equalsUnordered(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for _, v := range a {
		if _, ok := indexOf(b, v); !ok {
			return false
		}
	}
	return true
}

type redisClientMock struct {
	t            *testing.T
	expectHSetK  []string
	expectSAddK  []string
	expectLPushV []string
	*redis.Client
}

func newRedisClientMock(t *testing.T) redis.Cmdable {
	return &redisClientMock{
		t: t,
		expectHSetK: []string{
			"run:abc",
			"job:job1:run:abc",
			"job:job2:run:abc",
			"dependency:job1:job:job2:run:abc",
			"dependency:job42:job:job2:run:abc",
		},
		expectSAddK: []string{
			"runs",
			"jobs:run:abc",
			"dependencies:job:job2:run:abc",
		},
		expectLPushV: []string{"run:abc"},
	}
}
func (c *redisClientMock) HSet(key string, values ...interface{}) *redis.IntCmd {
	if i, ok := indexOf(c.expectHSetK, key); ok {
		expectedValues := []string{}
		switch c.expectHSetK[i] {
		case "run:abc":
			expectedValues = []string{"uid", "abc", "status", "PENDING"}
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
			c.t.Errorf("HSet: values = %v, expected %v", vals, expectedValues)
		}
		c.expectHSetK = remove(c.expectHSetK, i)
	} else {
		c.t.Errorf("HSet: unexpected key %v", key)
	}
	return redis.NewIntCmd()
}
func (c *redisClientMock) SAdd(key string, members ...interface{}) *redis.IntCmd {
	if i, ok := indexOf(c.expectSAddK, key); ok {
		expectedValues := []string{}
		switch c.expectSAddK[i] {
		case "runs":
			expectedValues = []string{"run:abc"}
		case "jobs:run:abc":
			expectedValues = []string{"job:job1:run:abc", "job:job2:run:abc"}
		case "dependencies:job:job1:run:abc":
		case "dependencies:job:job2:run:abc":
			expectedValues = []string{"dependency:job1:job:job2:run:abc", "dependency:job42:job:job2:run:abc"}
		}
		vals := make([]string, len(members))
		for i, v := range members {
			vals[i] = v.(string)
		}
		if !equalsUnordered(vals, expectedValues) {
			c.t.Errorf("SAdd: members = %v, expected %v", vals, expectedValues)
		}
		c.expectSAddK = remove(c.expectSAddK, i)
	} else {
		c.t.Errorf("SAdd: unexpected key %v", key)
	}
	return redis.NewIntCmd()
}
func (c *redisClientMock) LPush(key string, values ...interface{}) *redis.IntCmd {
	if key != "runs:work" {
		c.t.Errorf("LPush: unexpected key %v", key)
	}
	for _, val := range values {
		v := val.(string)
		if i, ok := indexOf(c.expectLPushV, v); ok {
			c.expectLPushV = remove(c.expectLPushV, i)
		} else {
			c.t.Errorf("LPush: unexpected value %v", v)
		}
	}
	return redis.NewIntCmd()
}
func (c *redisClientMock) SMembers(key string) *redis.StringSliceCmd {
	vals := make([]string, 0)
	switch key {
	case "runs":
		vals = append(vals, "run:abc")
	case "jobs:run:abc":
		vals = append(vals, "job:job1:run:abc", "job:job2:run:abc")
	default:
		c.t.Errorf("SMembers: unexpected key %v", key)
	}
	return redis.NewStringSliceResult(vals, nil)
}
func (c *redisClientMock) HGet(key, field string) *redis.StringCmd {
	if key != "run:abc" {
		c.t.Errorf("HGet: unexpected key %v, expected run:abc", key)
	}
	if field != "uid" {
		c.t.Errorf("HGet: unexpected field %v, expected uid", field)
	}
	return redis.NewStringResult("abc", nil)
}
func (c *redisClientMock) HGetAll(key string) *redis.StringStringMapCmd {
	vals := make(map[string]string)
	switch key {
	case "run:abc":
		vals["uid"] = "abc"
		vals["status"] = "RUNNING"
	case "run:notfound":
	case "job:job1:run:abc":
		vals["name"] = "job1"
		vals["image"] = "busybox"
		vals["run"] = "exit 0"
		vals["status"] = "RUNNING"
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
	p, err := NewPipelineFactory().Create(spec)
	if err != nil {
		t.Fatal(err)
	}
	run := New(p)
	run.Metadata.UID = "abc"
	run.Metadata.SelfLink = "/api/runs/" + run.Metadata.UID

	status, err := s.Schedule(run)
	if err != nil {
		t.Fatal(err)
	}

	if status.Run != "PENDING" {
		t.Errorf("status.Run = %v, expected PENDING", status.Run)
	}
	if status.Jobs["job1"] != "PENDING" {
		t.Errorf("status.Jobs[job1] = %v, expected PENDING", status.Jobs["job1"])
	}
	if status.Jobs["job2"] != "PENDING" {
		t.Errorf("status.Jobs[job2] = %v, expected PENDING", status.Jobs["job1"])
	}

	client := s.client.(*redisClientMock)
	if len(client.expectHSetK) > 0 {
		t.Errorf("missing calls to HSet: %v", client.expectHSetK)
	}
	if len(client.expectSAddK) > 0 {
		t.Errorf("missing calls to SAdd: %v", client.expectSAddK)
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

	if status.Run != "RUNNING" {
		t.Errorf("status.Run = %v, expected RUNNING", status.Run)
	}
	if status.Jobs["job1"] != "RUNNING" {
		t.Errorf("status.Jobs[job1] = %v, expected RUNNING", status.Jobs["job1"])
	}
	if status.Jobs["job2"] != "PENDING" {
		t.Errorf("status.Jobs[job2] = %v, expected PENDING", status.Jobs["job2"])
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

	if statusMap["abc"].Run != "RUNNING" {
		t.Errorf("statusMap[abc].Run = %v, expected RUNNING", statusMap["abc"].Run)
	}
	if statusMap["abc"].Jobs["job1"] != "RUNNING" {
		t.Errorf("statusMap[abc].Jobs[job1] = %v, expected RUNNING", statusMap["abc"].Jobs["job1"])
	}
	if statusMap["abc"].Jobs["job2"] != "PENDING" {
		t.Errorf("statusMap[abc].Jobs[job2] = %v, expected PENDING", statusMap["abc"].Jobs["job2"])
	}
}
