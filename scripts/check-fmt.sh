#!/bin/bash

set -e

gofmt -d .
exit $([[ $(gofmt -d . | wc -l) -eq 0 ]])
