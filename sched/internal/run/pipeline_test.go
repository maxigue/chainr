package run

import "testing"

// Calling NewPipelineFactory() should not panic.
func TestNewPipelineFactory(t *testing.T) {
	_ = NewPipelineFactory()
}

func TestNewPipelineFactoryFail(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("PipelineFactory creation did not panic")
		}
	}()

	_ = newPipelineFactory([]byte(`{test}`))
}

func TestCreate(t *testing.T) {
	spec := []byte(`{
		"kind": "Pipeline",
		"jobs": {
			"job1": {
				"image": "busybox",
				"run": "exit 0"
			}
		}
	}`)

	p, err := NewPipelineFactory().Create(spec)
	if err != nil {
		t.Fatal("err = nil, expected not nil")
	}
	if image := p.Jobs["job1"].Image; image != "busybox" {
		t.Errorf("image = %v, expected busybox", image)
	}
}

func TestNewPipelineBadFormat(t *testing.T) {
	spec := []byte(`{invalid}`)
	_, err := NewPipelineFactory().Create(spec)
	if err == nil {
		t.Fatal("NewPipeline from an invalid format returned a nil error")
	}
}

func TestNewPipelineBadSchema(t *testing.T) {
	spec := []byte(`{
		"kind": "Pipeline",
		"invalid": "hello"
	}`)
	_, err := NewPipelineFactory().Create(spec)
	if err == nil {
		t.Fatal("NewPipeline from an invalid schema returned a nil error")
	}
}
