package worker

import (
	"log"
	"os"
	"strconv"

	"github.com/go-redis/redis/v7"
)

type RedisRunStore struct {
	client redis.Cmdable
}

func NewRedisRunStore() RedisRunStore {
	addr := "chainr-redis:6379"
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

	return RedisRunStore{client}
}

func (rs RedisRunStore) NextRun() (string, error) {
	workQueue := "runs:work"

	vals, err := rs.client.BRPop(0, workQueue).Result()
	if err != nil {
		return "", err
	}

	return vals[1], nil
}

func (rs RedisRunStore) SetRunStatus(runKey, status string) error {
	return rs.client.HSet(runKey, "status", status).Err()
}

func (rs RedisRunStore) GetJobs(runKey string) ([]string, error) {
	runJobsKey := "jobs:" + runKey
	return rs.client.SMembers(runJobsKey).Result()
}

func (rs RedisRunStore) GetJob(jobKey string) (Job, error) {
	job, err := rs.client.HGetAll(jobKey).Result()
	if err != nil {
		return Job{}, err
	}

	return Job{
		job["name"],
		job["image"],
		job["run"],
	}, nil
}

func (rs RedisRunStore) SetJobStatus(jobKey, status string) error {
	return rs.client.HSet(jobKey, "status", status).Err()
}

func (rs RedisRunStore) GetJobDependencies(jobKey string) ([]JobDependency, error) {
	deps := make([]JobDependency, 0)

	jobDependenciesKey := "dependencies:" + jobKey
	depKeys, err := rs.client.SMembers(jobDependenciesKey).Result()
	if err != nil {
		return deps, err
	}

	for _, depKey := range depKeys {
		dep, err := rs.client.HGetAll(depKey).Result()
		if err != nil {
			return deps, err
		}

		deps = append(deps, JobDependency{
			dep["job"],
			(dep["failure"] == "true"),
		})
	}

	return deps, nil
}
