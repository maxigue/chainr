# chainr
Chains Kubernetes jobs and provides a monitoring interface.

## About this project
This is only a pet project, whose initial ambition is only to help me learn golang. As such, it may contain clumsy code, shortcuts in the design and over-engineered parts.

## Use cases
Chainr is a general purpose scheduler, allowing to run jobs on a Kubernetes cluster. It supports parallel executions and fallbacks.
The following use cases can be considered:
- CI/CD
- Datawarehouse provisioning
- ...

## Concepts
- **Pipeline**: A Pipeline is the top-level unit, it contains Jobs to be run on Kubernetes.
- **Job**: A Job is the execution unit. It starts a docker container on Kubernetes and runs commands inside.
- **Run**: A Run allows to follow the status of a scheduled pipeline.

## Installing
Clone the repository, make sure to have kubectl installed and pointing to your target namespace, and run `make deploy`.
The default ingress host is `chainr.io`. It can be overridden in sched's [values.yaml](sched/deployments/helm/sched/values.yaml).
Alternatively, the service type can be set to `NodePort`.

## Example
The following JSON is a representation of a basic pipeline, containing two jobs run in parallel. The first job triggers a different job in case of success or error.
To schedule the pipeline, it can be sent as data to `POST /api/runs`.

```json
{
  "kind": "Pipeline",
  "jobs": {
    "first": {
      "image": "busybox",
      "run": "exit 0"
    },
    "second": {
      "image": "busybox",
      "run": "exit 0"
    },
    "success": {
      "dependsOn": [{
        "job": "first"
      }],
      "image": "busybox",
      "run": "exit 0"
    },
    "error": {
      "dependsOn": [{
        "job": "first",
        "conditions": {
          "failure": true
        }
      }],
      "image": "busybox",
      "run": "exit 0"
    }
  }
}
```

## Architecture
This project is architectured in micro-services.
- **gate**: Used as a gateway to all micro-services.
- **sched**: Allows to schedule pipeline runs and get run status.
- **work**: Worker running pipeline jobs on the kubernetes cluster.
- **notif**: Supports notification medias, and triggers notifications when events occur.
- **ui**: Serves the UI.

Message passing and transient persistence is done through redis.

## More documentation
More documentation can be found in the `docs/` directory.
- [Architecture](docs/architecture.md)
- [Redis](docs/redis.md)
