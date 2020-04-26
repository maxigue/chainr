package run

import "testing"

func TestInit(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("initialization did not panic")
		}
	}()

	pipelineSchema = `{test}`
	initJSONSchema()
}

func TestNewPipeline(t *testing.T) {
	spec := []byte(`{
		"kind": "Pipeline",
		"jobs": {
			"job1": {
				"image": "busybox",
				"run": "exit 0"
			}
		}
	}`)

	p, err := NewPipeline(spec)
	if err != nil {
		t.Fatal("err = nil, expected not nil")
	}
	if image := p.Jobs["job1"].Image; image != "busybox" {
		t.Errorf("image = %v, expected busybox", image)
	}
}

func TestNewPipelineBadFormat(t *testing.T) {
	spec := []byte(`{invalid}`)
	_, err := NewPipeline(spec)
	if err == nil {
		t.Fatal("NewPipeline from an invalid format returned a nil error")
	}
}

func TestNewPipelineBadSchema(t *testing.T) {
	spec := []byte(`{
		"kind": "Pipeline",
		"invalid": "hello"
	}`)
	_, err := NewPipeline(spec)
	if err == nil {
		t.Fatal("NewPipeline from an invalid schema returned a nil error")
	}
}
