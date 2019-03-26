#! /bin/bash
# Usage:
#      bash scripts/test_endpoints.sh

set -e

echo "Building and deploying frontend"
make frontend_deploy
sleep 30
containers=$(cat scripts/target.json)
echo "Hitting /FuzzTarget"
pod=$(curl -d "$containers" http://35.238.131.114:30080/FuzzTarget)
if [ -z “$pod” ]; then
    echo "Error spawning target"; echo $pod; exit
fi
echo "spawned pod: $pod"
sleep 120
exceptions=$(curl http://35.238.131.114:30080/Exceptions/$pod)
echo "exceptions: $exceptions"
