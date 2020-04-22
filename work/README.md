# Worker
The worker processes pending jobs, manages dependencies, runs jobs on Kubernetes and update status and events.

## Environment variables
The configuration is read through the environment. The following variables can be overridden:
- **PORT**: The port the server listens on. Default: `8080`.
- **KUBECONFIG**: The kubeconfig file path. If not set , use the in-cluster configuration.  Default: `""`.

## Pending jobs
Pending jobs are read from redis on the `work:jobs` list, and matched with the corresponding redis key.
For more information on the format stored in redis, see the [redis](../docs/redis.md) documentation.
