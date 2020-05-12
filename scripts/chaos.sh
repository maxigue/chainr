#!/bin/bash

set -e

usage() {
    echo "Usage: $0 [-f]" 1>&2
    echo "  -f: fast mode, kills pods more often"
    exit 1
}

fast=false

while getopts "f" option; do
    case "${option}" in
        f)
            fast=true
            ;;
        *)
            usage
            ;;
    esac
done

divide=1000
if [ "$fast" == "true" ]; then
    divide=5000
fi

while :; do
    sleep $(echo "scale=2; $RANDOM / $divide" | bc)
    pods=($(kubectl get pods -o "jsonpath={.items[*].metadata.name}"))
    nb=${#pods[@]}
    idx=$(( RANDOM % $nb ))
    pod=${pods[$idx]}
    kubectl delete pod $pod &
done
