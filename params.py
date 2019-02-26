#! /usr/local/bin/python3

import json

import fuzzer.database.db as db
import fuzzer.lib.util as util


class Param(object):
    def __init__(self, name, har_values=None, types=None):
        self.name = name
        # Values from har requests or None if we discovered this param in the fuzzer
        self.har_values = filter_har_values(har_values) if har_values else []
        # Types infered
        self.types = types or []
        # List of list of metadata queries run with this parameter for every request sent
        # (will have lots of empty lists assuming param isn't present
        # in all requests sent)
        self.query_metadata_list = []
        # Pass the original parameter the first time or nothing at all.
        # Sanity check to ensure the fuzzer with reasonable har_values is
        # well behaved
        self.next_val = self.har_values[0] if len(self.har_values) > 0 else None
        # All the values we've sent for this param
        self.prev_vals = [self.next_val]

    # Not all params have default har values.  So either get a har value
    # we haven't sent before or return None
    def get_har_value(self):
        # This param has no default har values
        if len(self.har_values) == 0:
            return None
        for har_val in self.har_values:
            # We've never sent this before
            if har_val not in self.prev_vals:
                return har_val
        # We've sent all har vals before, default to the first one
        return self.har_values[0]

    # Attempt a db lookup.  If table is empty, default to har values.
    # Ensure it's not a value we've sent before.
    def _db_lookup(self, table, col):
        # Count rows in the table
        rows = db.count_rows(table)
        # The table is empty
        if rows == 0:
            # Default to the har values
            return self.get_har_value()
        for i in range(0, rows):
            # Query db
            val = db.lookup(table, col)
            # We've never seen that val before
            if val not in self.prev_vals:
                return val

    def update_next_val(self, val=None, table=None, col=None):
        # Store the old val
        self.prev_vals.insert(0, self.next_val)
        # A specific value was provided
        if val:
            self.next_val = val
        # Do a table and col lookup
        else:
            self.next_val = self._db_lookup(table, col)
        return self.next_val

    @classmethod
    def from_dict(cls, param_dict):
        return cls(
            param_dict["name"],
            har_values=param_dict["values"],
            types=param_dict["types"],
        )

    # Filter out empty lists from self.queries
    def get_queries(self):
        # Check if this param was ever in any queries
        # Returns a list of list of queries
        return [q for q in self.queries if len(q) > 0]

    # Naive type-based mutation
    def mutate(self, respect_har=False):
        # Don't send params unless they were from a human browsing session
        if respect_har and self.next_val is None:
            return
        # TODO: add support to mutate param if mapped to a db val
        # Get an arbitrary value of this type
        if len(self.types) > 0:
            if self.types[0] == "int":
                next_val = util.random_int()
            elif self.types[0] == "string":
                next_val = util.random_str()
            elif self.types[0] == "float":
                next_val = util.random_str()
            elif self.types[0] == "boolean":
                next_val = self.get_har_value()
            elif self.types[0] == "ts":
                next_val = util.random_str()
        # Get an arbitrary value and type
        else:
            next_val = util.random_str()
        self.update_next_val(val=next_val)


def filter_har_values(values):
    if values is None:
        return []
    # Filter the empty string
    return [v for v in values if v != ""]


# [["post", "raw"], ["post"]] => ["post", "raw"]
def dedup_params(raw_params):
    full_dict = {}

    def update_dict(d, arr):
        curr = d
        for i in range(len(arr)):
            key = arr[i]
            if key not in curr:
                curr[key] = {}
            curr = curr[key]

    def leaf_paths(d):
        output = []
        for k, v in d.items():
            if len(v) == 0:
                output.append([k])
            else:
                child_output = leaf_paths(v)
                output.extend([[k] + o for o in child_output])
        return output

    for arr in raw_params:
        try:
            update_dict(full_dict, arr)
        # For some reason sometimes we dump the key as as a dict or
        # array. Skip this for now
        except TypeError:
            pass

    return leaf_paths(full_dict)


# Read params rails dumps and return a flattened list for every param
def read_params(params_file):
    raw_params = []
    for json_line in params_file:
        json_line = json_line.strip()
        param = json.loads(json_line)
        # We are already tracking this param
        if param in raw_params:
            continue
        raw_params.append(param)
    # Dedup params
    raw_params = dedup_params(raw_params)
    # Flatten params
    return [flatten_param(p) for p in raw_params]


# Check if new param is present in list of old params
# new param is a string and old params is list of Param objs.
# Returns true if we found param, else false
def find_param(new_param, old_params):
    return len([p for p in old_params if p.name == new_param]) > 0


# Given a set of params, see if we are already tracking them in Route
def get_params_delta(params, route):
    new_params = []
    for param in params:
        # Check if we already have this as a query param
        if find_param(param, route.query_params):
            continue
        # Check if we already have this as a dynamic segment
        if find_param(param, route.dynamic_segments):
            continue
        # Check if we already have this as a body param
        if find_param(param, route.body_params):
            continue
        # We didn't find it, this is a new param, append
        new_params.append(param)
    return new_params


# Given new params, add them with the existing set of params
def merge_params(new_params, old_params):
    for new_param in new_params:
        old_params.append(Param(new_param))


def flatten_param(param):
    # Singleton list, i.e. ["mobile"] => "mobile"
    if len(param) == 1:
        return param[0]
    # Nested dictionary or array, i.e. ["post", "raw"] => "post[raw]"
    else:
        flattened = param[0]
        # build dict
        for i in range(1, len(param)):
            flattened += "[{}]".format(param[i])
        return flattened


# Given a list of params, return whether or not any new params were discovered.
# A param is "new" if param.next_val is None, this means we have not yet sent it
# in a request
def params_delta(params):
    for param in params:
        if param.next_val is None:
            return True
    return False
