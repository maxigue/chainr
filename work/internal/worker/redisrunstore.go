package worker

import "github.com/go-redis/redis/v7"

type RedisRunStore struct {
	info   Info
	client redis.Cmdable
}

func NewRedisRunStore(info Info) RedisRunStore {
	return RedisRunStore{info, NewRedisClient()}
}

func (rs RedisRunStore) NextRun() (string, error) {
	val, err := rs.client.BRPopLPush(rs.info.Queue, rs.info.ProcessQueue, 0).Result()
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
	return rs.client.LRem(rs.info.ProcessQueue, -1, runKey).Err()
}
