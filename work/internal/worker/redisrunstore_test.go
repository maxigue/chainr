package worker

import (
	"testing"

	"errors"
	"time"

	"github.com/go-redis/redis/v7"
)

func TestNewRedisRunStore(t *testing.T) {
	// Test that NewRedisRunStore does not panic.
	_ = NewRedisRunStore(testInfo)
}

type nextRunClientMock redisClientMock

func (c nextRunClientMock) BRPopLPush(source, destination string, timeout time.Duration) *redis.StringCmd {
	if timeout != 0 {
		c.t.Errorf("BRPopLPush should block indefinitely")
	}
	if source != "runs:work" {
		c.t.Errorf("BRPopLPush listens on %v, expected runs:work", source)
	}
	if destination != "runs:worker:xyz" {
		c.t.Errorf("BRPopLPush listens on %v, expected runs:worker:xyz", destination)
	}

	return redis.NewStringResult("run:abc", nil)
}

func TestNextRun(t *testing.T) {
	rs := RedisRunStore{testInfo, &nextRunClientMock{t: t}}
	runID, err := rs.NextRun()
	if err != nil {
		t.Fatal(err)
	}
	if runID != "run:abc" {
		t.Errorf("runID = %v, expected run:abc", runID)
	}
}

type nextRunClientErrorMock redisClientMock

func (c nextRunClientErrorMock) BRPopLPush(source, destination string, timeout time.Duration) *redis.StringCmd {
	return redis.NewStringResult("", errors.New("BRPopLPush failed"))
}

func TestNextRunError(t *testing.T) {
	rs := RedisRunStore{testInfo, &nextRunClientErrorMock{t: t}}
	_, err := rs.NextRun()
	if err.Error() != "BRPopLPush failed" {
		t.Errorf("redis error was not forwarded")
	}
}

type setRunStatusClientMock redisClientMock

func (c setRunStatusClientMock) HSet(key string, values ...interface{}) *redis.IntCmd {
	if key != "run:abc" {
		c.t.Errorf("key = %v, expected run:abc", key)
	}
	if values[0] != "status" {
		c.t.Errorf("values[0] = %v, expected status", values[0])
	}
	if values[1] != "RUNNING" {
		c.t.Errorf("values[1] = %v, expected RUNNING", values[1])
	}

	return redis.NewIntResult(0, nil)
}

func TestSetRunStatus(t *testing.T) {
	rs := RedisRunStore{testInfo, &setRunStatusClientMock{t: t}}
	err := rs.SetRunStatus("run:abc", "RUNNING")
	if err != nil {
		t.Fatal(err)
	}
}

type setRunStatusClientErrorMock redisClientMock

func (c setRunStatusClientErrorMock) HSet(key string, values ...interface{}) *redis.IntCmd {
	return redis.NewIntResult(0, errors.New("HSet failed"))
}

func TestSetRunStatusError(t *testing.T) {
	rs := RedisRunStore{testInfo, &setRunStatusClientErrorMock{t: t}}
	err := rs.SetRunStatus("run:abc", "RUNNING")
	if err.Error() != "HSet failed" {
		t.Errorf("redis error was not forwarded")
	}
}

type getJobsClientMock redisClientMock

func (c getJobsClientMock) LRange(key string, start, stop int64) *redis.StringSliceCmd {
	if start != 0 || stop != -1 {
		c.t.Errorf("start=%v, stop=%v, expected start=0, stop=-1", start, stop)
	}

	if key != "jobs:run:abc" {
		c.t.Errorf("key = %v, expected jobs:run:abc", key)
	}

	return redis.NewStringSliceResult([]string{"job:job1:run:abc", "job:job2:run:abc"}, nil)
}

func TestGetJobs(t *testing.T) {
	rs := RedisRunStore{testInfo, &getJobsClientMock{t: t}}
	jobIDs, err := rs.GetJobs("run:abc")
	if err != nil {
		t.Fatal(err)
	}
	if jobIDs[0] != "job:job1:run:abc" {
		t.Errorf("jobIDs[0] = %v, expected job:job1:run:abc", jobIDs[0])
	}
	if jobIDs[1] != "job:job2:run:abc" {
		t.Errorf("jobIDs[1] = %v, expected job:job2:run:abc", jobIDs[1])
	}
}

type getJobsClientErrorMock redisClientMock

func (c getJobsClientErrorMock) LRange(key string, start, stop int64) *redis.StringSliceCmd {
	return redis.NewStringSliceResult([]string{}, errors.New("SMembers failed"))
}

func TestGetJobsError(t *testing.T) {
	rs := RedisRunStore{testInfo, &getJobsClientErrorMock{t: t}}
	_, err := rs.GetJobs("run:abc")
	if err.Error() != "SMembers failed" {
		t.Errorf("redis error was not forwarded")
	}
}

type getJobClientMock redisClientMock

func (c getJobClientMock) HGetAll(key string) *redis.StringStringMapCmd {
	if key != "job:job1:run:abc" {
		c.t.Errorf("key = %v, expected job:job1:run:abc", key)
	}

	vals := map[string]string{
		"name":   "job1",
		"image":  "busybox",
		"run":    "exit 0",
		"status": "RUNNING",
	}
	return redis.NewStringStringMapResult(vals, nil)
}

func TestGetJob(t *testing.T) {
	rs := RedisRunStore{testInfo, &getJobClientMock{t: t}}
	job, err := rs.GetJob("job:job1:run:abc")
	if err != nil {
		t.Fatal(err)
	}
	if job.Name != "job1" {
		t.Errorf("job.Name = %v, expected job1", job.Name)
	}
	if job.Image != "busybox" {
		t.Errorf("job.Image = %v, expected busybox", job.Image)
	}
	if job.Run != "exit 0" {
		t.Errorf("job.Run = %v, expected exit 0", job.Run)
	}
}

type getJobClientErrorMock redisClientMock

func (c getJobClientErrorMock) HGetAll(key string) *redis.StringStringMapCmd {
	return redis.NewStringStringMapResult(make(map[string]string), errors.New("HGetAll failed"))
}

func TestGetJobError(t *testing.T) {
	rs := RedisRunStore{testInfo, &getJobClientErrorMock{t: t}}
	_, err := rs.GetJob("job:job1:run:abc")
	if err.Error() != "HGetAll failed" {
		t.Errorf("redis error was not forwarded")
	}
}

type setJobStatusClientMock redisClientMock

func (c setJobStatusClientMock) HSet(key string, values ...interface{}) *redis.IntCmd {
	if key != "job:job1:run:abc" {
		c.t.Errorf("key = %v, expected job:job1:run:abc", key)
	}
	if values[0] != "status" {
		c.t.Errorf("values[0] = %v, expected status", values[0])
	}
	if values[1] != "RUNNING" {
		c.t.Errorf("values[1] = %v, expected RUNNING", values[1])
	}

	return redis.NewIntResult(0, nil)
}

func TestSetJobStatus(t *testing.T) {
	rs := RedisRunStore{testInfo, &setJobStatusClientMock{t: t}}
	err := rs.SetJobStatus("job:job1:run:abc", "RUNNING")
	if err != nil {
		t.Fatal(err)
	}
}

type setJobStatusClientErrorMock redisClientMock

func (c setJobStatusClientErrorMock) HSet(key string, values ...interface{}) *redis.IntCmd {
	return redis.NewIntResult(0, errors.New("HSet failed"))
}

func TestSetJobStatusError(t *testing.T) {
	rs := RedisRunStore{testInfo, &setJobStatusClientErrorMock{t: t}}
	err := rs.SetJobStatus("job:job1:run:abc", "RUNNING")
	if err.Error() != "HSet failed" {
		t.Errorf("redis error was not forwarded")
	}
}

type getJobDependenciesClientMock redisClientMock

func (c getJobDependenciesClientMock) SMembers(key string) *redis.StringSliceCmd {
	if key != "dependencies:job:job1:run:abc" {
		c.t.Errorf("key = %v, expected dependencies:job:job1:run:abc", key)
	}

	return redis.NewStringSliceResult([]string{
		"dependency:0:job:job1:run:abc",
		"dependency:1:job:job1:run:abc",
	}, nil)
}
func (c getJobDependenciesClientMock) HGetAll(key string) *redis.StringStringMapCmd {
	vals := make(map[string]string)

	switch key {
	case "dependency:0:job:job1:run:abc":
		vals = map[string]string{
			"job":     "job:dep1:run:abc",
			"failure": "false",
		}
	case "dependency:1:job:job1:run:abc":
		vals = map[string]string{
			"job":     "job:dep2:run:abc",
			"failure": "true",
		}
	default:
		c.t.Errorf("key = %v, expected dependency:0:job:job1:run:abc or dependency:1:job:job1:run:abc", key)
	}

	return redis.NewStringStringMapResult(vals, nil)
}

func TestGetJobDependencies(t *testing.T) {
	rs := RedisRunStore{testInfo, &getJobDependenciesClientMock{t: t}}
	deps, err := rs.GetJobDependencies("job:job1:run:abc")
	if err != nil {
		t.Fatal(err)
	}
	if deps[0].JobID != "job:dep1:run:abc" {
		t.Errorf("deps[0].JobID = %v, expected job:dep1:run:abc", deps[0].JobID)
	}
	if deps[0].ExpectFailure != false {
		t.Errorf("deps[0].ExpectFailure = %v, expected false", deps[0].ExpectFailure)
	}
	if deps[1].JobID != "job:dep2:run:abc" {
		t.Errorf("deps[1].JobID = %v, expected job:dep2:run:abc", deps[1].JobID)
	}
	if deps[1].ExpectFailure != true {
		t.Errorf("deps[1].ExpectFailure = %v, expected true", deps[1].ExpectFailure)
	}
}

type getJobDependenciesClientErrorSMembersMock redisClientMock

func (c getJobDependenciesClientErrorSMembersMock) SMembers(key string) *redis.StringSliceCmd {
	return redis.NewStringSliceResult([]string{}, errors.New("SMembers failed"))
}
func (c getJobDependenciesClientErrorSMembersMock) HGetAll(key string) *redis.StringStringMapCmd {
	return redis.NewStringStringMapResult(make(map[string]string), errors.New("HGetAll failed"))
}

func TestGetJobDependenciesErrorSMembers(t *testing.T) {
	rs := RedisRunStore{testInfo, &getJobDependenciesClientErrorSMembersMock{t: t}}
	_, err := rs.GetJobDependencies("job:job1:run:abc")
	if err.Error() != "SMembers failed" {
		t.Errorf("redis error was not forwarded")
	}
}

type getJobDependenciesClientErrorHGetAllMock redisClientMock

func (c getJobDependenciesClientErrorHGetAllMock) SMembers(key string) *redis.StringSliceCmd {
	return redis.NewStringSliceResult([]string{"dependency:0:job:job1:run:abc"}, nil)
}
func (c getJobDependenciesClientErrorHGetAllMock) HGetAll(key string) *redis.StringStringMapCmd {
	return redis.NewStringStringMapResult(make(map[string]string), errors.New("HGetAll failed"))
}

func TestGetJobDependenciesErrorHGetAll(t *testing.T) {
	rs := RedisRunStore{testInfo, &getJobDependenciesClientErrorHGetAllMock{t: t}}
	_, err := rs.GetJobDependencies("job:job1:run:abc")
	if err.Error() != "HGetAll failed" {
		t.Errorf("redis error was not forwarded")
	}
}

type closeClientMock redisClientMock

func (c closeClientMock) LRem(key string, count int64, value interface{}) *redis.IntCmd {
	if key != "runs:worker:xyz" {
		c.t.Errorf("key = %v, expected runs:worker:xyz", key)
	}
	if count != -1 {
		c.t.Errorf("count = %v, expected -1 to remove only oldest element", count)
	}
	if value != "run:abc" {
		c.t.Errorf("value = %v, expected run:abc", value)
	}

	return redis.NewIntResult(1, nil)
}

func TestClose(t *testing.T) {
	rs := RedisRunStore{testInfo, &closeClientMock{t: t}}
	err := rs.Close("run:abc")
	if err != nil {
		t.Fatal(err)
	}
}

type closeClientErrorStub redisClientStub

func (c closeClientErrorStub) LRem(key string, count int64, value interface{}) *redis.IntCmd {
	return redis.NewIntResult(0, errors.New("LRem failed"))
}

func TestCloseError(t *testing.T) {
	rs := RedisRunStore{testInfo, &closeClientErrorStub{}}
	err := rs.Close("run:abc")
	if err.Error() != "LRem failed" {
		t.Errorf("redis error was not forwarded")
	}
}
