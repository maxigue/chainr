package worker

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"errors"
	"time"

	"github.com/go-redis/redis/v7"
)

// Test that the recycler creation does not panic.
func TestNewRecycler(t *testing.T) {
	_ = NewRecycler(testInfo)
}

type redisMock redisClientMock

func (r redisMock) HSet(key string, values ...interface{}) *redis.IntCmd {
	if key != "worker:xyz" {
		r.t.Errorf("HSet: key = %v, expected worker:xyz", key)
	}

	expectedValues := []string{"queue", "events:notif", "processQueue", "events:notifier:xyz", "expiry"}
	for i := range expectedValues {
		if values[i] != expectedValues[i] {
			r.t.Errorf("HSet: values[%v] = %v, expected %v", i, values[i], expectedValues[i])
		}
	}
	exp, err := time.Parse(time.RFC3339, values[len(expectedValues)].(string))
	if err != nil {
		r.t.Fatal(err)
	}
	now := time.Now()
	if !exp.After(now) {
		r.t.Errorf("HSet: expiry %v is not after now (%v)", exp, now)
	}

	return redis.NewIntResult(1, nil)
}

func (r redisMock) SAdd(key string, members ...interface{}) *redis.IntCmd {
	if key != "workers" {
		r.t.Errorf("SAdd: key = %v, expected workers", key)
	}
	if len(members) != 1 {
		r.t.Errorf("SAdd: adding %v members, expected 1", len(members))
	}
	if members[0].(string) != "worker:xyz" {
		r.t.Errorf("SAdd: member is %v, expected worker:xyz", members[0])
	}

	return redis.NewIntResult(1, nil)
}

type redisHSetErrorStub redisClientStub

func (r redisHSetErrorStub) HSet(key string, value ...interface{}) *redis.IntCmd {
	return redis.NewIntResult(0, errors.New("HSet error"))
}

func (r redisHSetErrorStub) SAdd(key string, members ...interface{}) *redis.IntCmd {
	return redis.NewIntResult(0, nil)
}

type redisSAddErrorStub redisClientStub

func (r redisSAddErrorStub) HSet(key string, value ...interface{}) *redis.IntCmd {
	return redis.NewIntResult(0, nil)
}

func (r redisSAddErrorStub) SAdd(key string, members ...interface{}) *redis.IntCmd {
	return redis.NewIntResult(0, errors.New("SAdd error"))
}

func TestSync(t *testing.T) {
	Convey("Scenario: synchronize with recycler", t, func() {
		Convey("Given the recycler is synchronizing", func() {
			Convey("When everything goes well", func() {
				RedisRecycler{testInfo, &redisMock{t: t}}.sync()

				Convey("The worker information should be synchronized", func() {
					// Tested in redisMock.HSet
				})

				Convey("The worker should be added to the workers set", func() {
					// Tested in redisMock.SAdd
				})
			})

			Convey("When an error occurs during worker information synchronization", func() {
				r := RedisRecycler{testInfo, &redisHSetErrorStub{}}

				Convey("The synchronizer should not panic", func() {
					r.sync()
				})
			})

			Convey("When an error occurs while adding worker to workers set", func() {
				r := RedisRecycler{testInfo, &redisSAddErrorStub{}}

				Convey("The synchronizer should not panic", func() {
					r.sync()
				})
			})
		})
	})
}
