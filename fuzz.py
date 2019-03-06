#!/usr/bin/python3

# Main fuzz routine.  Iterates through each route, identifying the required
# parameters and then mutating them so that progress is made (ie db queries
# are successful)
#
# Note: This is a library never meant to be invoked directly, only
# from process.py as that does STATE validation, directory setup, etc.

import argparse
import json
import logging
import os
import random

from fuzzer.database.db import init_pluralization, clear_rails_connections
import fuzzer.mutation.naive_mutator as naive_mutator
import fuzzer.routes as routes_lib
import fuzzer.lib.coverage as coverage
import fuzzer.database.postgres as postgres
import fuzzer.fuzz_target as fuzz_target
import fuzzer.database.postgres2 as postgres2
import fuzzer.fuzz_state as fuzz_state
import fuzzer.fuzz_stats as fuzz_stats
import fuzzer.lib.netutils as netutils

# DB dump, cookie, routes.json, pluralizations and any other app specific
# state should be stored here
STATE = "/state"

# Max attempts allowed for querying an endpoint
# after all parameters have been mutated
MAX_FAILURES = 10

# Max number of parameters before power set explosion and we are killed by OOM
MAX_PARAMS = 20

HAR_DUMP = "preprocess/visited_routes.json"
# Logger for general debugging
logger = logging.getLogger("debug")


def init_logger(results_path, quiet=None):
    global logger
    # Write everything to stdout
    ch = logging.StreamHandler()
    logger.addHandler(ch)
    # Log to a file as well
    fh = logging.FileHandler(os.path.join(results_path, "client.stdout"))
    fh.setLevel(logging.DEBUG)
    logger.addHandler(fh)
    if quiet is None:
        logger.setLevel(logging.DEBUG)
    elif quiet == 0:
        logger.setLevel(logging.ERROR)
    else:
        assert False


def get_snapshot_name(target, state, route):
    uid = "{}.{}".format(route.unique_id(), random.randint(0, 10000))
    return os.path.join(target.results_path, "snapshots", uid)


def get_mutator(target):
    har = False
    all_routes = routes_lib.read_routes(
        os.path.join(target.results_path, "routes.json")
    )
    # We don't have a har to drive mutation
    if not har:
        return naive_mutator.NaiveMutator(all_routes)
    # read in routes dumped by preprocessor
    har_routes = routes_lib.Route.from_har_file(HAR_DUMP)
    routes_lib.merge_with_har(all_routes, har_routes)
    return naive_mutator.HarMutator(
        har_routes, all_routes, stop_after_har=True, stop_after_all_routes=True
    )


def run(
    target, state, target_route=None, stop_after_har=False, stop_after_all_routes=False
):
    mutator = get_mutator(target)

    # open a connection with the server (need this to keep track of cookies)
    conn = netutils.Connection(state.cookies)

    stats = fuzz_stats.FuzzStats()

    last_route = None
    state_dir = None
    while True:
        route = mutator.next_route()
        if route is None:
            break
        elif target_route is not None and not route.matches(target_route):
            continue

        state_dir = get_snapshot_name(target, state, route)
        if last_route != route:
            print("\n\n\n***%s %s***" % (route.verb, route.path))
            state.save(state_dir)
            print("State saved at %s with %d cookies" % (state_dir, len(state.cookies)))
        last_route = route

        keep_snapshot = False
        try:
            status_code = conn.send_request(
                route.url(target.port),
                route.verb,
                body_params=route.get_body_params(),
                query_params=route.get_query_params(),
                headers=route.headers,
            )
            exceptions = target.latest_exns()
            keep_snapshot = keep_snapshot or len(exceptions) > 0
            stats.record_stats(
                route.verb, route.path, status_code, exceptions, state_dir
            )
            mutator.on_response(target, status_code)
        # Our fuzzer raised an exception
        except:
            # Skip this route and pick another one
            mutator.force_next_route()
            target.on_fuzz_exception(route, state_dir)
            keep_snapshot = True

        percentage = coverage.calculate_coverage_percentage(
            target.cov.cumulative_coverage
        )
        stats.record_coverage(route.verb, route.path, percentage)
        print("\n\tcumulative cov: %f" % percentage)
        stats.save(target.results_path)

        if not keep_snapshot:
            state.delete(state_dir)

    counts = json.dumps(stats.get_code_counts(), sort_keys=True)
    print("Code Counts: {}".format(counts))
    print("Final Coverage: {}".format(stats.final_coverage()))
    print("Success Ratio: {}".format(stats.get_success_ratio()))
    print("Total requests: {}".format(len(stats.get_results())))


def fuzz(
    fuzz_dir,
    db,
    port,
    snapshot=None,
    route=None,
    load_db=False,
    route_prefix=None,
    any_route=None,
    stop_after_har=False,
    stop_after_all_routes=False,
):
    random.seed(a=0)
    init_logger(fuzz_dir)
    init_pluralization(STATE)

    pg = postgres2.Postgres()
    state = fuzz_state.FuzzState(pg, db)
    if snapshot:
        clear_rails_connections(hostname=netutils.target_hostname(), port=port)
        logger.info("Loading all state from %s" % snapshot)
        state.load(snapshot)

    # TODO: Get rid of this or move it to postgres2
    postgres.connect_to_db(db)

    target = fuzz_target.Target(fuzz_dir, port, db, snapshot=snapshot)

    run(
        target,
        state,
        target_route=route,
        stop_after_har=stop_after_har,
        stop_after_all_routes=stop_after_all_routes,
    )


def run_parser():
    parser = argparse.ArgumentParser(description="Fuzzing client")
    parser.add_argument("fuzz_dir", help="Destination for results")
    parser.add_argument("db", help="db to connect to")
    parser.add_argument("port", help="port to query")
    parser.add_argument("--route")
    parser.add_argument("--snapshot")
    parser.add_argument("--load_db", action="store_true")
    parser.add_argument("--any-route", action="store_true")
    parser.add_argument("--stop_after_har", action="store_true")
    parser.add_argument("--stop_after_all_routes", action="store_true")
    args = parser.parse_args()
    return args


def main():
    args = run_parser()
    fuzz(
        args.fuzz_dir,
        args.db,
        args.port,
        snapshot=args.snapshot,
        route=args.route,
        load_db=args.load_db,
        any_route=args.any_route,
        stop_after_har=args.stop_after_har,
        stop_after_all_routes=args.stop_after_all_routes,
    )


if __name__ == "__main__":
    main()
