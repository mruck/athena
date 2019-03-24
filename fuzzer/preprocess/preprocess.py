#!/usr/bin/env python3
import argparse
import collections
import json
import os
import urllib.parse as urlparse

import fuzzer.infer as infer
import fuzzer.preprocess.har as har
import fuzzer.preprocess.route_matching as route_matching

CWD = os.path.dirname(os.path.realpath(__file__))
# HAR_CORPUS = "har_inputs/small_har.json"
# HAR_CORPUS = "bigger_har.json"
HAR_CORPUS = os.path.join(CWD, "har_inputs/ryan-har.json")
ROUTES_PATH = "/discourse-fork/state/routes.json"


def load_routes():
    raw_routes = json.loads(open(ROUTES_PATH, "r").read())
    routes = [r["path"] for r in raw_routes]
    processed_routes = route_matching.process_routes(routes)
    for i in range(len(processed_routes)):
        processed_routes[i]["verb"] = raw_routes[i]["verb"]
    return processed_routes


def canonicalize_route(path, host, verb, processed_routes):
    """
    Map localhost:1234/user/567 -> /user/:id
    """
    # Strip host and port
    path = path.split(host)[1]
    return route_matching.find_route(processed_routes, path, verb)


def read_params(params_file):
    """
    Read params accessed by rails
    """
    params_accessed = []
    for json_line in params_file:
        json_line = json_line.strip()
        param_path = json.loads(json_line)
        params_accessed.append(param_path[0])
    return params_accessed


def map_params(params_accessed, har_params):
    final_params = []
    for p in params_accessed:
        # TODO: Handle this case
        # http://localhost:50121/topics/timings
        # Sent params: {'timings[1]': '1002', 'topic_time': '2003', 'topic_id': '10'}
        # Rails accessed: {'timings': {}}
        if isinstance(p, dict) or isinstance(p, list):
            continue
        # Params are {"name" : name, "values" : [values], "type" : type}
        param = {}
        param["name"] = p
        if har_params and p in har_params:
            param["values"] = [har_params[p]]
        else:
            param["values"] = [None]
        final_params.append(param)
    return final_params


# This performs an in-place update of the old params.
def update_params(old_params, new_params):
    for new_param in new_params:
        # Check if our new param is in old params
        old_param = next(
            (p for p in old_params if p["name"] == new_param["name"]), None
        )
        # Add our value to the list of values
        if old_param:
            old_param["values"].extend(new_param["values"])
        # This param hasn't been seen before
        else:
            old_params.append(new_param)


# returns the index of the route if we find a route with that
# path, otherwise returns None.
def find_in_routes(visited_routes, path, verb):
    return next(
        (
            index
            for (index, d) in enumerate(visited_routes)
            if d["path"] == path and d["verb"] == verb
        ),
        None,
    )


def infer_routes(visited_routes, results_path, fname):
    for route in visited_routes:
        for param in route["body_params"]:
            infer.infer(param)
        for param in route["query_params"]:
            infer.infer(param)
        for param in route["dynamic_segments"]:
            infer.infer(param)

    with open(os.path.join(results_path, fname), "w") as f:
        json.dump(visited_routes, f)


# Ignore weird headers
BLACKLIST = ["referrer", "host", "x-csrf-token", "origin"]
# Header must be sent at least 1/5 of the time for us to use as default
THRESHOLD = 5


def get_default_headers():
    har_file = json.load(open(HAR_CORPUS))
    counters = {
        "get": collections.Counter(),
        "post": collections.Counter(),
        "put": collections.Counter(),
        "delete": collections.Counter(),
        "patch": collections.Counter(),
    }
    for entry in har_file["log"]["entries"]:
        headers = har.get_har_headers(entry)
        verb = entry["request"]["method"].lower()
        # Keep track of number of requests with this verb
        counters[verb]["num_requests"] += 1
        # Keep track of how many times each header occurs
        for k, v in headers.items():
            # Ignore some weird headers
            if k.lower() not in BLACKLIST:
                counters[verb][k + "|" + v] += 1
    default_headers = {"get": {}, "post": {}, "put": {}, "delete": {}, "patch": {}}
    for verb in counters:
        # Get total number of requests sent with this verb
        total_num_requests = counters[verb]["num_requests"]
        # Don't want to confuse this key with other headers
        del counters[verb]["num_requests"]
        # Header must be sent at least 1/5 of the time for us to use as default
        min_requests = total_num_requests / THRESHOLD
        for header, count in counters[verb].items():
            # This header meets the threshold
            if count > min_requests:
                name, val = header.split("|")
                default_headers[verb][name] = val
    return default_headers


def preprocess(results_path, port):
    """
    Replay har requests and discover params/types
    """
    try:
        os.remove(os.path.join(results_path, "params"))
    except FileNotFoundError:
        pass
    har_file = json.load(open(HAR_CORPUS))

    # Send HAR request
    replayer = har.HarReplayer("localhost:" + str(port))
    visited_routes = []
    processed_routes = load_routes()
    for entry in har_file["log"]["entries"]:
        status_code = replayer.replay_har(entry)
        if status_code is None:
            # Route timed out
            continue
        url = entry["request"]["url"]
        verb = entry["request"]["method"]
        parsed_url = urlparse.urlparse(url)
        query_params = urlparse.parse_qs(parsed_url.query)
        query_params = [{"name": k, "values": vs} for k, vs in query_params.items()]

        # Read body params from har
        body_params = [
            {"name": k, "values": [v]} for k, v in har.get_body_params(entry).items()
        ]
        # Figure out what route was hit using path
        patched_url = har.patch_url(entry, "localhost:" + str(port))
        canonicalized_route, dynamic_segments = canonicalize_route(
            patched_url, "localhost:" + str(port), verb, processed_routes
        )
        # This just gets static content
        if canonicalized_route is None:
            continue

        # Each segment should be a list of values
        dynamic_segments = [
            {"name": k, "values": [v]} for k, v in dynamic_segments.items()
        ]

        # Make canonicalized route a dict
        canonicalized_route = {
            "url": entry["request"]["url"],
            "path": canonicalized_route,
            "query_params": query_params,
            "body_params": body_params,
            "dynamic_segments": dynamic_segments,
            "verb": entry["request"]["method"],
            "browser_status_code": entry["response"]["status"],
            "har_status_code": status_code,
            "headers": har.get_har_headers(entry),
        }

        # Check if it's in our list of visited routes
        index = find_in_routes(
            visited_routes, canonicalized_route["path"], canonicalized_route["verb"]
        )

        # We already saw this route
        if index is not None:
            # Grab the current list of params and update
            update_params(visited_routes[index]["body_params"], body_params)
            update_params(visited_routes[index]["query_params"], query_params)
            update_params(visited_routes[index]["dynamic_segments"], dynamic_segments)
        # This is a new entry
        else:
            visited_routes.append(canonicalized_route)

    # in place type inference
    infer_routes(visited_routes, results_path, "parsed_har_requests.json")


def run_parser():
    parser = argparse.ArgumentParser()
    parser.add_argument("results_path", help="results path")
    parser.add_argument("port", help="port")
    args = parser.parse_args()
    return args


def main():
    args = run_parser()
    preprocess(args.results_path, args.port)


if __name__ == "__main__":
    main()
