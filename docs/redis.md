# Redis
This documentation explains the redis structure.

## Role of redis
Redis is used for message passing and transient storage. It contains jobs specs and status, and channels for jobs events.

## Keys
- **runs**: List containing all runs keys. It needs to be a list to ensure runs are ordered in descending order of creation.
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
- CANCELED: The run was canceled.
```
- **jobs:run:\<uid\>**: List containing all jobs keys for a run. It needs to be a list to ensure they are always ordered correctly.
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
- SKIPPED: The job's dependencies conditions were not met, and the job was skipped.
- RUNNING: The job is running on Kubernetes.
- SUCCESSFUL: The job has completed successfully.
- FAILED: The job has completed with an error.
```
- **dependencies:job:\<name\>:run:\<uid\>**: Set containing all dependencies keys for a job.
- **dependency:\<index\>:job:\<name\>:run:\<uid\>**: Hash containing a single dependency for a job. `index` is the index of the dependency job. The hash contains the following fields:
```
job: string: Key of the dependency job.
failure: true|false: If set to true, the job will only be run if the dependency fails. If set to false, the job will only be run if the dependency succeeds.
```
- **runs:work**: List containing the pending runs, formatted as `run:<uid>`. This list is consumed by workers.
- **events:runs:notif**: List containing events on runs, formatted as `event:<event>:run:<uid>`. This list is consumed by notifiers.
- **events:jobs:notif**: List containing events on jobs, formatted as `event:<event>:job<name>:run:<uid>`. This list is consumed by notifiers.

## PubSub
- **status:job:\<name\>:run:\<uid\>**: Channel passing the completion status of a job, formatted as `<status>`. This channel is consumed by workers, in the dependency resolution.
