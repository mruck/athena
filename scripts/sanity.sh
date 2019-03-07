#! /bin/bash
# Usage:
#      bash scripts/sanity.sh
#      bash scripts/sanity.sh all
set -ex

branch=$(git branch | grep \* | cut -d ' ' -f 2)
ref=$(git rev-parse $branch | tr -d '\n')
echo "Git branch/ref: $branch/$ref"
GIT_SHA=$(git log | head -n 1 | cut -f 2 -d ' ' | head -c 10)
#[ $(git diff --shortstat 2> /dev/null | tail -n1) != "" ] && GIT_SHA=$GIT_SHA-$RANDOM
POD_NAME=$GIT_SHA-$RANDOM

# Build img and tag with git sha
#make fuzzer-client
# docker push gcr.io/athena-fuzzer/athena:$GIT_SHA

mkdir -p /tmp/sanity/$POD_NAME

# Update image to reflect sha
# Update pod name to reflect sha
jq '.spec.containers[2].image = "gcr.io/athena-fuzzer/athena:'$GIT_SHA'"' pods/sanity_pod.json | \
    jq '.metadata.name = "'$POD_NAME'"' > /tmp/sanity/$POD_NAME/pod.json 
kubectl apply -f /tmp/sanity/$POD_NAME/pod.json 

echo "Tail logs of client at /tmp/sanity/client"
(kubectl logs -f $POD_NAME athena  2>&1) > /tmp/sanity/$POD_NAME/client

# Delete pod
#kubectl delete pod $POD_NAME

#output="misc/sanity.txt"
#if [ -z "$1" ]; then
#    (./orchestrate.py repro --client --stop_after_har $port 2>&1) > /tmp/sanity/$port/client
#elif [ "$1" = "all" ]; then
#    (./orchestrate.py repro --client --stop_after_all_routes $port 2>&1) > /tmp/sanity/$port/client
#    output="misc/all_routes.txt"
#fi
#
## Delete the pod
#

# Parse the client logs for run info. Info should look like this:
# Code Counts: {"200": 174, "404": 79, "500": 9}
# Final Coverage: 45.955721482311
# Success Ratio: 0.5667752442996743
# Total requests: 307
cnt=$(cat /tmp/sanity/$POD_NAME/client | grep "Code Counts" | cut -d ':' -f 2- | tr -d ' ' | tr -d '\r')
cov=$(cat /tmp/sanity$POD_NAME/client | grep "Final Coverage" | cut -d ':' -f 2 | tr -d ' ' | tr -d '\r')
succ=$(cat /tmp/sanity/$POD_NAME/client | grep "Success Ratio" | cut -d ':' -f 2 | tr -d ' ' | tr -d '\r')
reqs=$(cat /tmp/sanity/$POD_NAME/client | grep "Total requests" | cut -d ':' -f 2 | tr -d ' ' | tr -d '\r')

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
