#!/bin/bash

set -xe

kubectl delete pod discourse || true
sleep 10
kubectl apply -f discourse.test.pod.json
sleep 60
kubectl port-forward discourse 8080:8080 > /dev/null 2>&1 &
sleep 10
go test -v
