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

func makeRunKey(runUID string) string {
	return "run:" + runUID
}

func makeJobKey(runUID string, jobName string) string {
	return "job:" + jobName + ":run:" + runUID
}

func makeJobDependencyKey(runUID string, jobName string, depName string) string {
	return "dependency:" + depName + ":job:" + jobName + ":run:" + runUID
}

// The Schedule method schedules the run for workers.
// It adds the run key in redis, and keys for each job and their dependencies.
// It also adds the jobs to the job queue.
func (s RedisScheduler) Schedule(run Run) (Status, error) {
	if err := s.scheduleRun(run.Metadata.UID); err != nil {
		return Status{}, err
	}
	if err := s.scheduleJobs(run.Metadata.UID, run.p.Jobs); err != nil {
		return Status{}, err
	}

	status := make(Status, len(run.p.Jobs))
	for k := range run.p.Jobs {
		status[k] = "PENDING"
	}
	return status, nil
}

func (s RedisScheduler) scheduleRun(runUID string) error {
	runKey := makeRunKey(runUID)
	return s.client.Set(runKey, runUID, 0).Err()
}

func (s RedisScheduler) scheduleJobs(runUID string, jobs map[string]Job) error {
	pendingJobsChannel := "work:jobs"

	for jobName, job := range jobs {
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
		if err := s.scheduleDependencies(runUID, jobName, job); err != nil {
			return err
		}
		if err := s.client.LPush(pendingJobsChannel, jobKey).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (s RedisScheduler) scheduleDependencies(runUID string, jobName string, job Job) error {
	for _, dep := range job.DependsOn {
		depKey := makeJobDependencyKey(runUID, jobName, dep.Job)
		failure := "false"
		if dep.Conditions.Failure {
			failure = "true"
		}
		if err := s.client.HSet(depKey, "failure", failure).Err(); err != nil {
			return err
		}
	}

	return nil
}

func (s RedisScheduler) Status(runUID string) (Status, error) {
	status := Status{}

	vals, err := s.client.Keys(makeJobKey(runUID, "*")).Result()
	if err != nil {
		return nil, err
	}

	if len(vals) == 0 {
		return status, &NotFoundError{runUID}
	}

	for _, key := range vals {
		job, err := s.client.HGetAll(key).Result()
		if err != nil {
			return status, err
		}

		status[job["name"]] = job["status"]
	}

	return status, nil
}

func (s RedisScheduler) StatusMap() (map[string]Status, error) {
	statusMap := make(map[string]Status)

	vals, err := s.client.Keys(makeRunKey("*")).Result()
	if err != nil {
		return statusMap, err
	}
	for _, key := range vals {
		runUID, err := s.client.Get(key).Result()
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
