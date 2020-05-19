# Worker
The worker processes pending jobs, manages dependencies, runs jobs on Kubernetes and update status and events.

## Environment variables
The configuration is read through the environment. The following variables can be overridden:
- **REDIS_ADDR**: The redis address, in the format `hostname` or `hostname:port`. Default: `chainr-redis:6379`.
- **REDIS_ADDRS**: The redis address list, used when failover is setup in redis, in the format `hostname1 hostname2 hostnameN` or `hostname1:port hostname2:port hostnameN:port`. Default: value of **REDIS_ADDR**.
- **REDIS_MASTER**: The name of the master when failover is setup in redis.
- **REDIS_PASSWORD**: The redis password. Default: `""` (no password).
- **REDIS_DB**: The redis database. Default: `0` (default db).
- **KUBECONFIG**: The kubeconfig file path. If not set , use the in-cluster configuration.  Default: `""`.

## Behaviour
Pending jobs are read from redis, and matched with the corresponding redis key.
For more information on the format stored in redis, see the [redis](../docs/redis.md) documentation.
