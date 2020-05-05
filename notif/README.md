# Notifier
The notifier analyzes events, and sends notifications on supported medias.

## Environment variables
The configuration is read through the environment. The following variables can be overridden:
- **REDIS_ADDR**: The redis address, in the format `hostname` or `hostname:port`. Default: `chainr-redis:6379`.
- **REDIS_PASSWORD**: The redis password. Default: `""` (no password).
- **REDIS_DB**: The redis database. Default: `0` (default db).

## Behaviour
Events are read from redis, and matched with the corresponding redis key.
For more information on the format stored in redis, see the [redis](../docs/redis.md) documentation.
