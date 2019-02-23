#! /bin/bash
set -e

# Get a new fuzz duo; parse the output with some shell magic.
port=$(./orchestrate.py fuzz_duo 1 | grep "Saved" | cut -d ':' -f 2 | tr -d ' ')
echo "Port number of fuzz duo: $port"

mkdir -p /tmp/sanity/$port

echo "Tail logs of server at /tmp/sanity/$port/server"
((./orchestrate.py repro --server --load_db $port 2>&1) > /tmp/sanity/$port/server) &

# TODO remove this in favor of polling a health endpoint or smth.
echo "Sleeping for 30 seconds to wait for server"
sleep 30

echo "Tail logs of client at /tmp/sanity/$port/client"
(./orchestrate.py repro --client --load_db $port 2>&1) > /tmp/sanity/$port/client

# Remove the server container after the run is over.
docker rm -f "server_$port" > /dev/null

# Parse the client logs for run info. Info should look like this:
# Final Coverage: 45.95579621482311
# Success Ratio: 0.5667752442996743
# Code Counts: Counter({200: 174, 404: 79, 400: 16, 403: 15, 422: 12, 500: 9, 409: 2})
# Total requests: 307
cnt=$(cat /tmp/sanity/$port/client | grep "Code Counts" | cut -d ':' -f 2- | tr -d ' ')
cov=$(cat /tmp/sanity/$port/client | grep "Final Coverage" | cut -d ':' -f 2 | tr -d ' ')
succ=$(cat /tmp/sanity/$port/client | grep "Success Ratio" | cut -d ':' -f 2 | tr -d ' ')
reqs=$(cat /tmp/sanity/$port/client | grep "Total requests" | cut -d ':' -f 2 | tr -d ' ')

echo "Code counts: $cnt"
echo "Cov: $cov"
echo "Success Rate: $succ"
echo "Requests: $reqs"
