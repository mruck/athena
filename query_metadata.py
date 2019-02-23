# Library for manipulating query metadata object, where a query metadata obj
# contains metadata for a given param (i.e. table and col that param map to in
# the query)
import copy

import query as query_lib


class QueryMetaData(object):
    # Contains generic info about a query, but also info specific to the param in the query
    def __init__(self, table, col, query_obj):
        self.table = table
        self.col = col
        self.query_obj = query_obj


# Return list of params that showed up in most recent set of queries
def check_for_new_queries(params):
    params_prime = []
    for p in params:
        # This param was present in the most recent queries
        if len(p.query_metadata_list[0]) > 0:
            params_prime.append(p)
    return params_prime


# Skip params that we shouldn't search query strings for
# TODO: skip stale params, new params, nesting parent
def skip_param(param):
    # No param was sent in past request so it wont be present in the query
    if param.next_val is None:
        return True
    # The param value is boolean and too generic to search
    if "boolean" in param.types:
        return True
    # The param is the empty string and too generic to search
    if param.next_val == "":
        return True
    return False


def find_param_in_query_wrapper(param_value, query):
    # We are going to mutate the AST and strip the param out of the
    # query so make a copy
    new_ast = copy.deepcopy(query.ast["where_clause"])
    table, col, ast, constraints = query_lib.find_param_in_query(
        param_value, query, new_ast
    )
    if table is None:
        return None
    else:
        return QueryMetaData(table, col, query)


# Update each param in place if it was present in queries in the most recent
# request
def search_queries_for_params(params, new_queries):
    for param in params:
        # list of [QueryMetaData]
        # Empty list indicates no queries for this param
        param.query_metadata_list.insert(0, [])
        # No queries were made with this parameter
        if skip_param(param):
            continue
        for q in new_queries:
            query_metadata = find_param_in_query_wrapper(param.next_val, q)
            # Didn't find the param
            if query_metadata is None:
                continue
            print(
                "\t(%s:%s) -> select %s from %s"
                % (
                    param.name,
                    str(param.next_val),
                    query_metadata.col.upper(),
                    query_metadata.table.upper(),
                )
            )
            print("{}\n".format(query_lib.stringify_query(q)))
            param.query_metadata_list[0].append(query_metadata)
