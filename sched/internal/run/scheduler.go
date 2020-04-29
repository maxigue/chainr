package run

import (
	"log"
	"os"
	"strconv"

	"github.com/go-redis/redis/v7"
)

// Scheduler allows to schedule a run, which can then be processed by a worker.
// It also allows to get the runs status.
type Scheduler interface {
	Schedule(run Run) (Status, error)
	Status(runUID string) (Status, error)
	StatusMap() (map[string]Status, error)
}

type Status struct {
	Run  string
	Jobs map[string]string
}

type NotFoundError struct {
	RunUID string
}

func (e NotFoundError) Error() string {
	return "run " + e.RunUID + " was not found"
}

type RedisScheduler struct {
	client redis.Cmdable
}

func NewScheduler() Scheduler {
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

	return &RedisScheduler{client}
}

func makeRunsKey() string {
	return "runs"
}

func makeRunKey(runUID string) string {
	return "run:" + runUID
}

func makeRunJobsKey(runUID string) string {
	return "jobs:" + makeRunKey(runUID)
}

func makeWorkRunsKey() string {
	return "runs:work"
}

func makeJobKey(runUID string, jobName string) string {
	return "job:" + jobName + ":" + makeRunKey(runUID)
}

func makeJobDependenciesKey(runUID string, jobName string) string {
	return "dependencies:" + makeJobKey(runUID, jobName)
}

func makeJobDependencyKey(runUID string, jobName string, depName string) string {
	return "dependency:" + depName + ":" + makeJobKey(runUID, jobName)
}

// The Schedule method schedules the run for workers.
// It adds the run key in redis, and keys for each job and their dependencies.
// It also adds the jobs to the job queue.
func (s RedisScheduler) Schedule(run Run) (Status, error) {
	if err := s.scheduleJobs(run.Metadata.UID, run.p.Jobs); err != nil {
		return Status{}, err
	}
	if err := s.scheduleRun(run.Metadata.UID); err != nil {
		return Status{}, err
	}

	status := Status{
		Run:  "PENDING",
		Jobs: make(map[string]string),
	}
	for k := range run.p.Jobs {
		status.Jobs[k] = "PENDING"
	}
	return status, nil
}

// Jobs are added to the jobs set and the work queue only
// if all jobs were successfully created.
// This avoids partial scheduling due to technical errors.
func (s RedisScheduler) scheduleJobs(runUID string, jobs map[string]Job) error {
	jobKeys := make([]interface{}, 0, len(jobs))

	for jobName, job := range jobs {
		if err := s.scheduleDependencies(runUID, jobName, job); err != nil {
			return err
		}

		jobKey := makeJobKey(runUID, jobName)
		fields := []interface{}{
			"name", jobName,
			"image", job.Image,
			"run", job.Run,
			"status", "PENDING",
		}
		if err := s.client.HSet(jobKey, fields...).Err(); err != nil {
			return err
		}
		jobKeys = append(jobKeys, jobKey)
	}

	if len(jobKeys) > 0 {
		if err := s.client.SAdd(makeRunJobsKey(runUID), jobKeys...).Err(); err != nil {
			return err
		}
	}

	return nil
}

func (s RedisScheduler) scheduleDependencies(runUID string, jobName string, job Job) error {
	depKeys := make([]interface{}, 0, len(job.DependsOn))

	for _, dep := range job.DependsOn {
		failure := "false"
		if dep.Conditions.Failure {
			failure = "true"
		}
		depKey := makeJobDependencyKey(runUID, jobName, dep.Job)
		if err := s.client.HSet(depKey, "failure", failure).Err(); err != nil {
			return err
		}
		depKeys = append(depKeys, depKey)
	}

	if len(depKeys) > 0 {
		depsKey := makeJobDependenciesKey(runUID, jobName)
		if err := s.client.SAdd(depsKey, depKeys...).Err(); err != nil {
			return err
		}
	}

	return nil
}

func (s RedisScheduler) scheduleRun(runUID string) error {
	runKey := makeRunKey(runUID)
	fields := []interface{}{
		"uid", runUID,
		"status", "PENDING",
	}
	if err := s.client.HSet(runKey, fields...).Err(); err != nil {
		return err
	}
	if err := s.client.LPush(makeWorkRunsKey(), runKey).Err(); err != nil {
		return err
	}
	runsKey := makeRunsKey()
	if err := s.client.SAdd(runsKey, runKey).Err(); err != nil {
		return err
	}

	return nil
}

func (s RedisScheduler) Status(runUID string) (Status, error) {
	status := Status{
		Run:  "PENDING",
		Jobs: make(map[string]string),
	}

	run, err := s.client.HGetAll(makeRunKey(runUID)).Result()
	if err != nil {
		return status, err
	}
	if len(run) == 0 {
		return status, &NotFoundError{runUID}
	}
	status.Run = run["status"]

	jobKeys, err := s.client.SMembers(makeRunJobsKey(runUID)).Result()
	if err != nil {
		return status, err
	}

	for _, jobKey := range jobKeys {
		job, err := s.client.HGetAll(jobKey).Result()
		if err != nil {
			return status, err
		}

		status.Jobs[job["name"]] = job["status"]
	}

	return status, nil
}

func (s RedisScheduler) StatusMap() (map[string]Status, error) {
	statusMap := make(map[string]Status)

	runKeys, err := s.client.SMembers(makeRunsKey()).Result()
	if err != nil {
		return statusMap, err
	}
	for _, runKey := range runKeys {
		runUID, err := s.client.HGet(runKey, "uid").Result()
		if err != nil {
			return statusMap, err
		}

		status, err := s.Status(runUID)
		if err != nil {
			return statusMap, err
		}

		statusMap[runUID] = status
	}

	return statusMap, nil
}
