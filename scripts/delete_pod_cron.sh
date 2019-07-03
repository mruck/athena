#!/bin/bash

set -x

# delete all evicted pods
kubectl get pods | grep Evicted | awk '{print $1}' | xargs kubectl delete pod

# delete all stale sanity pods
kubectl get pods -l duo=sanity | awk '{print $1}' | xargs kubectl delete pod
