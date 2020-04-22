# Redis
This documentation explains the redis structure.

## Role of redis
Redis is used for message passing and transient storage. It contains jobs specs and status, and channels for jobs events.

## Keys
- **job:\<name\>:run:\<uid\>**: Hash containing the job's spec and status. A new key is created for each job. The run uid is set as a suffix to allow searchs by run. The hash contains the following fields:
```
image: string: The docker image to use.
run: string: The command to run.
status: status: The job status.
```
Status can be:
```
- PENDING: The job has not been consumed by a worker yet.
- WAITING: The job is waiting for its dependencies.
- RUNNING: The job is running on Kubernetes.
- SUCCESSFUL: The job has completed successfully.
- FAILED: The job has completed with an error.
```
- **dependency:\<dep\>:job:\<name\>:run:\<uid\>**: Hash containing a single dependency for a job. `dep` is the jame of the dependency job. The hash contains the following fields:
```
failure: true|false: If set to true, the job will only be run if the dependency fails. If set to false, the job will only be run if the dependency succeeds.
```
- **work:jobs**: List containing the pending jobs, formatted as `run:<uid>:job:<name>`. This list is consumed by workers.
- **notif:jobs:events**: List containing events on jobs, formatted as `run:<uid>:job<name>:event:<event>`. This list is consumed by notifiers.

## PubSub
- **status:job:\<name\>:run:\<uid\>**: Channel passing the completion status of a job, formatted as `<status>`. This channel is consumed by workers, in the dependency resolution.
