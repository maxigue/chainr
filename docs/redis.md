# Redis
This documentation explains the redis structure.

## Role of redis
Redis is used for message passing and transient storage. It contains jobs specs and status, and channels for jobs events.

## Keys
- **run:\<uid\>:job:\<name\>**: Hash containing the job's spec and status. A new key is created for each job. The run uid is set as a prefix to allow searchs by run.
- **work:jobs**: List containing the pending jobs, formatted as `run:<uid>:job:<name>`. This list is consumed by workers.
- **notif:jobs:events**: List containing events on jobs, formatted as `run:<uid>:job<name>:event:<event>`. This list is consumed by notifiers.

## PubSub
- **run:\<uid\>:job:\<name\>:status**: Channel passing the completion status of a job, formatted as `<status>`. This channel is consumed by workers, in the dependency resolution.
