# chainr
Chains Kubernetes jobs and provides a monitoring interface.

## About this project
This is only a pet project, whose initial ambition is only to help me learn golang. As such, it may contain clumsy code, and shortcuts in the design.

## Use cases
Chainr is a general purpose scheduler, allowing to run jobs on a Kubernetes cluster. It supports parallel executions and fallbacks.
The following use cases can be considered:
- CI/CD
- Datawarehouse provisioning
- ...

## Concepts
- *Pipeline*: A pipeline is the top-level unit, it contains jobs.
- *Job*: A job is the execution unit. It starts a docker container on Kubernetes and runs commands inside.

## Example
The following YAML is a representation of a basic pipeline, containing two jobs run in parallel. The first job triggers a different job in case of success or error.

```yaml
jobs:
  - name: first
    image: busybox
    run: exit 0
  - name: second
    image: busybox
    run: exit 0
  - name: success
    dependsOn:
      - job: first
    image: busybox
    run: exit 0
  - name: error
    dependsOn:
      - job: first
        conditions:
          failure: true
    image: busybox
    run: exit 0
```
