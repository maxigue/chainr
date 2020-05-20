// Package recycler contains the recycling logic.
// The recycler reads the workers status on redis and re-schedules
// items from expired workers.
package recycler

import (
	"log"
	"strings"
	"time"

	"github.com/go-redis/redis/v7"
)

type recycler struct {
	client redis.Cmdable
}

func (r recycler) Recycle() error {
	workerKeys, err := r.client.SMembers("workers").Result()
	if err != nil {
		return err
	}

	for _, workerKey := range workerKeys {
		if err := r.recycleWorker(workerKey); err != nil {
			return err
		}
	}

	return nil
}

func (r recycler) recycleWorker(workerKey string) error {
	worker, err := r.client.HGetAll(workerKey).Result()
	if err != nil {
		return err
	}

	expiry, err := time.Parse(time.RFC3339, worker["expiry"])
	if err != nil {
		return err
	}

	if expiry.Before(time.Now()) {
		log.Println("Worker", workerKey, "has expired")
		if err := r.rescheduleItems(worker["processQueue"], worker["queue"]); err != nil {
			return err
		}
		if err := r.deleteWorker(workerKey); err != nil {
			return err
		}
	}

	return nil
}

func (r recycler) rescheduleItems(processQueue string, queue string) error {
	items, err := r.client.LRange(processQueue, 0, -1).Result()
	if err != nil {
		return err
	}

	if len(items) > 0 {
		log.Println("Rescheduling items", strings.Join(items, ", "), "on queue", queue)
		values := make([]interface{}, len(items))
		for i, v := range items {
			values[i] = v
		}
		if err := r.client.RPush(queue, values...).Err(); err != nil {
			return err
		}
	}

	log.Println("Deleting process queue", processQueue)
	if err := r.client.Del(processQueue).Err(); err != nil {
		return err
	}

	return nil
}

func (r recycler) deleteWorker(workerKey string) error {
	log.Println("Deleting worker", workerKey)
	if err := r.client.SRem("workers", workerKey).Err(); err != nil {
		return err
	}
	if err := r.client.Del(workerKey).Err(); err != nil {
		return err
	}

	return nil
}

func Start() {
	r := recycler{NewRedisClient()}

	for {
		if err := r.Recycle(); err != nil {
			log.Println("An error occurred while recycling:", err.Error())
		}
		time.Sleep(10 * time.Second)
	}
}
