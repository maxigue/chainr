// Package pipeline contains the code handling pipelines, including HTTP
// handlers and the whole pipeline representation and execution.
package pipeline

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/qri-io/jsonschema"

	"github.com/Tyrame/chainr/sched/internal/httputil"
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
// If the pipeline is invalid (either because the format is invalid, or the
// pipeline validation fails), an error is returned.
func NewFromSpec(spec []byte) (*Pipeline, httputil.ErrorWithStatus) {
	var p Pipeline
	if err := json.Unmarshal(spec, &p); err != nil {
		return nil, httputil.NewErrorWithStatus(err, http.StatusBadRequest)
	}

	if errs, _ := schema.ValidateBytes(spec); len(errs) > 0 {
		arr := make([]string, 0, len(errs))
		for _, e := range errs {
			arr = append(arr, e.Error())
		}
		err := errors.New(strings.Join(arr, ", "))
		return nil, httputil.NewErrorWithStatus(err, http.StatusBadRequest)
	}

	if err := p.Validate(); err != nil {
		return nil, httputil.NewErrorWithStatus(err, http.StatusUnprocessableEntity)
	}

	return &p, nil
}

// Validates the dependency tree.
// Returns an error describing the problem or nil.
func (p *Pipeline) Validate() error {
	// TODO: implement
	return nil
}
