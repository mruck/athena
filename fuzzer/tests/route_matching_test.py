#! /usr/bin/python3

# /a/b/x1/c/x2/d
# /a/b/([^/]+)/c/([^/]+)/d
# ==> match, [x1, x2]

# /a/b/:p1/c/:p2/d
# /:([^/]+)
# ==> match, [p1, p2]

import fuzzer.preprocess.route_matching as route_matching

routes = ["/a/b/:p1/c", "/d/:p1/e/:p2", "/d/:p1/e/f", "/a/b", "/a/b/c"]

table = [
    {"path": "/a/b/x1/c", "route": "/a/b/:p1/c", "params": {"p1": "x1"}},
    {"path": "/d/x1/e/x2", "route": "/d/:p1/e/:p2", "params": {"p1": "x1", "p2": "x2"}},
    {"path": "/d/x1/e/f", "route": "/d/:p1/e/f", "params": {"p1": "x1"}},
    {"path": "/d/e", "route": "", "params": {}},
    {"path": "/a", "route": "", "params": {}},
    {"path": "/a/b", "route": "/a/b", "params": {}},
    {"path": "/a/b/c", "route": "/a/b/c", "params": {}},
]


processed_routes = route_matching.process_routes(routes)
for test in table:
    route, params = route_matching.find_route(processed_routes, test["path"], "GET")
    assert route == test["route"], route
    assert params == test["params"], params
print("PASSED")
