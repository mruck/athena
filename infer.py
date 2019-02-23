#! /usr/local/bin/python3
import json
import random
import time

import fuzzer.util as util

PARAM_TYPE_UNKNOWN = "unknown"
PARAM_TYPE_STRING = "string"
PARAM_TYPE_TIMESTAMP = "ts"
PARAM_TYPE_BOOL = "boolean"
PARAM_TYPE_INT = "int"
PARAM_TYPE_FLOAT = "float"


def gen_random_param(param_type=PARAM_TYPE_UNKNOWN):
    if param_type == PARAM_TYPE_UNKNOWN or param_type == PARAM_TYPE_STRING:
        return util.random_str()
    elif param_type == PARAM_TYPE_INT:
        return util.random_int()
    elif param_type == PARAM_TYPE_FLOAT:
        return random.uniform(0.0, 100.0)
    elif param_type == PARAM_TYPE_BOOL:
        return True if random.uniform(0, 1) > 0.5 else False
    elif param_type == PARAM_TYPE_TIMESTAMP:
        return random.uniform(time.time() - 100, time.time())
    return util.random_str()


def read_file(filepath):
    output = {}
    fh = open(filepath, "r")
    contents = fh.read().split("\n")

    for line in contents:
        if line == "":
            continue
        obj = json.loads(line)
        print(obj)

        for key in obj:
            output[key] = {"types": [], "values": list(set(obj[key]))}
    return output


def maybe_bool(vals):
    for val in vals:
        if val != "true" and val != "false":
            return False
    return True


def maybe_int(vals):
    for val in vals:
        try:
            int(val)
        except ValueError:
            return False
    return True


def maybe_ts(vals):
    for val in vals:
        if len(val) != 13:
            return False

        try:
            int(val)
        except ValueError:
            return False

    return True


def maybe_float(vals):
    for val in vals:
        try:
            float(val)
        except ValueError:
            return False

    return True


def infer(param):
    """
    In place type inference.  Side effect: filters the "values"
    """
    param["values"] = list(set(param["values"]))
    param["values"] = [x for x in param["values"] if x is not None]
    param["types"] = []
    if len(param["values"]) == 0 or param["values"] == [None]:
        param["types"] = [PARAM_TYPE_UNKNOWN]
        return

    if maybe_bool(param["values"]):
        param["types"].append(PARAM_TYPE_BOOL)

    if maybe_int(param["values"]):
        param["types"].append(PARAM_TYPE_INT)
    elif maybe_float(param["values"]):
        param["types"].append(PARAM_TYPE_FLOAT)

    if maybe_ts(param["values"]):
        param["types"].append(PARAM_TYPE_TIMESTAMP)

    if len(param["types"]) == 0:
        param["types"] = [PARAM_TYPE_STRING]


dump = "/tmp/results_01_29_19_10_25/parsed_har_requests.json"


def analyze():
    routes = json.loads(open(dump, "r").read())
    for route in routes:
        for param in route["params"]:
            infer(param)
    print(json.dumps(routes))


if __name__ == "__main__":
    analyze()
