# After sending a request, update the route object with new information aout
# src coverage, queries and params
import json
import os

import fuzzer.query as query_lib
import fuzzer.params as params_lib


# Read queries from rails dump and store in route obj
def update_queries(queries_file, route):
    new_queries = query_lib.update_queries(queries_file, route)
    route.queries.insert(0, new_queries)


# Read params from  rails dump and store in route obj
def update_params(params_file, route):
    # Read params rails accessed
    new_params = params_lib.read_params(params_file)
    # Given the new params, see if we are already tracking them in Route
    new_params = params_lib.get_params_delta(new_params, route)

    # There are no new params
    if len(new_params) == 0:
        return False
    # These are query params
    if route.verb == "GET":
        params_lib.merge_params(new_params, route.query_params)
    # These are body params
    else:
        params_lib.merge_params(new_params, route.body_params)
    # TODO: Add check for dynamic segments for when we hit routes that
    # aren't from hars
    return True


def check_for_sql_inj(target, route):
    params_sent = route.params_sent()
    queries = route.queries[0]
    for q in queries:
        vuln = q.is_vulnerable(params_sent)
        if vuln is None:
            continue
        # Our query showed up in a literal sql inj
        ast, param = vuln
        sql_dict = {"path": route.path, "verb": route.verb, "param": param, "ast": ast}
        with open(os.path.join(target.results_path, "sql_inj.json"), "a") as f:
            json.dump(sql_dict, f)


# Collect information relevant to the mutation decision.
# Specifically src cov deltas, query deltas and param deltas
def update_route_state(target, route):
    # read params file
    update_params(target.params_file, route)
    # read queries file
    update_queries(target.queries_file, route)
    check_for_sql_inj(target, route)
