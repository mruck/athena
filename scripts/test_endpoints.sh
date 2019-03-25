#! /bin/bash
# Usage:
#      bash scripts/test_endpoints.sh

set -e
#make frontend_deploy
#sleep 120
## Upload target
containers=$(cat containers.json)
pod=$(curl -d "$containers" http://35.192.59.73:30080/FuzzTarget)
if [ -z “$pod” ]; then
    echo "Error spawning target"; echo $pod; exit
fi
echo "spawned pod: $pod"
echo "Sleeping 60 before polling exceptions..."
sleep 60
exceptions=$(curl http://35.192.59.73:30080/Exceptions/$pod)
echo "exceptions: $exceptions"
