package run

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/qri-io/jsonschema"
)

type Pipeline struct {
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

var schema = &jsonschema.RootSchema{}

// Initialize the JSON schema and the redis client.
func init() {
	initJSONSchema()
}

func initJSONSchema() {
	if err := json.Unmarshal([]byte(pipelineSchema), schema); err != nil {
		panic("unmarshal pipeline schema: " + err.Error())
	}
}

// Creates a pipeline from a JSON spec given as an array of bytes.
// If the spec has an invalid format, an error is returned.
func NewPipeline(spec []byte) (Pipeline, error) {
	if errs, _ := schema.ValidateBytes(spec); len(errs) > 0 {
		arr := make([]string, 0, len(errs))
		for _, e := range errs {
			arr = append(arr, e.Error())
		}
		return Pipeline{}, errors.New(strings.Join(arr, ", "))
	}

	var p Pipeline
	if err := json.Unmarshal(spec, &p); err != nil {
		return Pipeline{}, err
	}

	return p, nil
}
