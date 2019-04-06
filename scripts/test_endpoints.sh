#! /bin/bash
# Usage:
#      bash scripts/test_endpoints.sh

set -e

echo "Building and deploying frontend"
make frontend_deploy
sleep 30
containers=$(cat target_configs/discourse.json)
echo "Hitting /FuzzTarget"
target_id=$(curl -d "$containers" http://35.238.131.114:30080/FuzzTarget)
if [ -z “$target_id” ]; then
    echo "Error spawning target"; exit
fi
echo "spawned pod with target id: $target_id"
sleep 120
exceptions=$(curl http://35.238.131.114:30080/Exceptions/$target_id)
echo "exceptions: $exceptions"
