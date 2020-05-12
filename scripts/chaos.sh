#!/bin/bash

set -e

while :; do
    sleep $(echo "scale=2; $RANDOM / 1000" | bc)
    pods=($(kubectl get pods -o "jsonpath={.items[*].metadata.name}"))
    nb=${#pods[@]}
    idx=$(( RANDOM % $nb ))
    pod=${pods[$idx]}
    kubectl delete pod $pod
done
