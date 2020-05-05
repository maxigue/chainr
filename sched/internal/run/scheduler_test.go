package run

import (
	"testing"

	"os"

	"github.com/go-redis/redis/v7"
)

func TestNewScheduler(t *testing.T) {
	s := NewScheduler().(*RedisScheduler)
	client := s.client.(*redis.Client)
	expected := "Redis<chainr-redis:6379 db:0>"
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

	expected := "Redis<chainr-redis:6379 db:0>"
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
	expectLPushK []string
	expectRPushK []string
	*redis.Client
}

func newRedisClientMock(t *testing.T) redis.Cmdable {
	return &redisClientMock{
		t: t,
		expectHSetK: []string{
			"run:abc",
			"job:job1:run:abc",
			"job:job2:run:abc",
			"dependency:0:job:job2:run:abc",
			"dependency:1:job:job2:run:abc",
		},
		expectSAddK: []string{
			"dependencies:job:job2:run:abc",
		},
		expectLPushK: []string{
			"runs",
			"runs:work",
		},
		expectRPushK: []string{
			"jobs:run:abc",
		},
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
		case "dependency:0:job:job2:run:abc":
			expectedValues = []string{"job", "job:job1:run:abc", "failure", "true"}
		case "dependency:1:job:job2:run:abc":
			expectedValues = []string{"job", "job:job42:run:abc", "failure", "false"}
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
		case "dependencies:job:job2:run:abc":
			expectedValues = []string{"dependency:0:job:job2:run:abc", "dependency:1:job:job2:run:abc"}
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
	if i, ok := indexOf(c.expectLPushK, key); ok {
		expectedValues := []string{}
		switch c.expectLPushK[i] {
		case "runs":
			expectedValues = []string{"run:abc"}
		case "runs:work":
			expectedValues = []string{"run:abc"}
		}
		vals := make([]string, len(values))
		for i, v := range values {
			vals[i] = v.(string)
		}
		if !equalsUnordered(vals, expectedValues) {
			c.t.Errorf("LPush: values = %v, expected %v", vals, expectedValues)
		}
		c.expectLPushK = remove(c.expectLPushK, i)
	} else {
		c.t.Errorf("LPush: unexpected key %v", key)
	}
	return redis.NewIntCmd()
}
func (c *redisClientMock) RPush(key string, values ...interface{}) *redis.IntCmd {
	if i, ok := indexOf(c.expectRPushK, key); ok {
		expectedValues := []string{}
		switch c.expectRPushK[i] {
		case "jobs:run:abc":
			expectedValues = []string{"job:job1:run:abc", "job:job2:run:abc"}
		}
		vals := make([]string, len(values))
		for i, v := range values {
			vals[i] = v.(string)
		}
		if !equalsUnordered(vals, expectedValues) {
			c.t.Errorf("RPush: values = %v, expected %v", vals, expectedValues)
		}
		c.expectRPushK = remove(c.expectRPushK, i)
	} else {
		c.t.Errorf("RPush: unexpected key %v", key)
	}
	return redis.NewIntCmd()
}
func (c *redisClientMock) LRange(key string, start, stop int64) *redis.StringSliceCmd {
	if start != 0 || stop != -1 {
		c.t.Errorf("LRange: start=%v, stop=%v, expected start=0, stop=-1", start, stop)
	}

	vals := make([]string, 0)
	switch key {
	case "runs":
		vals = append(vals, "run:abc")
	case "jobs:run:abc":
		vals = append(vals, "job:job1:run:abc", "job:job2:run:abc")
	default:
		c.t.Errorf("LRange: unexpected key %v", key)
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

	// job2 is before job1 in the map to ensure the ordering is done correctly.
	// After ordering, job1 should always be before job2 in the list.
	spec := []byte(`{
		"kind": "Pipeline",
		"jobs": {
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
			},
			"job1": {
				"image": "busybox",
				"run": "exit 0"
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
	if status.Jobs[0].Name != "job1" {
		t.Errorf("status.Jobs[0].Name = %v, expected job1", status.Jobs[0].Name)
	}
	if status.Jobs[0].Status != "PENDING" {
		t.Errorf("status.Jobs[0].Status = %v, expected PENDING", status.Jobs[0].Status)
	}
	if status.Jobs[1].Name != "job2" {
		t.Errorf("status.Jobs[1].Name = %v, expected job2", status.Jobs[1].Name)
	}
	if status.Jobs[1].Status != "PENDING" {
		t.Errorf("status.Jobs[1].Status = %v, expected PENDING", status.Jobs[1].Status)
	}

	client := s.client.(*redisClientMock)
	if len(client.expectHSetK) > 0 {
		t.Errorf("missing calls to HSet: %v", client.expectHSetK)
	}
	if len(client.expectSAddK) > 0 {
		t.Errorf("missing calls to SAdd: %v", client.expectSAddK)
	}
	if len(client.expectLPushK) > 0 {
		t.Errorf("missing calls to LPush: %v", client.expectLPushK)
	}
	if len(client.expectRPushK) > 0 {
		t.Errorf("missing calls to RPush: %v", client.expectRPushK)
	}
}

// This function, though private, is complex enough to need its own unit test.
func TestSortJobs(t *testing.T) {
	jobs := map[string]Job{
		"job1": Job{
			DependsOn: []JobDependency{
				JobDependency{Job: "job2"},
				JobDependency{Job: "job3"},
			},
		},
		"job2": Job{
			DependsOn: []JobDependency{},
		},
		"job3": Job{
			DependsOn: []JobDependency{
				JobDependency{Job: "job2"},
			},
		},
	}

	expectedOrder := []string{"job2", "job3", "job1"}
	sorted := sortJobs(jobs)
	for i := range sorted {
		if sorted[i].Name != expectedOrder[i] {
			t.Errorf("sorted[%v].Name = %v, expected %v", i, sorted[i].Name, expectedOrder[i])
		}
	}
}

func TestSortJobsNotFound(t *testing.T) {
	jobs := map[string]Job{
		"job1": Job{
			DependsOn: []JobDependency{
				JobDependency{Job: "job2"},
				JobDependency{Job: "notfound"},
			},
		},
		"job2": Job{
			DependsOn: []JobDependency{
				JobDependency{Job: "notfound"},
			},
		},
	}

	expectedOrder := []string{"job2", "job1"}
	sorted := sortJobs(jobs)
	for i := range sorted {
		if sorted[i].Name != expectedOrder[i] {
			t.Errorf("sorted[%v].Name = %v, expected %v", i, sorted[i].Name, expectedOrder[i])
		}
	}
}

func TestSortJobsLoop(t *testing.T) {
	jobs := map[string]Job{
		"job1": Job{
			DependsOn: []JobDependency{
				JobDependency{Job: "job2"},
			},
		},
		"job2": Job{
			DependsOn: []JobDependency{
				JobDependency{Job: "job1"},
			},
		},
	}

	// The order does not really matter here, the test
	// ensures that there is no infinite loop.
	_ = sortJobs(jobs)
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
	if status.Jobs[0].Name != "job1" {
		t.Errorf("status.Jobs[0].Name = %v, expected job1", status.Jobs[0].Name)
	}
	if status.Jobs[0].Status != "RUNNING" {
		t.Errorf("status.Jobs[0].Status = %v, expected RUNNING", status.Jobs[0].Status)
	}
	if status.Jobs[1].Name != "job2" {
		t.Errorf("status.Jobs[1].Name = %v, expected job2", status.Jobs[1].Name)
	}
	if status.Jobs[1].Status != "PENDING" {
		t.Errorf("status.Jobs[1].Status = %v, expected PENDING", status.Jobs[1].Status)
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

func TestStatusList(t *testing.T) {
	s := RedisScheduler{newRedisClientMock(t)}
	statusList, err := s.StatusList()
	if err != nil {
		t.Fatal(err)
	}

	if statusList[0].RunUID != "abc" {
		t.Errorf("statusList[0].RunUID = %v, expected run1", statusList[0].RunUID)
	}
	if statusList[0].Status.Run != "RUNNING" {
		t.Errorf("statusList[0].Status.Run = %v, expected RUNNING", statusList[0].Status.Run)
	}
	if statusList[0].Status.Jobs[0].Name != "job1" {
		t.Errorf("statusList[0].Status.Jobs[0].Name = %v, expected job1", statusList[0].Status.Jobs[0].Name)
	}
	if statusList[0].Status.Jobs[0].Status != "RUNNING" {
		t.Errorf("statusList[0].Status.Jobs[0].Status = %v, expected RUNNING", statusList[0].Status.Jobs[0].Status)
	}
	if statusList[0].Status.Jobs[1].Name != "job2" {
		t.Errorf("statusList[0].Status.Jobs[1].Name = %v, expected job2", statusList[0].Status.Jobs[1].Name)
	}
	if statusList[0].Status.Jobs[1].Status != "PENDING" {
		t.Errorf("statusList[0].Status.Jobs[1].Status = %v, expected PENDING", statusList[0].Status.Jobs[1].Status)
	}
}
