package worker

import (
	"log"
	"time"

	"github.com/go-redis/redis/v7"
)

type RedisRecycler struct {
	info   Info
	client redis.Cmdable
}

func NewRecycler(info Info) Recycler {
	return &RedisRecycler{info, NewRedisClient()}
}

// This function registers the worker with the recycler, and
// frequently sends keepalives.
// It should be called in a goroutine.
// In case of error, it keeps retrying.
func (r RedisRecycler) StartSync() {
	log.Println("Starting recycler synchronization loop")

	for {
		r.sync()
		time.Sleep(4 * time.Second)
	}
}

// This function sets the worker information and registers the worker
// for the recycler.
func (r RedisRecycler) sync() {
	workerKey := "worker:" + r.info.Name
	workersKey := "workers"

	fields := []interface{}{
		"queue", r.info.Queue,
		"processQueue", r.info.ProcessQueue,
		"expiry", time.Now().Add(15 * time.Second).Format(time.RFC3339),
	}
	if err := r.client.HSet(workerKey, fields...).Err(); err != nil {
		log.Println("Unable to set worker information for recycler:", err)
	} else if err := r.client.SAdd(workersKey, workerKey).Err(); err != nil {
		log.Println("Unable to register worker for recycler:", err)
	}
}
