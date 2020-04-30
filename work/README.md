# Worker
The worker processes pending jobs, manages dependencies, runs jobs on Kubernetes and update status and events.

## Environment variables
The configuration is read through the environment. The following variables can be overridden:
- **REDIS_ADDR**: The redis address, in the format `hostname` or `hostname:port`. Default: `redis:6379`.
- **REDIS_PASSWORD**: The redis password. Default: `""` (no password).
- **REDIS_DB**: The redis database. Default: `0` (default db).
- **KUBECONFIG**: The kubeconfig file path. If not set , use the in-cluster configuration.  Default: `""`.

## Behaviour
Pending jobs are read from redis, and matched with the corresponding redis key.
For more information on the format stored in redis, see the [redis](../docs/redis.md) documentation.
