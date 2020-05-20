package recycler

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"time"

	"github.com/go-redis/redis/v7"
)

type redisMock struct {
	t *testing.T
	*redis.Client

	rPushCalled bool
	sRemCalled  bool
}

func NewRedisMock(t *testing.T) *redisMock {
	return &redisMock{t: t}
}

func (r *redisMock) SMembers(key string) *redis.StringSliceCmd {
	if key != "workers" {
		r.t.Errorf("SMembers: key = %v, expected workers", key)
	}

	return redis.NewStringSliceResult([]string{
		"running",
		"expired",
	}, nil)
}

func (r *redisMock) HGetAll(key string) *redis.StringStringMapCmd {
	var res map[string]string
	switch key {
	case "running":
		res = map[string]string{
			"queue":        "running:items",
			"processQueue": "running:processing",
			"expiry":       time.Now().Add(1 * time.Second).Format(time.RFC3339),
		}
	case "expired":
		res = map[string]string{
			"queue":        "expired:items",
			"processQueue": "expired:processing",
			"expiry":       time.Now().Add(-1 * time.Second).Format(time.RFC3339),
		}
	default:
		r.t.Errorf("HGetAll: key = %v, expected running or expired", key)
	}

	return redis.NewStringStringMapResult(res, nil)
}

func (r *redisMock) LRange(key string, start, stop int64) *redis.StringSliceCmd {
	if key != "expired:processing" {
		r.t.Errorf("LRange: key = %v, expected expired:processing", key)
	}
	if start != 0 || stop != -1 {
		r.t.Errorf("LRange: expected full range of the list")
	}

	return redis.NewStringSliceResult([]string{"expired:item1", "expired:item2"}, nil)
}

func (r *redisMock) RPush(key string, values ...interface{}) *redis.IntCmd {
	if key != "expired:items" {
		r.t.Errorf("RPush: key = %v, expected expired:items", key)
	}

	expectedValues := []string{"expired:item1", "expired:item2"}
	if len(values) != len(expectedValues) || values[0] != expectedValues[0] || values[1] != expectedValues[1] {
		r.t.Errorf("RPush: values = %v, expected %v", values, expectedValues)
	}

	r.rPushCalled = true
	return redis.NewIntResult(0, nil)
}

func (r *redisMock) Del(keys ...string) *redis.IntCmd {
	for _, key := range keys {
		switch key {
		case "expired:processing":
		case "expired":
		default:
			r.t.Errorf("Del: key = %v, expected expired:processing or expired", key)
		}
	}
	return redis.NewIntResult(1, nil)
}

func (r *redisMock) SRem(key string, members ...interface{}) *redis.IntCmd {
	if key != "workers" {
		r.t.Errorf("SRem: key = %v, expected workers", key)
	}

	expected := []string{"expired"}
	if len(members) != len(expected) || members[0] != expected[0] {
		r.t.Errorf("SRem: members = %v, expected %v", members, expected)
	}

	r.sRemCalled = true
	return redis.NewIntResult(1, nil)
}

func TestRecycle(t *testing.T) {
	Convey("Scenario: recycle workers", t, func() {
		Convey("Given the recycle procedure is started", func() {
			r := NewRedisMock(t)
			recycler{r}.Recycle()

			Convey("When there are expired workers", func() {
				Convey("Expired workers items should be rescheduled", func() {
					So(r.rPushCalled, ShouldBeTrue)
				})

				Convey("Expired workers should be deleted", func() {
					So(r.sRemCalled, ShouldBeTrue)
				})
			})
		})
	})
}
