package k8s

import "testing"

func TestNew(t *testing.T) {
	_, err := New("testdata/kubeconfig")
	if err != nil {
		t.Errorf("err is not-nil, expected nil\n%v", err)
	}
}

func TestNewFail(t *testing.T) {
	_, err := New("testdata/kubeconfig_invalid")
	if err == nil {
		t.Errorf("err is nil, expected non-nil")
	}
}

func TestNewStub(t *testing.T) {
	stub := NewStub()
	_ = stub.(Client)
}
