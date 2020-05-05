// Package worker contains tools to interact with
// the runs store, and process runs.
// A worker consumes runs, starts their jobs on Kubernetes,
// monitors them and updates their status.
package worker

import (
	"errors"
	"log"
	"sync"
)

type Worker struct {
	rs RunStore
	cp CloudProvider
	es EventStore
}

type RunStore interface {
	// Actively listens to new runs, and when a new run is available,
	// returns an arbitrary string identifier referencing it.
	// This identifier is used to refer to the run in store operations.
	NextRun() (string, error)

	// Persists the run status in the store.
	// Status can be:
	// - PENDING
	// - RUNNING
	// - SUCCESSFUL
	// - FAILED
	// - CANCELED
	SetRunStatus(runID, status string) error

	// Returns a list of arbitrary string identifiers referencing all
	// jobs contained in the run.
	// A job identifier must be globally unique, meaning that "job1" from "run1"
	// and "job1" from "run2" must have different identifiers.
	GetJobs(runID string) ([]string, error)

	// Returns the Job structure corresponding to the identifier.
	GetJob(jobID string) (Job, error)

	// Persists the job status in the store.
	// Status can be:
	// - PENDING
	// - SKIPPED
	// - RUNNING
	// - SUCCESSFUL
	// - FAILED
	SetJobStatus(jobID, status string) error

	// Returns a list of arbitrary string identifiers referencing all
	// dependencies for the job.
	// A dependency identifier must be globally unique.
	GetJobDependencies(jobID string) ([]JobDependency, error)
}

type Job struct {
	Name  string
	Image string
	Run   string
}

type JobDependency struct {
	JobID         string
	ExpectFailure bool
}

type CloudProvider interface {
	// Runs the job on the cloud provider.
	// Blocks until the job completes.
	RunJob(job Job) error
}

type EventStore interface {
	// Creates a new event in the store,
	// ready to be consumed.
	CreateEvent(event Event) error
}

type Event struct {
	Type    string
	Title   string
	Message string
}

func New() Worker {
	return Worker{
		NewRedisRunStore(),
		NewK8SCloudProvider(),
		NewRedisEventStore(),
	}
}

type mutexedJobStatus struct {
	L      sync.Locker
	Status string
}

type dependencyMap map[string](*mutexedJobStatus)

func newDependencyMap(jobIDs []string) dependencyMap {
	dm := make(dependencyMap)

	for _, jobID := range jobIDs {
		m := &sync.Mutex{}
		m.Lock()
		dm[jobID] = &mutexedJobStatus{m, "PENDING"}
	}

	return dm
}

// Wait for the job with identifier jobID to finish,
// and return its status.
func (dm dependencyMap) Wait(jobID string) string {
	mjs, ok := dm[jobID]
	if !ok {
		log.Println("Wait: job", jobID, "was not found in the dependency tree")
		return "FAILED"
	}

	mjs.L.Lock()
	status := mjs.Status
	mjs.L.Unlock()
	return status
}

// Tells all the goroutines waiting for jobID that it completed with the given
// status.
func (dm dependencyMap) Broadcast(jobID, status string) {
	mjs, ok := dm[jobID]
	if !ok {
		log.Println("Broadcast: job", jobID, "was not found in the dependency tree")
		return
	}

	mjs.Status = status
	mjs.L.Unlock()
}

// Returns the overall status of all jobs.
// If at least one job fails, the overall status is failed.
func (dm dependencyMap) Status() string {
	for _, mjs := range dm {
		if mjs.Status == "FAILED" {
			return "FAILED"
		}
	}

	return "SUCCESSFUL"
}

// Start launches the worker loop.
// It stays running as long as there is no internal error while processing
// the next run.
// In case of internal error, it waits for all running goroutines to finish.
func (w Worker) Start() error {
	var wg sync.WaitGroup

	var err error = nil
	for err == nil {
		err = w.ProcessNextRun(&wg)
	}

	wg.Wait()
	return err
}

// ProcessNextRun is a blocking function, listening for a new run,
// and processing it in a goroutine.
func (w Worker) ProcessNextRun(wg *sync.WaitGroup) error {
	runID, err := w.rs.NextRun()
	if err != nil {
		return err
	}

	wg.Add(1)
	go w.processRun(wg, runID)
	return nil
}

// ProcessRun blocks until the run is completed.
// It should be called in a specific goroutine.
func (w Worker) processRun(rwg *sync.WaitGroup, runID string) {
	defer rwg.Done()

	status := "CANCELED"
	defer func() { w.setRunStatus(runID, status) }()

	jobIDs, err := w.rs.GetJobs(runID)
	if err != nil {
		log.Println("Unable to get run", runID, "jobs:", err.Error())
		status = "FAILED"
		return
	}

	if err := w.checkDependencyTree(jobIDs); err != nil {
		log.Printf("Dependency tree check failed: %v", err.Error())
		status = "FAILED"
		return
	}

	if err := w.startRun(runID); err != nil {
		log.Printf("Unable to start run %v: %v", runID, err.Error())
		status = "FAILED"
		return
	}

	status = w.processJobs(jobIDs)
}

// SetRunStatus update the run status in the run store,
// and manages status updates on recovery, to indicate that the run failed.
// Note that on recovery, only the run status is updated. If the run status is
// CANCELED, one can assume that something went wrong on the server during run
// processing.
func (w Worker) setRunStatus(runID, status string) {
	r := recover()
	if r != nil {
		log.Println("Run", runID, "processing was interrupted by a panic:", r)
	}

	log.Println("Run", runID, "completed with status", status)
	if err := w.rs.SetRunStatus(runID, status); err != nil {
		log.Printf("Unable to set run %v status to %v: %v", runID, status, err.Error())
	}

	var event Event
	switch status {
	case "SUCCESSFUL":
		event = Event{"SUCCESS", "A run completed successfully", "Run with id " + runID + " completed successfully."}
	case "FAILED":
		event = Event{"FAILURE", "A run failed", "Run with id " + runID + " failed."}
	}
	if err := w.es.CreateEvent(event); err != nil {
		log.Println("Unable to create event for run completion:", err.Error())
	}

	if r != nil {
		panic(r)
	}
}

func (w Worker) checkDependencyTree(jobIDs []string) error {
	for _, jobID := range jobIDs {
		if err := w.checkDependencyPath([]string{}, jobID, jobIDs); err != nil {
			return err
		}
	}

	return nil
}

func (w Worker) checkDependencyPath(path []string, jobID string, jobIDs []string) error {
	deps, err := w.rs.GetJobDependencies(jobID)
	if err != nil {
		return err
	}

	for _, dep := range deps {
		if !contains(jobIDs, dep.JobID) {
			return errors.New("job " + dep.JobID + " was not found in the dependency tree")
		}

		if contains(path, dep.JobID) {
			return errors.New("dependency loop found in job " + dep.JobID)
		}

		subPath := make([]string, len(path), len(path)+1)
		copy(subPath, path)
		subPath = append(subPath, jobID)
		if err := w.checkDependencyPath(subPath, dep.JobID, jobIDs); err != nil {
			return err
		}
	}

	return nil
}

func contains(arr []string, val string) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
}

func (w Worker) startRun(runID string) error {
	log.Println("Starting run", runID)

	if err := w.rs.SetRunStatus(runID, "RUNNING"); err != nil {
		return err
	}
	if err := w.es.CreateEvent(Event{"START", "A run started", "A new run with id " + runID + " is processing."}); err != nil {
		return err
	}

	return nil
}

func (w Worker) processJobs(jobIDs []string) string {
	dm := newDependencyMap(jobIDs)
	var jwg sync.WaitGroup
	for _, jobID := range jobIDs {
		jwg.Add(1)
		go w.processJob(&jwg, dm, jobID)
	}
	jwg.Wait()

	return dm.Status()
}

func (w Worker) processJob(wg *sync.WaitGroup, dm dependencyMap, jobID string) {
	defer wg.Done()

	status := "SUCCESSFUL"
	defer func() { w.setJobStatus(dm, jobID, status) }()

	if err := w.waitJobDependencies(jobID, dm); err != nil {
		log.Println("Conditions for job", jobID, "are not met:", err.Error())
		status = "SKIPPED"
		return
	}

	if err := w.runJob(jobID); err != nil {
		log.Println("Job", jobID, "failed:", err.Error())
		status = "FAILED"
	}
}

func (w Worker) setJobStatus(dm dependencyMap, jobID string, status string) {
	log.Println("Job", jobID, "completed with status", status)
	if err := w.rs.SetJobStatus(jobID, status); err != nil {
		log.Printf("Unable to set job %v status to %v: %v", jobID, status, err.Error())
	}

	var event Event
	switch status {
	case "SUCCESSFUL":
		event = Event{"SUCCESS", "A job completed successfully", "Job with id " + jobID + " completed successfully."}
	case "FAILED":
		event = Event{"FAILURE", "A job failed", "Job with id " + jobID + " failed."}
	}
	if err := w.es.CreateEvent(event); err != nil {
		log.Println("Unable to create event for job completion:", err.Error())
	}

	dm.Broadcast(jobID, status)
}

func (w Worker) waitJobDependencies(jobID string, dm dependencyMap) error {
	deps, err := w.rs.GetJobDependencies(jobID)
	if err != nil {
		return err
	}

	for _, dep := range deps {
		status := dm.Wait(dep.JobID)
		if status == "SKIPPED" {
			return errors.New("dependency " + dep.JobID + " was skipped")
		}
		if status != "SUCCESSFUL" && !dep.ExpectFailure {
			return errors.New("dependency " + dep.JobID + " failed")
		}
		if status == "SUCCESSFUL" && dep.ExpectFailure {
			return errors.New("dependency " + dep.JobID + " did not fail")
		}
	}

	return nil
}

func (w Worker) runJob(jobID string) error {
	job, err := w.rs.GetJob(jobID)
	if err != nil {
		return err
	}

	log.Printf(`Starting job %v
	name: %v
	image: %v
	run: %v`, jobID, job.Name, job.Image, job.Run)

	if err := w.rs.SetJobStatus(jobID, "RUNNING"); err != nil {
		return err
	}
	if err := w.es.CreateEvent(Event{"START", "A job started", "Job with id " + jobID + " is processing."}); err != nil {
		return err
	}

	if err := w.cp.RunJob(job); err != nil {
		return err
	}

	return nil
}
