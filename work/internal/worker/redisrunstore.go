package worker

import (
	"os"

	"github.com/go-redis/redis/v7"
)

type RedisRunStore struct {
	client       redis.Cmdable
	processQueue string
}

func NewRedisRunStore() RedisRunStore {
	workerName, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	processQueue := "runs:worker:" + workerName
	return RedisRunStore{NewRedisClient(), processQueue}
}

func (rs RedisRunStore) NextRun() (string, error) {
	workQueue := "runs:work"

	val, err := rs.client.BRPopLPush(workQueue, rs.processQueue, 0).Result()
	if err != nil {
		return "", err
	}

	return val, nil
}

func (rs RedisRunStore) SetRunStatus(runKey, status string) error {
	return rs.client.HSet(runKey, "status", status).Err()
}

func (rs RedisRunStore) GetJobs(runKey string) ([]string, error) {
	runJobsKey := "jobs:" + runKey
	return rs.client.LRange(runJobsKey, 0, -1).Result()
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

func (rs RedisRunStore) Close(runKey string) error {
	return rs.client.LRem(rs.processQueue, -1, runKey).Err()
}
