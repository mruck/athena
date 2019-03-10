#! /bin/bash
# Usage:
#      bash scripts/sanity.sh
#      bash scripts/sanity.sh all
set -e

function __is_pod_ready() {
  ready=$(kubectl get po "$1" -o 'jsonpath={.status.conditions[?(@.type=="Ready")].status}')
  [ "$ready" = "True" ] && echo "OK"
}

branch=$(git branch | grep \* | cut -d ' ' -f 2)
ref=$(git rev-parse $branch | tr -d '\n')
echo "Git branch/ref: $branch/$ref"
GIT_SHA=$(git log | head -n 1 | cut -f 2 -d ' ' | head -c 10)
ORIGINAL_GIT_SHA=$GIT_SHA

if [ "$(git diff --shortstat 2> /dev/null | tail -n1)" != "" ]; then
    GIT_SHA=$GIT_SHA-$RANDOM
fi
POD_NAME=$GIT_SHA-$RANDOM

# Build img and tag with git sha
make fuzzer-client
docker tag gcr.io/athena-fuzzer/athena:$ORIGINAL_GIT_SHA gcr.io/athena-fuzzer/athena:$GIT_SHA
docker push gcr.io/athena-fuzzer/athena:$GIT_SHA

mkdir -p /tmp/sanity/$POD_NAME

# Update image to reflect sha
# Update pod name to reflect sha
jq '.spec.containers[2].image = "gcr.io/athena-fuzzer/athena:'$GIT_SHA'"' pods/pod_sanity_template.json | \
    jq '.metadata.name = "'$POD_NAME'"' > /tmp/sanity/$POD_NAME/pod.json 
kubectl apply -f /tmp/sanity/$POD_NAME/pod.json 

# Wait for pod to be created
while [ ! "$(__is_pod_ready $POD_NAME)" = "OK" ]; do echo "Polling pod..."; sleep 1; done

echo "Tail logs of client at /tmp/sanity/$POD_NAME/client"
(kubectl logs -f $POD_NAME athena  2>&1) > /tmp/sanity/$POD_NAME/client

kubectl delete pod $POD_NAME

# Parse the client logs for run info. Info should look like this:
# Code Counts: {"200": 174, "404": 79, "500": 9}
# Final Coverage: 45.955721482311
# Success Ratio: 0.5667752442996743
# Total requests: 307
cnt=$(cat /tmp/sanity/$POD_NAME/client | grep "Code Counts" | cut -d ':' -f 2- | tr -d ' ' | tr -d '\r')
cov=$(cat /tmp/sanity/$POD_NAME/client | grep "Final Coverage" | cut -d ':' -f 2 | tr -d ' ' | tr -d '\r')
succ=$(cat /tmp/sanity/$POD_NAME/client | grep "Success Ratio" | cut -d ':' -f 2 | tr -d ' ' | tr -d '\r')
reqs=$(cat /tmp/sanity/$POD_NAME/client | grep "Total requests" | cut -d ':' -f 2 | tr -d ' ' | tr -d '\r')
output="misc/sanity.txt"

mkdir -p misc
echo "Code counts: $cnt"
echo "Cov: $cov"
echo "Success Rate: $succ"
echo "Requests: $reqs"

echo "{}" | \
    jq ".git_ref = \"$ref\"" | \
    jq ".git_branch= \"$branch\"" | \
    jq ".coverage = $cov" | \
    jq ".success_rate = $succ" | \
    jq ".total_requests = $reqs" | \
    jq ".status_codes = $cnt" | \
    jq -c '.' | tee -a $output
