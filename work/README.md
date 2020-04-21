# Worker
The worker processes pending jobs, manages dependencies, runs jobs on Kubernetes and update status and events.

## Pending jobs
Pending jobs are read from redis on the `work:jobs` list, and matched with the corresponding redis key.
For more information on the format stored in redis, see the [redis](../docs/redis.md) documentation.
