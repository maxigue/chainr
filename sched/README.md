# Scheduler
The scheduler allows to schedule pipeline runs, and get status.

## Environment variables
The configuration is read through the environment. The following variables can be overridden:
- **PORT**: The port the server listens on. Default: `8080`.
- **REDIS_ADDR**: The redis address, in the format `hostname` or `hostname:port`. Default: `chainr-redis:6379`.
- **REDIS_ADDRS**: The redis address list, used when failover is setup in redis, in the format `hostname1 hostname2 hostnameN` or `hostname1:port hostname2:port hostnameN:port`. Default: value of **REDIS_ADDR**.
- **REDIS_MASTER**: The name of the master when failover is setup in redis.
- **REDIS_PASSWORD**: The redis password. Default: `""` (no password).
- **REDIS_DB**: The redis database. Default: `0` (default db).
