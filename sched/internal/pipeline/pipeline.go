// Package pipeline contains the code handling pipelines, including HTTP
// handlers and the whole pipeline representation and scheduling.
package pipeline

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v7"
	"github.com/qri-io/jsonschema"
)

type Pipeline interface {
	Run(runUID string) error
}

type pipeline struct {
	Kind string         `json:"kind"`
	Jobs map[string]Job `json:"jobs"`
}

var pipelineSchema = `{
	"title": "Pipeline",
	"type": "object",
	"properties": {
		"kind": {
			"const": "Pipeline"
		},
		"jobs": {
			"type": "object",
			"properties": {},
			"additionalProperties": ` + jobSchema + `
		}
	},
	"additionalProperties": false,
	"required": ["kind", "jobs"]
}`

type Job struct {
	Image     string          `json:"image"`
	Run       string          `json:"run"`
	DependsOn []JobDependency `json:"dependsOn"`
}

var jobSchema = `{
	"type": "object",
	"properties": {
		"image": {
			"type": "string"
		},
		"run": {
			"type": "string"
		},
		"dependsOn": {
			"type": "array",
			"items": ` + jobDependencySchema + `
		}
	},
	"additionalProperties": false,
	"required": ["image", "run"]
}`

type JobDependency struct {
	Job        string                  `json:"job"`
	Conditions JobDependencyConditions `json:"conditions"`
}

var jobDependencySchema = `{
	"type": "object",
	"properties": {
		"job": {
			"type": "string"
		},
		"conditions": ` + jobDependencyConditionsSchema + `
	},
	"additionalProperties": false,
	"required": ["job"]
}`

type JobDependencyConditions struct {
	Failure bool `json:"failure"`
}

var jobDependencyConditionsSchema = `{
	"type": "object",
	"properties": {
		"failure": {
			"type": "boolean"
		}
	},
	"additionalProperties": false
}`

// Initialize the JSON schema and the redis client.
func init() {
	initJSONSchema()
	initRedisClient()
}

var schema = &jsonschema.RootSchema{}

func initJSONSchema() {
	if err := json.Unmarshal([]byte(pipelineSchema), schema); err != nil {
		panic("unmarshal pipeline schema: " + err.Error())
	}
}

var redisClient redis.Cmdable = nil

const pendingJobsChannel = "work:jobs"

func initRedisClient() {
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
			panic("evaluate redis db: " + err.Error())
		}
		db = d
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
}

// Creates a pipeline with the minimal valid configuration.
func New() Pipeline {
	return &pipeline{
		Kind: "Pipeline",
		Jobs: make(map[string]Job),
	}
}

// Creates a pipeline from a JSON spec given as an array of bytes.
// If the spec has an invalid format, an error is returned.
func NewFromSpec(spec []byte) (Pipeline, error) {
	if errs, _ := schema.ValidateBytes(spec); len(errs) > 0 {
		arr := make([]string, 0, len(errs))
		for _, e := range errs {
			arr = append(arr, e.Error())
		}
		return nil, errors.New(strings.Join(arr, ", "))
	}

	var p pipeline
	if err := json.Unmarshal(spec, &p); err != nil {
		return nil, err
	}

	return &p, nil
}

// The Run method schedules the pipeline for workers.
// It adds the run key in redis, and keys for each job and their dependencies.
// It also adds the jobs to the job queue.
func (p *pipeline) Run(runUID string) error {
	if err := p.scheduleRun(runUID); err != nil {
		return err
	}
	if err := p.scheduleJobs(runUID); err != nil {
		return err
	}
	return nil
}

func (p *pipeline) scheduleRun(runUID string) error {
	runKey := makeRunKey(runUID)
	return redisClient.Set(runKey, "0", 0).Err()
}

func (p *pipeline) scheduleJobs(runUID string) error {
	for k, v := range p.Jobs {
		jobKey := makeJobKey(runUID, k)
		fields := []interface{}{
			"image", v.Image,
			"run", v.Run,
			"status", "PENDING",
		}
		if err := redisClient.HSet(jobKey, fields...).Err(); err != nil {
			return err
		}
		if err := p.scheduleDependencies(runUID, k, v); err != nil {
			return err
		}
		if err := redisClient.LPush(pendingJobsChannel, jobKey).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (p *pipeline) scheduleDependencies(runUID string, jobName string, job Job) error {
	for _, dep := range job.DependsOn {
		depKey := makeJobDependencyKey(runUID, jobName, dep.Job)
		failure := "false"
		if dep.Conditions.Failure {
			failure = "true"
		}
		if err := redisClient.HSet(depKey, "failure", failure).Err(); err != nil {
			return err
		}
	}

	return nil
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
