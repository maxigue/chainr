package worker

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"errors"
	"sync"
)

// As the worker mostly works as a black box,
// most tests in this file are written with stubs.
// To embed inline documentation on the behaviour,
// behaviour-driven testing is still present, despite
// not being perfectly suited.

type brokenRunStoreStub struct{}

func (rs brokenRunStoreStub) NextRun() (string, error) {
	return "", errors.New("failed")
}
func (rs brokenRunStoreStub) SetRunStatus(runId, status string) error {
	return nil
}
func (rs brokenRunStoreStub) GetJobs(runID string) ([]string, error) {
	return []string{}, nil
}
func (rs brokenRunStoreStub) GetJob(jobID string) (Job, error) {
	return Job{}, nil
}
func (rs brokenRunStoreStub) SetJobStatus(jobID, status string) error {
	return nil
}
func (rs brokenRunStoreStub) GetJobDependencies(jobID string) ([]JobDependency, error) {
	return []JobDependency{}, nil
}

type cloudProviderStub struct{}

func (cp cloudProviderStub) RunJob(job Job) error {
	return nil
}

type eventStoreStub struct{}

func (es eventStoreStub) CreateEvent(event Event) error {
	return nil
}

func TestStartError(t *testing.T) {
	Convey("Scenario: the runs can not be retrieved", t, func() {
		Convey("Given the worker can not access the runs queue", func() {
			w := Worker{&brokenRunStoreStub{}, &cloudProviderStub{}, &eventStoreStub{}}

			Convey("When the worker tries to process the next run", func() {
				err := w.Start()

				Convey("The worker should stop its loop with an error to avoid spamming", func() {
					So(err.Error(), ShouldEqual, "failed")
				})
			})
		})
	})
}

type runStoreDepMock struct {
	t             *testing.T
	setRunStatusI int
	setJobStatusI int
}

func (rs *runStoreDepMock) NextRun() (string, error) {
	return "run:abc", nil
}
func (rs *runStoreDepMock) SetRunStatus(runId, status string) error {
	expectedStatus := ""

	switch rs.setRunStatusI {
	case 0:
		expectedStatus = "RUNNING"
	case 1:
		expectedStatus = "SUCCESSFUL"
	}

	if status != expectedStatus {
		rs.t.Errorf("SetRunStatus = %v, expected %v", status, expectedStatus)
	}

	rs.setRunStatusI++
	return nil
}
func (rs *runStoreDepMock) GetJobs(runID string) ([]string, error) {
	return []string{"job:job1:run:abc", "job:dep1:run:abc"}, nil
}
func (rs *runStoreDepMock) GetJob(jobID string) (Job, error) {
	return Job{"", "busybox", "exit 0"}, nil
}
func (rs *runStoreDepMock) SetJobStatus(jobID, status string) error {
	expectedStatus := ""

	switch rs.setJobStatusI {
	case 0:
		expectedStatus = "RUNNING"
	case 1:
		expectedStatus = "SUCCESSFUL"
	case 2:
		expectedStatus = "RUNNING"
	case 3:
		expectedStatus = "SUCCESSFUL"
	}

	if status != expectedStatus {
		rs.t.Errorf("SetJobStatus = %v, expected %v", status, expectedStatus)
	}

	rs.setJobStatusI++
	return nil
}
func (rs *runStoreDepMock) GetJobDependencies(jobID string) ([]JobDependency, error) {
	deps := []JobDependency{}

	if jobID == "job:job1:run:abc" {
		deps = append(deps, JobDependency{"job:dep1:run:abc", false})
	}

	return deps, nil
}

func TestProcessNextRun(t *testing.T) {
	Convey("Scenario: process a valid run", t, func() {
		Convey("Given a run is processed", func() {
			Convey("When its dependency tree is valid, and everything goes well", func() {
				Convey("The worker should run each job according to the dependency tree, and set statuses to SUCCESSFUL", func() {
					w := Worker{&runStoreDepMock{t: t}, &cloudProviderStub{}, &eventStoreStub{}}
					var wg sync.WaitGroup
					w.ProcessNextRun(&wg)
					wg.Wait()
				})
			})
		})
	})
}

type runStoreFailureMock struct {
	t             *testing.T
	setRunStatusI int
	setJobStatusI int
}

func (rs *runStoreFailureMock) NextRun() (string, error) {
	return "run:abc", nil
}
func (rs *runStoreFailureMock) SetRunStatus(runId, status string) error {
	expectedStatus := ""

	switch rs.setRunStatusI {
	case 0:
		expectedStatus = "RUNNING"
	case 1:
		expectedStatus = "FAILED"
	}

	if status != expectedStatus {
		rs.t.Errorf("SetRunStatus = %v, expected %v", status, expectedStatus)
	}

	rs.setRunStatusI++
	return nil
}
func (rs *runStoreFailureMock) GetJobs(runID string) ([]string, error) {
	return []string{"job:job1:run:abc", "job:dep1:run:abc"}, nil
}
func (rs *runStoreFailureMock) GetJob(jobID string) (Job, error) {
	if jobID == "job:dep1:run:abc" {
		return Job{"", "busybox", "exit 1"}, nil
	}
	return Job{"", "busybox", "exit 0"}, nil
}
func (rs *runStoreFailureMock) SetJobStatus(jobID, status string) error {
	expectedStatus := ""

	switch rs.setJobStatusI {
	case 0:
		expectedStatus = "RUNNING"
	case 1:
		expectedStatus = "FAILED"
	case 2:
		expectedStatus = "RUNNING"
	case 3:
		expectedStatus = "SUCCESSFUL"
	}

	if status != expectedStatus {
		rs.t.Errorf("SetJobStatus = %v, expected %v", status, expectedStatus)
	}

	rs.setJobStatusI++
	return nil
}
func (rs *runStoreFailureMock) GetJobDependencies(jobID string) ([]JobDependency, error) {
	deps := []JobDependency{}

	if jobID == "job:job1:run:abc" {
		deps = append(deps, JobDependency{"job:dep1:run:abc", true})
	}

	return deps, nil
}

type cloudProviderFailureStub struct{}

func (cp cloudProviderFailureStub) RunJob(job Job) error {
	if job.Run == "exit 1" {
		return errors.New("failure")
	}
	return nil
}

func TestProcessNextRunFailure(t *testing.T) {
	Convey("Scenario: process run with failed jobs", t, func() {
		Convey("Given a run is processed", func() {
			Convey("When a job fails in the dependency tree", func() {
				Convey("Subsequent jobs should be run if expecting a failure", func() {
					Convey("And run should be set as failed", func() {
						w := Worker{&runStoreFailureMock{t: t}, &cloudProviderFailureStub{}, &eventStoreStub{}}
						var wg sync.WaitGroup
						w.ProcessNextRun(&wg)
						wg.Wait()
					})
				})
			})
		})
	})
}

type runStoreSkippedMock struct {
	t             *testing.T
	setRunStatusI int
	setJobStatusI int
}

func (rs *runStoreSkippedMock) NextRun() (string, error) {
	return "run:abc", nil
}
func (rs *runStoreSkippedMock) SetRunStatus(runId, status string) error {
	expectedStatus := ""

	switch rs.setRunStatusI {
	case 0:
		expectedStatus = "RUNNING"
	case 1:
		expectedStatus = "SUCCESSFUL"
	}

	if status != expectedStatus {
		rs.t.Errorf("SetRunStatus = %v, expected %v", status, expectedStatus)
	}

	rs.setRunStatusI++
	return nil
}
func (rs *runStoreSkippedMock) GetJobs(runID string) ([]string, error) {
	return []string{"job:job1:run:abc", "job:dep1:run:abc", "job:dep2:run:abc"}, nil
}
func (rs *runStoreSkippedMock) GetJob(jobID string) (Job, error) {
	return Job{"", "busybox", "exit 0"}, nil
}
func (rs *runStoreSkippedMock) SetJobStatus(jobID, status string) error {
	expectedStatus := ""

	switch rs.setJobStatusI {
	case 0:
		expectedStatus = "RUNNING"
	case 1:
		expectedStatus = "SUCCESSFUL"
	case 2:
		expectedStatus = "SKIPPED"
	case 3:
		expectedStatus = "SKIPPED"
	}

	if status != expectedStatus {
		rs.t.Errorf("SetJobStatus on %v = %v, expected %v", jobID, status, expectedStatus)
	}

	rs.setJobStatusI++
	return nil
}
func (rs *runStoreSkippedMock) GetJobDependencies(jobID string) ([]JobDependency, error) {
	deps := []JobDependency{}

	switch jobID {
	case "job:job1:run:abc":
		deps = append(deps, JobDependency{"job:dep1:run:abc", false})
	case "job:dep1:run:abc":
		deps = append(deps, JobDependency{"job:dep2:run:abc", true})
	}

	return deps, nil
}

func TestProcessNextRunSkipped(t *testing.T) {
	Convey("Scenario: process run with skipped jobs", t, func() {
		Convey("Given a run is processed", func() {
			Convey("When the dependency tree contains jobs whose conditions are not met", func() {
				Convey("The jobs, and all subsequent jobs in the branch, should be skipped", func() {
					w := Worker{&runStoreSkippedMock{t: t}, &cloudProviderStub{}, &eventStoreStub{}}
					var wg sync.WaitGroup
					w.ProcessNextRun(&wg)
					wg.Wait()
				})
			})
		})
	})
}

type runStoreNotFoundMock struct {
	t             *testing.T
	setRunStatusI int
}

func (rs *runStoreNotFoundMock) NextRun() (string, error) {
	return "run:abc", nil
}
func (rs *runStoreNotFoundMock) SetRunStatus(runId, status string) error {
	expectedStatus := ""

	switch rs.setRunStatusI {
	case 0:
		expectedStatus = "FAILED"
	}

	if status != expectedStatus {
		rs.t.Errorf("SetRunStatus = %v, expected %v", status, expectedStatus)
	}

	rs.setRunStatusI++
	return nil
}
func (rs *runStoreNotFoundMock) GetJobs(runID string) ([]string, error) {
	return []string{"job:job1:run:abc"}, nil
}
func (rs *runStoreNotFoundMock) GetJob(jobID string) (Job, error) {
	return Job{"", "busybox", "exit 0"}, nil
}
func (rs *runStoreNotFoundMock) SetJobStatus(jobID, status string) error {
	rs.t.Errorf("SetJobStatus should not have been called")
	return nil
}
func (rs *runStoreNotFoundMock) GetJobDependencies(jobID string) ([]JobDependency, error) {
	return []JobDependency{
		JobDependency{"job:dep1:run:abc", false},
	}, nil
}

func TestProcessNextRunNotFound(t *testing.T) {
	Convey("Scenario: process run with not found dependencies", t, func() {
		Convey("Given a run is processed", func() {
			Convey("When the run contains references to unknown dependencies", func() {
				Convey("The run should be set to FAILED, and its jobs should not be run", func() {
					w := Worker{&runStoreNotFoundMock{t: t}, &cloudProviderStub{}, &eventStoreStub{}}
					var wg sync.WaitGroup
					w.ProcessNextRun(&wg)
					wg.Wait()
				})
			})
		})
	})
}

type runStoreDepLoopMock struct {
	t             *testing.T
	setRunStatusI int
}

func (rs *runStoreDepLoopMock) NextRun() (string, error) {
	return "run:abc", nil
}
func (rs *runStoreDepLoopMock) SetRunStatus(runId, status string) error {
	expectedStatus := ""

	switch rs.setRunStatusI {
	case 0:
		expectedStatus = "FAILED"
	}

	if status != expectedStatus {
		rs.t.Errorf("SetRunStatus = %v, expected %v", status, expectedStatus)
	}

	rs.setRunStatusI++
	return nil
}
func (rs *runStoreDepLoopMock) GetJobs(runID string) ([]string, error) {
	return []string{
		"job:job1:run:abc",
		"job:job2:run:abc",
		"job:job3:run:abc",
		"job:job4:run:abc",
	}, nil
}
func (rs *runStoreDepLoopMock) GetJob(jobID string) (Job, error) {
	return Job{"", "busybox", "exit 0"}, nil
}
func (rs *runStoreDepLoopMock) SetJobStatus(jobID, status string) error {
	rs.t.Errorf("SetJobStatus should not have been called")
	return nil
}
func (rs *runStoreDepLoopMock) GetJobDependencies(jobID string) ([]JobDependency, error) {
	deps := []JobDependency{}

	switch jobID {
	case "job:job1:run:abc":
		deps = append(deps, JobDependency{"job:job4:run:abc", false})
	case "job:job2:run:abc":
		deps = append(deps, JobDependency{"job:job1:run:abc", false})
	case "job:job3:run:abc":
		deps = append(deps, JobDependency{"job:job1:run:abc", false})
	case "job:job4:run:abc":
		deps = append(deps, JobDependency{"job:job2:run:abc", false})
		deps = append(deps, JobDependency{"job:job3:run:abc", false})
	}

	return deps, nil
}

func TestProcessNextRunDepLoop(t *testing.T) {
	Convey("Scenario: process run with dependency loop", t, func() {
		Convey("Given a run is processed", func() {
			Convey("When the run has a loop in its dependencies", func() {
				Convey("The run should be set to FAILED, and its jobs should not be run", func() {
					w := Worker{&runStoreDepLoopMock{t: t}, &cloudProviderStub{}, &eventStoreStub{}}
					var wg sync.WaitGroup
					w.ProcessNextRun(&wg)
					wg.Wait()
				})
			})
		})
	})
}
