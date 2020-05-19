package run

import (
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v7"
)

// Scheduler allows to schedule a run, which can then be processed by a worker.
// It also allows to get the runs status.
type Scheduler interface {
	Schedule(run Run) (Status, error)
	Status(runUID string) (Status, error)
	StatusList() ([]StatusListItem, error)
}

type Status struct {
	Run  string
	Jobs []RunJob
}

type StatusListItem struct {
	RunUID string
	Status Status
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
	addrs := []string{"chainr-redis:6379"}
	masterName := ""
	password := ""
	db := 0
	if val, ok := os.LookupEnv("REDIS_ADDR"); ok {
		addrs = []string{val}
	}
	if val, ok := os.LookupEnv("REDIS_ADDRS"); ok {
		addrs = strings.Split(val, " ")
	}
	if val, ok := os.LookupEnv("REDIS_MASTER"); ok {
		masterName = val
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

	client := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:      addrs,
		MasterName: masterName,

		Password: password,
		DB:       db,

		MaxRetries:      6,
		MaxRetryBackoff: 10 * time.Second,
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

func makeJobDependencyKey(runUID string, jobName string, depIndex int) string {
	return "dependency:" + strconv.Itoa(depIndex) + ":" + makeJobKey(runUID, jobName)
}

// The jobItem struct is used internally to make a sorted
// list from the jobs map.
// The list is roughly sorted according to the dependency tree.
type jobItem struct {
	Name string
	Job  Job
}

// The Schedule method schedules the run for workers.
// It adds the run key in redis, and keys for each job and their dependencies.
// It also adds the jobs to the job queue.
func (s RedisScheduler) Schedule(run Run) (Status, error) {
	jobs := sortJobs(run.p.Jobs)

	if err := s.scheduleJobs(run.Metadata.UID, jobs); err != nil {
		return Status{}, err
	}
	if err := s.scheduleRun(run.Metadata.UID); err != nil {
		return Status{}, err
	}

	status := Status{
		Run:  "PENDING",
		Jobs: make([]RunJob, 0, len(jobs)),
	}
	for _, job := range jobs {
		status.Jobs = append(status.Jobs, RunJob{
			job.Name,
			"PENDING",
		})
	}
	return status, nil
}

// Roughly sorts jobs according to the dependency tree.
// This function does not manage dangling dependencies or loops.
func sortJobs(jobs map[string]Job) []jobItem {
	sorted := make([]jobItem, 0, len(jobs))

	for name, job := range jobs {
		sorted = append(sorted, jobItem{name, job})
	}

	sort.Slice(sorted, func(i, j int) bool {
		for _, dep := range sorted[j].Job.DependsOn {
			if dep.Job == sorted[i].Name {
				return true
			}
		}
		return false
	})

	return sorted
}

func bla(s []jobItem) []string {
	b := make([]string, 0, len(s))

	for _, v := range s {
		b = append(b, v.Name)
	}

	return b
}

// Jobs are added to the jobs list only if all jobs were successfully created.
// This avoids partial scheduling due to technical errors.
func (s RedisScheduler) scheduleJobs(runUID string, jobs []jobItem) error {
	jobKeys := make([]interface{}, 0, len(jobs))

	for _, jobItem := range jobs {
		jobName := jobItem.Name
		job := jobItem.Job

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
		if err := s.client.RPush(makeRunJobsKey(runUID), jobKeys...).Err(); err != nil {
			return err
		}
	}

	return nil
}

func (s RedisScheduler) scheduleDependencies(runUID string, jobName string, job Job) error {
	depKeys := make([]interface{}, 0, len(job.DependsOn))

	for i, dep := range job.DependsOn {
		failure := "false"
		if dep.Conditions.Failure {
			failure = "true"
		}
		depKey := makeJobDependencyKey(runUID, jobName, i)
		fields := []interface{}{
			"job", makeJobKey(runUID, dep.Job),
			"failure", failure,
		}
		if err := s.client.HSet(depKey, fields...).Err(); err != nil {
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
	if err := s.client.LPush(runsKey, runKey).Err(); err != nil {
		return err
	}

	return nil
}

func (s RedisScheduler) Status(runUID string) (Status, error) {
	status := Status{
		Run:  "PENDING",
		Jobs: make([]RunJob, 0),
	}

	run, err := s.client.HGetAll(makeRunKey(runUID)).Result()
	if err != nil {
		return status, err
	}
	if len(run) == 0 {
		return status, &NotFoundError{runUID}
	}
	status.Run = run["status"]

	jobKeys, err := s.client.LRange(makeRunJobsKey(runUID), 0, -1).Result()
	if err != nil {
		return status, err
	}

	for _, jobKey := range jobKeys {
		job, err := s.client.HGetAll(jobKey).Result()
		if err != nil {
			return status, err
		}

		status.Jobs = append(status.Jobs, RunJob{
			job["name"],
			job["status"],
		})
	}

	return status, nil
}

func (s RedisScheduler) StatusList() ([]StatusListItem, error) {
	statusList := make([]StatusListItem, 0)

	runKeys, err := s.client.LRange(makeRunsKey(), 0, -1).Result()
	if err != nil {
		return statusList, err
	}
	for _, runKey := range runKeys {
		runUID, err := s.client.HGet(runKey, "uid").Result()
		if err != nil {
			return statusList, err
		}

		status, err := s.Status(runUID)
		if err != nil {
			return statusList, err
		}

		statusList = append(statusList, StatusListItem{runUID, status})
	}

	return statusList, nil
}
