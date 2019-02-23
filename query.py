#!/bin/bash/python3
import json
import logging

from util import case_insensitive_contains

logger = logging.getLogger("debug")


# TODO: Keep a list of unique queries in the route obj
def queries_delta(new_queries, unique_queries):
    return False


# args:
#   param: param to search for
#   query: query with model information
#   ast: ast to recurse on
# returns:
#   (table, col, new_ast, constraints)
#   table: table to query
#   col: col that maps to parameter
#   ast: ast of new query to run (we use this to build sql query string)
#   constraints: constraints to apply to parameter (ie greater than)
#                in the form of a list of strings
def find_param_in_query(param, query, ast):
    if ast == {}:
        return (None, None, {}, [])
    elif ast["type"] == "in":
        values = [x for x in ast["values"] if case_insensitive_contains(param, x)]
        # The param we control showed up in the "in" clause. We have no constraints
        # i.e. age IN [?]
        if values:
            return (
                ast["attribute"]["table"],
                ast["attribute"]["attribute_name"],
                {},
                [],
            )
        # Accumulate this constraint
        # i.e. id IN [1,2,3]
        else:
            return (None, None, ast, [])
    elif ast["type"] == "equality":
        if case_insensitive_contains(param, ast["right"]["value"]):
            return (ast["left"]["table"], ast["left"]["attribute_name"], {}, [])
        else:
            return (None, None, ast, [])
    elif ast["type"] == "not_equal":
        if case_insensitive_contains(param, ast["right"]["value"]):
            # We have full control over equality check
            # i.e. age NOT IN [?]
            return (ast["left"]["table"], ast["left"]["attribute_name"], {}, [])
        else:
            # We have no control over equality check so append as a constraint
            # i.e. age NOT IN [5]
            return (None, None, ast, [])
    elif ast["type"] == "not in":
        values = [x for x in ast["values"] if case_insensitive_contains(param, x)]
        if values:
            # Remove our param from the list of values
            ast["values"] = [
                x for x in ast["values"] if not case_insensitive_contains(param, x)
            ]
            # Add this slightly modified ast as a constraint and note that our
            # param will also be checked for equality
            return (
                ast["attribute"]["table"],
                ast["attribute"]["attribute_name"],
                ast,
                ["not_equal"],
            )
        else:
            return (None, None, ast, [])
    elif ast["type"] == "literal":
        return (None, None, ast, [])
    elif ast["type"] == "parsed_literal_expression":
        # we found the parameter we passed in
        if case_insensitive_contains(param, ast["right"]):
            # grab the column
            col = ast["left"]
            table = query.model
            if ast["operator"] == "=":
                return (table, col, {}, [])
            else:
                # TODO: ast["operator"] can be other things besides "=", i.e. "<"
                # We don't handle this right now.
                return (None, None, {}, [])
                # return (table, col, {}, ast["operator"])
        else:
            return (None, None, ast, [])
    elif ast["type"] == "and" or ast["type"] == "or":
        (left_table, left_col, ast["left"], left_constraints) = find_param_in_query(
            param, query, ast["left"]
        )
        (right_table, right_col, ast["right"], right_constraints) = find_param_in_query(
            param, query, ast["right"]
        )
        table = left_table or right_table
        col = left_col or right_col
        # union constraints
        new_constraints = left_constraints + right_constraints
        # Return the entire AST
        if ast["left"] and ast["right"]:
            return (table, col, ast, new_constraints)
        # Either ast[left] or ast[right]
        else:
            ast = ast["left"] or ast["right"]
            return (table, col, ast, new_constraints)
    else:
        raise Exception("Unhandled AST: {}".format(ast))


# Takes in ast and returns stringified query to run
# Query is valid sql syntax, i.e. "= None" is translated to
# "IS NULL"
def build_query_string(query, ast):
    if ast == {}:
        return ""
    elif ast["type"] == "in" or ast["type"] == "not in":
        table = ast["attribute"]["table"]
        col = ast["attribute"]["attribute_name"]
        vals = tuple([x["value"] for x in ast["values"]])
        return "%s %s %s " % (col, ast["type"].upper(), vals)
    elif ast["type"] == "equality" or ast["type"] == "not_equal":
        table = ast["left"]["table"]
        col = ast["left"]["attribute_name"]
        val = ast["right"]["value"]
        operator = "=" if ast["type"] == "equality" else "<>"
        # TODO: This is a hacky way to tell if value is string/should
        # be in quotes.  Eventually support floats, other types?
        if isinstance(val, int):
            return "%s %s %s" % (col, operator, val)
        # Special case for IS NULL syntax
        elif val is None:
            if operator == "=":
                return "%s IS NULL" % (col)
            else:
                return "%s IS NOT NULL" % (col)
        else:
            return "%s %s '%s'" % (col, operator, val)
    elif ast["type"] == "literal":
        return ast["literal"]
    elif ast["type"] == "parsed_literal_expression":
        table = query.model
        col = ast["left"]
        val = ast["right"]
        operator = ast["operator"]
        return "%s %s '%s' " % (col, operator, val)
    elif ast["type"] == "and" or ast["type"] == "or":
        left_query = build_query_string(query, ast["left"])
        right_query = build_query_string(query, ast["right"])
        return "%s %s %s" % (left_query, ast["type"].upper(), right_query)
    else:
        raise Exception("Unhandled AST: {}".format(ast))


# Filter all requests for successful creates/updates
def log_creates(reqs, level=0):
    creates = []
    for req in reqs:
        for q in req.queries:
            if q.method == "create" or q.method == "update":
                if q.successful:
                    creates.append(q)
    if creates and level > 0:
        logger.info("previous creates/updates:")
        for q in creates:
            q.log(level)


def stringify_ast(ast, canonicalize=False):
    if "where_clause" in ast:
        ast = ast["where_clause"]
    if ast == {}:
        return ""
    elif ast["type"] == "equality" or ast["type"] == "not_equal":
        if canonicalize:
            val = "?"
        else:
            val = ast["right"]["value"]
        operand = "=" if ast["type"] == "equality" else "<>"
        ast_str = "%s[%s] %s %s" % (
            ast["left"]["table"],
            ast["left"]["attribute_name"],
            operand,
            val,
        )
        return ast_str
    elif ast["type"] == "parsed_literal_expression":
        return ast["literal"]
    elif ast["type"] == "literal":
        return ast["literal"]
    elif ast["type"] == "in" or ast["type"] == "not in":
        table = ast["attribute"]["table"]
        col = ast["attribute"]["attribute_name"]
        if canonicalize:
            values = "(?)"
        else:
            values = str([x["value"] for x in ast["values"]])
        ast_str = "%s[%s] %s %s" % (table, col, ast["type"], values)
        return ast_str
    elif ast["type"] == "and" or ast["type"] == "or":
        left = stringify_ast(ast["left"], canonicalize=canonicalize)
        conjunction = ast["type"]
        right = stringify_ast(ast["right"], canonicalize=canonicalize)
        ast_str = "%s %s %s" % (left, conjunction, right)
        return ast_str
    else:
        raise Exception("Unhandled AST: {}".format(ast))


def stringify_query(q, canonicalize=False):
    status = "successful" if q.successful else "failed"
    # This is a lookup
    if q.ast:
        ast = stringify_ast(q.ast["where_clause"], canonicalize=canonicalize)
        q_str = "\t%s %s: %s" % (status, q.method, ast)
    # There's no AST
    else:
        q_str = "\t%s %s for %s model\n" % (status, q.method, q.model)
        if not canonicalize:
            q_str += "\t\t%s" % q.record
    return q_str


def dedup_queries(queries):
    # Remove duplicate queries so that all queries in list are unique
    deduped_queries = []
    for q in queries:
        tmp = [x for x in deduped_queries if x.is_equal(q)]
        # We didn't see this query in our list
        if len(tmp) == 0:
            deduped_queries.append(q)
    return deduped_queries


def get_literal_expression_nodes(node):
    if "type" not in node:
        return []

    if node["type"] == "parsed_literal_expression":
        return [node]

    if node["type"] in ["and", "or"]:
        right = get_literal_expression_nodes(node["right"])
        left = get_literal_expression_nodes(node["left"])
        return right + left

    return []


class Query(object):
    def __init__(self, successful, method, model, ast, record):
        self.successful = successful
        self.method = method
        self.model = model
        self.ast = ast
        self.record = record

    def is_equal(self, query2):
        return (
            self.method == query2.method
            and self.model == query2.model
            and self.successful == query2.successful
            and stringify_ast(self.ast) == stringify_ast(query2.ast)
        )

    def is_vulnerable(self, param_values):
        if self.ast is None:
            return False

        literal_expression_nodes = get_literal_expression_nodes(
            self.ast["where_clause"]
        )

        vulns = []
        for node in literal_expression_nodes:
            for value in param_values:
                if str(value).upper() in node["literal"].upper():
                    print("Found param %s in query:" % value)
                    print(stringify_ast(self.ast))
                    exit(1)
                    vulns.append(value)

        return vulns

    # Args:
    #   query_dict: query dictionary
    # Returns:
    #   query object
    @classmethod
    def from_queries_file(cls, query_dict):
        successful = query_dict["successful"]
        method = query_dict["method"]
        model = query_dict["model"]
        record = query_dict["results"] if "results" in query_dict else None
        ast = query_dict["query"] if "query" in query_dict else None
        return cls(successful, method, model, ast, record)

    def log(self, level=0):
        # Only log failed models
        if level == 0:
            if not self.successful:
                logger.info("Failed model: {}".format(self.model))
        # Log all models + stringified AST
        if level == 1:
            self.log_metadata()
        # Log all models, stringified AST, and jsonified AST
        if level == 2:
            self.log_metadata()
            if self.ast is not None:
                logger.info(self.ast["where_clause"])

    def log_metadata(self):
        logger.info("***query***")
        logger.info("model: {}".format(self.model))
        logger.info("successful: {}".format(self.successful))
        logger.info("method: {}".format(self.method))
        if self.ast is not None:
            ast_str = stringify_ast(self.ast["where_clause"])
            logger.info(ast_str)


# Args:
#   Route
# Returns:
#   True if all queries succeeded else false
def queries_succeeded(route):
    if len(route.most_recent_queries) == 0:
        return route.attempts > 3
    for q in route.most_recent_queries:
        if not q.successful:
            return False
    return True


# Compare 2 lists of queries
def queries_are_equal(query_list1, query_list2):
    if len(query_list1) != len(query_list2):
        return False
    for i in range(0, len(query_list1)):
        if not query_list1[i].is_equal(query_list2[i]):
            return False
    return True


# Args:
#   route: updates route["queries"] if new queries have been made
# Returns: True if queries have changed, otherwise False
def read_queries(queries_file):
    new_queries = []
    for q in queries_file:
        q = json.loads(q)
        q = Query.from_queries_file(q)
        new_queries.append(q)
    return new_queries


# Read queries from json and store in route obj
def update_queries(queries_file, route):
    # read queries
    new_queries = read_queries(queries_file)
    # filter queries out that don't have ast
    new_queries = [q for q in new_queries if q.ast and "where_clause" in q.ast]
    # filter out duplicate queries
    return dedup_queries(new_queries)
