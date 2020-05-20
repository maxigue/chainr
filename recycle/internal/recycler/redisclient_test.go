package recycler

import (
	"testing"

	"os"

	"github.com/go-redis/redis/v7"
)

type redisClientMock struct {
	t *testing.T
	*redis.Client
}

type redisClientStub struct {
	*redis.Client
}

func TestNewRedisClient(t *testing.T) {
	client := NewRedisClient().(*redis.Client)
	expected := "Redis<chainr-redis:6379 db:0>"
	if client.String() != expected {
		t.Errorf("client = %v, expected %v", client, expected)
	}
}

func TestNewRedisClientSingleNode(t *testing.T) {
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
	client := NewRedisClient().(*redis.Client)

	expected := "Redis<test:1234 db:1>"
	if client.String() != expected {
		t.Errorf("client = %v, expected %v", client, expected)
	}
}

func TestNewRedisClientFailover(t *testing.T) {
	if err := os.Setenv("REDIS_ADDRS", "test:1234 test2:1234"); err != nil {
		t.Fatal(err)
	}
	defer os.Unsetenv("REDIS_ADDRS")
	if err := os.Setenv("REDIS_MASTER", "test"); err != nil {
		t.Fatal(err)
	}
	defer os.Unsetenv("REDIS_MASTER")
	client := NewRedisClient().(*redis.Client)

	expected := "Redis<FailoverClient db:0>"
	if client.String() != expected {
		t.Errorf("client = %v, expected %v", client, expected)
	}
}

// In case of error reading the redis database, the default database is used.
func TestNewRedisClientWithEnvError(t *testing.T) {
	if err := os.Setenv("REDIS_DB", "test"); err != nil {
		t.Fatal(err)
	}
	defer os.Unsetenv("REDIS_DB")
	client := NewRedisClient().(*redis.Client)

	expected := "Redis<chainr-redis:6379 db:0>"
	if client.String() != expected {
		t.Errorf("client = %v, expected %v", client, expected)
	}
}
