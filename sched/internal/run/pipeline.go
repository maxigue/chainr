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

const pipelineSchema = `{
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

const jobSchema = `{
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

const jobDependencySchema = `{
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

const jobDependencyConditionsSchema = `{
	"type": "object",
	"properties": {
		"failure": {
			"type": "boolean"
		}
	},
	"additionalProperties": false
}`

// The PipelineFactory allows to create pipelines.
// It initializes the pipeline json schema.
type PipelineFactory struct {
	schema *jsonschema.RootSchema
}

func NewPipelineFactory() PipelineFactory {
	return newPipelineFactory([]byte(pipelineSchema))
}
func newPipelineFactory(schema []byte) PipelineFactory {
	rootSchema := &jsonschema.RootSchema{}

	if err := json.Unmarshal(schema, rootSchema); err != nil {
		panic("unmarshal pipeline schema: " + err.Error())
	}

	return PipelineFactory{rootSchema}
}

// Creates a pipeline from a JSON spec given as an array of bytes.
// If the spec has an invalid format, an error is returned.
func (pf PipelineFactory) Create(spec []byte) (Pipeline, error) {
	if errs, _ := pf.schema.ValidateBytes(spec); len(errs) > 0 {
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
