#! /usr/local/bin/python3

import collections
import json

import fuzzer.params as params
import fuzzer.preprocess.preprocess as preprocess
import fuzzer.netutils as netutils


# Check if a route dict is present in a list of route objs
def find_route(route, routes):
    for r in routes:
        if route["verb"] == r.verb and route["path"] == r.path:
            return r
    return None


class Route(object):
    def __init__(
        self,
        path,
        verb,
        browser_status_code=None,
        har_status_code=None,
        headers=None,
        query_params=None,
        body_params=None,
        dynamic_segments=None,
    ):
        self.path = path
        self.verb = verb
        self.browser_status_code = browser_status_code
        self.har_status_code = har_status_code
        self.headers = headers
        self.query_params = query_params or []
        self.body_params = body_params or []
        self.dynamic_segments = dynamic_segments or []
        # List of list of queries sent with every request
        self.queries = [[]]
        self.unique_queries = []

    # Create a Route object from a har
    @classmethod
    def from_har(cls, route_dict):
        query_params = [params.Param.from_dict(p) for p in route_dict["query_params"]]
        body_params = [params.Param.from_dict(p) for p in route_dict["body_params"]]
        dynamic_segments = [
            params.Param.from_dict(p) for p in route_dict["dynamic_segments"]
        ]
        return cls(
            route_dict["path"],
            route_dict["verb"],
            browser_status_code=route_dict["browser_status_code"],
            har_status_code=route_dict["har_status_code"],
            headers=route_dict["headers"],
            query_params=query_params,
            body_params=body_params,
            dynamic_segments=dynamic_segments,
        )

    # Create a Route object from a dict
    @classmethod
    def from_dict(cls, route_dict):
        # TODO: Dynamic segments
        dynamic_segments = [params.Param(p) for p in route_dict["segments"]]
        return cls(
            route_dict["path"], route_dict["verb"], dynamic_segments=dynamic_segments
        )

    def unique_id(self):
        return self.escape_filename()

    # Preprocessed har requests
    def from_har_file(file):
        routes = json.loads(open(file, "r").read())
        return [Route.from_har(r) for r in routes]

    # Given rails dump of endpoints, convert to route objs and merge with
    # route objs from har dump
    def from_routes_file(file, har_routes=None):
        # Grab default headers from preprocessing har
        default_headers = preprocess.get_default_headers()
        routes = json.loads(open(file, "r").read())
        all_routes = []
        for route in routes:
            matched_route = find_route(route, har_routes) if har_routes else None
            if matched_route:
                all_routes.append(matched_route)
            else:
                r = Route.from_dict(route)
                r.headers = default_headers[r.verb.lower()]
                all_routes.append(r)

        return all_routes

    def escape_filename(self):
        filename = self.verb + self.path
        return filename.replace("/", "_")

    def url(self, port):
        path = self.path
        # Populate dynamic segments
        for segment in self.dynamic_segments:
            path = path.replace(":" + segment.name, str(segment.next_val))
        return "http://{}:{}{}".format(netutils.target_hostname(), str(port), path)

    def get_body_params(self):
        return {p.name: p.next_val for p in self.body_params if p.next_val is not None}

    def get_query_params(self):
        return {p.name: p.next_val for p in self.query_params if p.next_val is not None}

    def matches(self, route_str):
        verb, path = route_str.split(":", 1)
        return self.path.lower() == path.lower() and self.verb.lower() == verb.lower()
