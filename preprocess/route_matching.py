#! /bin/python

import re


# Takes in an array of routes such as: `["/a/:b/c/:d", "/x/y", ...]`.
def process_routes(routes):
    processed_routes = []
    for route in routes:
        params = re.findall("/:([^/]+)", route)

        route_regexp = route
        for param in params:
            route_regexp = route_regexp.replace("/:" + param, "/([^/]+)")
        route_regexp += "$"

        processed_routes.append(
            {"path": route, "regexp": route_regexp, "params": params}
        )
    return processed_routes


# Path is a path like "/a/1/c/2", routes is the output of `process_routes`
def find_route(routes, path, verb):
    path = path.split("?")[0]

    candidates = []
    for route in routes:
        route_regexp = route.get("regexp")

        if route.get("verb", "GET") != verb or re.match(route_regexp, path) is None:
            continue

        param_vals = list(re.match(route_regexp, path).groups())
        assert len(param_vals) == len(route.get("params"))

        param_dict = {}
        for idx, param in enumerate(route.get("params")):
            param_dict[param] = param_vals[idx]

        candidates.append((route.get("path"), param_dict))

    if len(candidates) == 0:
        return "", {}

    best = candidates[0]
    for candidate in candidates:
        if len(best[1]) > len(candidate[1]):
            best = candidate

    return best[0], best[1]
