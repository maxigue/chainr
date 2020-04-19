// Package pipeline contains the code handling pipelines, including HTTP
// handlers and the whole pipeline representation and execution.
package pipeline

import (
	"encoding/json"
	"errors"
	"log"
	"strings"

	"github.com/qri-io/jsonschema"

	"github.com/Tyrame/chainr/sched/internal/httputil"
	"github.com/Tyrame/chainr/sched/internal/k8s"
)

type Pipeline struct {
	httputil.Kindable
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
	Job        string                `json:"job"`
	Conditions []ConditionDependency `json:"conditions"`
}

var jobDependencySchema = `{
	"type": "object",
	"properties": {
		"job": {
			"type": "string"
		},
		"conditions": {
			"type": "array",
			"items": ` + conditionDependencySchema + `
		}
	},
	"additionalProperties": false,
	"required": ["job"]
}`

type ConditionDependency struct {
	Failure bool `json:"failure"`
}

var conditionDependencySchema = `{
	"type": "object",
	"properties": {
		"failure": {
			"type": "boolean"
		}
	},
	"additionalProperties": false
}`

var schema = &jsonschema.RootSchema{}

// Initialize the JSON schema.
func init() {
	if err := json.Unmarshal([]byte(pipelineSchema), schema); err != nil {
		panic("unmarshal pipeline schema: " + err.Error())
	}
}

// Creates a pipeline with the minimal valid configuration.
func New() *Pipeline {
	return &Pipeline{
		Kindable: httputil.Kindable{"Pipeline"},
		Jobs:     make(map[string]Job),
	}
}

// Creates a pipeline from a JSON spec given as an array of bytes.
// If the spec has an invalid format, an error is returned.
func NewFromSpec(spec []byte) (*Pipeline, error) {
	var p Pipeline
	if err := json.Unmarshal(spec, &p); err != nil {
		return nil, err
	}

	if errs, _ := schema.ValidateBytes(spec); len(errs) > 0 {
		arr := make([]string, 0, len(errs))
		for _, e := range errs {
			arr = append(arr, e.Error())
		}
		return nil, errors.New(strings.Join(arr, ", "))
	}

	return &p, nil
}

// Runs the pipeline.
// This method can take a long time, so it should be called
// in a goroutine.
func (p *Pipeline) Run(k8s k8s.Client) {
	log.Println("Pipeline.Run()", k8s)
}
