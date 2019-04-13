#! /bin/bash
# Usage:
#      bash scripts/test_endpoints.sh

set -ex

function __is_pod_ready() {
  ready=$(kubectl get po "$1" -o 'jsonpath={.status.conditions[?(@.type=="Ready")].status}')
  [ "$ready" = "True" ] && echo "OK"
}
branch=$(git branch | grep \* | cut -d ' ' -f 2)
ref=$(git rev-parse $branch | tr -d '\n')

echo "Building and deploying frontend"
make frontend_deploy
sleep 30

echo "Hitting /FuzzTarget"
IP_ADDR=104.154.144.253
CONTAINERS=$(cat target_configs/discourse.json)
TARGET_META=$(curl -d "$CONTAINERS" http://$IP_ADDR:30080/FuzzTarget)

TARGET_ID=$(jq -r .TargetID <<< $TARGET_META)
POD_NAME=$(jq -r .PodName <<< $TARGET_META)

if [ -z “$TARGET_ID” ]; then
    echo "Error spawning target"; exit
fi
echo "spawned pod $POD_NAME with target id $TARGET_ID"


# Wait for pod to be created
while [ ! "$(__is_pod_ready $POD_NAME)" = "OK" ]; do echo "Polling pod..."; sleep 1; done

echo "Tail logs of client at /tmp/sanity/$POD_NAME/client"
mkdir -p /tmp/sanity/$POD_NAME
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

echo "Looking up exceptions..."
EXCEPTIONS=$(curl http://$IP_ADDR:30080/Exceptions/$TARGET_ID)
echo "EXCEPTIONS: $EXCEPTIONS"
