// Package pipeline contains the code handling pipelines, including HTTP
// handlers and the whole pipeline representation and execution.
package pipeline

type Pipeline struct {
	Jobs map[string]Job `json:"jobs"`
}

type Job struct {
	Image     string          `json:"image"`
	Run       string          `json:"run"`
	DependsOn []JobDependency `json:"dependsOn"`
}

type JobDependency struct {
	Job        string                `json:"job"`
	Conditions []ConditionDependency `json:"conditions"`
}

type ConditionDependency struct {
	Failure bool `json:"failure"`
}
