package worker

import (
	"errors"
	"math/rand"
	"time"
)

type K8SCloudProvider struct{}

func NewK8SCloudProvider() K8SCloudProvider {
	return K8SCloudProvider{}
}

func (cp K8SCloudProvider) RunJob(job Job) error {
	// TODO: implement
	ms := 1000 + rand.Intn(2000)
	time.Sleep(time.Duration(ms) * time.Millisecond)

	if job.Run == "exit 1" {
		return errors.New("failed")
	}

	return nil
}
