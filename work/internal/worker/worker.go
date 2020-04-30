// Package worker contains tools to interact with
// the jobs store, and process jobs.
// A worker consumes jobs, starts them on Kubernetes,
// monitors them and updates their status.
package worker

import (
	"log"
	"os"
	"strconv"

	"github.com/go-redis/redis/v7"
)

type Worker interface {
	Start() error
}

type RedisWorker struct {
	client redis.Cmdable
}

func New() Worker {
	addr := "redis:6379"
	password := ""
	db := 0
	if val, ok := os.LookupEnv("REDIS_ADDR"); ok {
		addr = val
	}
	if val, ok := os.LookupEnv("REDIS_PASSWORD"); ok {
		password = val
	}
	if val, ok := os.LookupEnv("REDIS_DB"); ok {
		d, err := strconv.Atoi(val)
		if err != nil {
			log.Println("Invalid REDIS_DB value " + val + ", using default 0")
			d = 0
		}
		db = d
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisWorker{client}
}

func (w RedisWorker) Start() error {
	workQueue := "runs:work"

	for {
		vals, err := w.client.BRPop(0, workQueue).Result()
		if err != nil {
			return err
		}

		go w.process(vals[1])
	}

	return nil
}

func (w RedisWorker) process(runKey string) {
	defer w.recover(runKey)

	if err := w.startRun(runKey); err != nil {
		log.Println("Run", runKey, "failed:", err.Error())
	}
}

func (w RedisWorker) startRun(runKey string) error {
	log.Println("Starting run", runKey)

	if err := w.client.HSet(runKey, "status", "RUNNING").Err(); err != nil {
		return err
	}

	runUID, err := w.client.HGet(runKey, "uid").Result()
	if err != nil {
		return err
	}

	runJobsKey := "jobs:run:" + runUID
	jobKeys, err := w.client.SMembers(runJobsKey).Result()
	if err != nil {
		return err
	}

	for _, jobKey := range jobKeys {
		if err := w.startJob(jobKey); err != nil {
			return err
		}
	}

	status := "SUCCESSFUL"

	if err := w.client.HSet(runKey, "status", status).Err(); err != nil {
		return err
	}
	log.Println("Run", runKey, "completed with status", status)

	return nil
}

func (w RedisWorker) startJob(jobKey string) error {
	job, err := w.client.HGetAll(jobKey).Result()
	if err != nil {
		return err
	}

	log.Printf(`Starting job %v
	image: %v
	run: %v`, jobKey, job["image"], job["run"])

	if err := w.client.HSet(jobKey, "status", "RUNNING").Err(); err != nil {
		return err
	}

	// TODO: run and monitor

	status := "SUCCESSFUL"

	if err := w.client.HSet(jobKey, "status", status).Err(); err != nil {
		return err
	}
	log.Println("Job", jobKey, "completed with status", status)

	return nil
}

func (w RedisWorker) recover(runKey string) {
	if r := recover(); r != nil {
		log.Println("Run", runKey, "processing was interrupted by a panic:", r)
		if err := w.client.HSet(runKey, "status", "CANCELED").Err(); err != nil {
			log.Println("Unable to set run", runKey, "status to CANCELED")
		}
		panic(r)
	}
}
