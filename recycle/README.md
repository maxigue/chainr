# Recycler
The recycler collects items that were not fully processed by workers (e.g. due to outages), and re-schedules them.

## Environment variables
The configuration is read through the environment. The following variables can be overridden:
- **REDIS_ADDR**: The redis address, in the format `hostname` or `hostname:port`. Default: `chainr-redis:6379`.
- **REDIS_ADDRS**: The redis address list, used when failover is setup in redis, in the format `hostname1 hostname2 hostnameN` or `hostname1:port hostname2:port hostnameN:port`. Default: value of **REDIS_ADDR**.
- **REDIS_MASTER**: The name of the master when failover is setup in redis.
- **REDIS_PASSWORD**: The redis password. Default: `""` (no password).
- **REDIS_DB**: The redis database. Default: `0` (default db).

## How it works
When a worker takes an item from its worker queue for processing, it pushes it in a processing queue. The recycler checks periodically if all the registred workers are still live, and for workers that are no longer live, it re-schedules items from the processing queue to the worker queue.

## How to use
The recycler can be used by any worker satisfying the following requirements:
- The worker reads from a worker queue that is a redis list.
- The worker puts its processing items in a processing queue, which is a redis list.

To use the recycler, add a key (e.g. the name of the worker) in redis containing the following fields:
```
queue: string: The key of the queue consumed by the worker.
processQueue: The key of the processing queue filled when the worker takes an item.
expiry: ISO8601: The expiration date.
```
Once done, add a member to the `workers` set containing the previous key.

The recommended way to use the recycle service is to periodically update the expiry with a short deadline.
