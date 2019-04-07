#! /bin/bash
# Usage:
#      bash scripts/test_endpoints.sh

set -e

echo "Building and deploying frontend"
make frontend_deploy
sleep 30

echo "Hitting /FuzzTarget"
CONTAINERS=$(cat target_configs/discourse.json)
TARGET_META=$(curl -d "$CONTAINERS" http://35.238.131.114:30080/FuzzTarget)

TARGET_ID=$(jq .TargetID <<< $TARGET_META)

if [ -z “$TARGET_ID” ]; then
    echo "Error spawning target"; exit
fi
echo "spawned pod with target id: $TARGET_ID"
sleep 120
EXCEPTIONS=$(curl http://35.238.131.114:30080/Exceptions/$TARGET_ID)
echo "EXCEPTIONS: $exceptions"
