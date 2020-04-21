#!/bin/bash

set -e

mkdir -p deployments/helm/tmp
cd deployments/helm/tmp
helm package ../storage
helm template --name=chainr *.tgz | kubectl apply -f -
