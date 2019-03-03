#! /usr/local/bin/python3

import json
import os

import fuzzer.params as params
import fuzzer.preprocess.preprocess as preprocess
import fuzzer.lib.netutils as netutils

STATE = "/state"
DEFAULT_ROUTE_EXCLUDES = [
    # Makes app RO
    "/admin/backups/readonly",
    # There's a bug in this route so don't hit
    "/admin/site_settings/:id",
    # Don't want to log out
    "logout",
    # This drops all db connections
    "clear_all_connections",
]
ROUTES_DUMP = os.path.join(STATE, "routes.json")


# Check if a route obj is present in a list of route objs
def find_route(needle, routes):
    for r in routes:
        if needle.verb == r.verb and needle.path == r.path:
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
    def from_har_file(har_file):
        routes = json.loads(open(har_file, "r").read())
        return [Route.from_har(r) for r in routes]

    def default_headers():
        return {"get": "", "put": "", "post": "", "patch": "", "delete": ""}

    # Given rails dump of endpoints, conv    ert to route objs and merge with
    # route objs from har dump
    def from_routes_file(routes_file):
        default_headers = preprocess.get_default_headers()
        all_routes = []
        fp = open(routes_file, "r")
        for json_line in fp:
            json_line = json_line.strip()
            route_dict = json.loads(json_line)
            r = Route.from_dict(route_dict)
            # Not sure why but some routes dont have verbs
            if r.verb == "":
                continue
            # Some routes have verb GET|POST
            if r.verb == "GET|POST":
                continue
            r.headers = default_headers[r.verb.lower()]
            all_routes.append(r)
        return all_routes

    # Merge har routes objs with all routes objs
    def merge_with_har(all_routes, har_routes):
        pass

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

    # Get values of params sent in most recent request
    # NOTE: assumes params have not yet been mutated!!!
    def params_sent(self):
        dynamic_segments = [
            p.next_val for p in self.dynamic_segments if p.next_val is not None
        ]
        body_params = [p.next_val for p in self.body_params if p.next_val is not None]
        query_params = [p.next_val for p in self.query_params if p.next_val is not None]
        return dynamic_segments + body_params + query_params

    def matches(self, route_str):
        verb, path = route_str.split(":", 1)
        return self.path.lower() == path.lower() and self.verb.lower() == verb.lower()


def filter_routes(routes, blacklist):
    return [r for r in routes if r.path not in blacklist]


# Given a list of route objs, order such that routes that create content run first
def order_routes(routes):
    ordering = ["post", "put", "patch", "get", "delete"]
    ordered = []
    for order in ordering:
        for r in routes:
            if r.verb.lower() == order:
                ordered.append(r)
    return ordered


def read_routes(routes_file):
    # read in routes dumped by rails
    all_routes = Route.from_routes_file(routes_file)
    filtered = filter_routes(all_routes, DEFAULT_ROUTE_EXCLUDES)
    ordered = order_routes(filtered)
    return ordered
