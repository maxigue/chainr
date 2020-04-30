# Scheduler
The scheduler allows to schedule pipeline runs, and get status.

## Environment variables
The configuration is read through the environment. The following variables can be overridden:
- **PORT**: The port the server listens on. Default: `8080`.
- **REDIS_ADDR**: The redis address, in the format `hostname` or `hostname:port`. Default: `chainr-redis:6379`.
- **REDIS_PASSWORD**: The redis password. Default: `""` (no password).
- **REDIS_DB**: The redis database. Default: `0` (default db).
