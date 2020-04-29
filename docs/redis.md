# Redis
This documentation explains the redis structure.

## Role of redis
Redis is used for message passing and transient storage. It contains jobs specs and status, and channels for jobs events.

## Keys
- **runs**: Set containing all runs keys.
- **run:\<uid\>**: Hash containing the run's status. The hash contains the following fields:
```
uid: string: The run UID.
status: status: The run status.
```
Status can be:
```
- PENDING: The run has not been consumed by a worker yet.
- RUNNING: The run is being processed by a worker.
- SUCCESSFUL: The run has completed successfully.
- FAILED: The run has completed with an error.
```
- **jobs:run:\<uid\>**: Set containing all jobs keys for a run.
- **job:\<name\>:run:\<uid\>**: Hash containing the job's spec and status. A new key is created for each job. The run uid is set as a suffix to allow searchs by run. The hash contains the following fields:
```
name: string: The job name.
image: string: The docker image to use.
run: string: The command to run.
status: status: The job status.
```
Status can be:
```
- PENDING: The job has not been started yet.
- RUNNING: The job is running on Kubernetes.
- SUCCESSFUL: The job has completed successfully.
- FAILED: The job has completed with an error.
```
- **dependencies:job:\<name\>:run:\<uid\>**: Set containing all dependencies keys for a job.
- **dependency:\<dep\>:job:\<name\>:run:\<uid\>**: Hash containing a single dependency for a job. `dep` is the jame of the dependency job. The hash contains the following fields:
```
failure: true|false: If set to true, the job will only be run if the dependency fails. If set to false, the job will only be run if the dependency succeeds.
```
- **runs:work**: List containing the pending runs, formatted as `run:<uid>`. This list is consumed by workers.
- **events:runs:notif**: List containing events on runs, formatted as `event:<event>:run:<uid>`. This list is consumed by notifiers.
- **events:jobs:notif**: List containing events on jobs, formatted as `event:<event>:job<name>:run:<uid>`. This list is consumed by notifiers.

## PubSub
- **status:job:\<name\>:run:\<uid\>**: Channel passing the completion status of a job, formatted as `<status>`. This channel is consumed by workers, in the dependency resolution.
