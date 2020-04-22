package pipeline

import (
	"testing"

	"os"
	"time"

	"github.com/go-redis/redis/v7"
)

func TestJSONSchemaError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("JSON schema initialization did not panic")
		}
	}()

	pipelineSchema = `{test}`
	initJSONSchema()
}

func TestRedisClient(t *testing.T) {
	old := redisClient
	defer func() {
		redisClient = old
	}()

	initRedisClient()
	client := redisClient.(*redis.Client)
	expected := "Redis<redis:6379 db:0>"
	if client.String() != expected {
		t.Errorf("client = %v, expected %v", client, expected)
	}
}

func TestRedisClientWithEnv(t *testing.T) {
	old := redisClient
	defer func() {
		redisClient = old
	}()

	if err := os.Setenv("REDIS_ADDR", "test:1234"); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("REDIS_PASSWORD", "passw0rd"); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("REDIS_DB", "1"); err != nil {
		t.Fatal(err)
	}
	initRedisClient()
	client := redisClient.(*redis.Client)

	expected := "Redis<test:1234 db:1>"
	if client.String() != expected {
		t.Errorf("client = %v, expected %v", client, expected)
	}
}

func TestRedisClientWithEnvError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Redis client initialization did not panic")
		}
	}()

	old := redisClient
	defer func() {
		redisClient = old
	}()

	if err := os.Setenv("REDIS_DB", "test"); err != nil {
		t.Fatal(err)
	}
	initRedisClient()
}

func TestNew(t *testing.T) {
	p := New().(*pipeline)
	if p.Kind != "Pipeline" {
		t.Errorf("p.Kind = %v, expected Pipeline", p.Kind)
	}
	if p.Jobs == nil {
		t.Errorf("p.Jobs is nil, expected map")
	}
}

func TestNewFromSpec(t *testing.T) {
	spec := []byte(`{
		"kind": "Pipeline",
		"jobs": {
			"job1": {
				"image": "busybox",
				"run": "exit 0"
			}
		}
	}`)

	pip, err := NewFromSpec(spec)
	if err != nil {
		t.Fatal("err = nil, expected not nil")
	}
	p := pip.(*pipeline)
	if image := p.Jobs["job1"].Image; image != "busybox" {
		t.Errorf("image = %v, expected busybox", image)
	}
}

func TestNewFromSpecBadFormat(t *testing.T) {
	spec := []byte(`{invalid}`)
	_, err := NewFromSpec(spec)
	if err == nil {
		t.Fatal("NewFromSpec from an invalid format returned a nil error")
	}
}

func TestNewFromSpecBadSchema(t *testing.T) {
	spec := []byte(`{
		"kind": "Pipeline",
		"invalid": "hello"
	}`)
	_, err := NewFromSpec(spec)
	if err == nil {
		t.Fatal("NewFromSpec from an invalid schema returned a nil error")
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
		if value != "0" {
			c.t.Errorf("value = %v, expected 0", value)
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
			expectedValues = []string{"image", "busybox", "run", "exit 0"}
		case "job:job2:run:abc":
			expectedValues = []string{"image", "busybox", "run", "exit 1"}
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

func TestRun(t *testing.T) {
	redisClient = newRedisClientMock(t)

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
	p, err := NewFromSpec(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Run("abc"); err != nil {
		t.Fatal(err)
	}

	client := redisClient.(*redisClientMock)
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
