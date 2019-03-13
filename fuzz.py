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

# Max attempts allowed for querying an endpoint
# after all parameters have been mutated
MAX_FAILURES = 10

# Max number of parameters before power set explosion and we are killed by OOM
MAX_PARAMS = 20

HAR_DUMP = "preprocess/visited_routes.json"
# Logger for general debugging
logger = logging.getLogger("debug")

RESULTS_PATH = "/tmp/results"
PORT = 8080
FUZZ_DB = "FUZZ_DB"


def init_logger(quiet=None):
    global logger
    # Write everything to stdout
    ch = logging.StreamHandler()
    logger.addHandler(ch)
    # Log to a file as well
    fh = logging.FileHandler(os.path.join(RESULTS_PATH, "client.stdout"))
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
    har = True
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
    conn,
    target,
    state,
    target_route=None,
    stop_after_har=True,
    stop_after_all_routes=False,
    # Should we take snapshots at all?
    should_snapshot=True,
    # Should we swallow exceptions in the fuzzer?
    debug_mode=True,
    # Should we keep snapshots lying around or do gc?
    keep_snapshot=False,
):
    mutator = get_mutator(target)

    stats = fuzz_stats.FuzzStats()

    last_route = None
    while True:
        route = mutator.next_route()
        if route is None:
            break
        elif target_route is not None and not route.matches(target_route):
            continue

        state_dir = get_snapshot_name(target, state, route)
        if last_route != route:
            print("\n\n\n***%s %s***" % (route.verb, route.path))
            if should_snapshot:
                state.save(state_dir)
                print(
                    "State saved at %s with %d cookies"
                    % (state_dir, len(state.cookies))
                )
        last_route = route

        try:
            status_code = conn.send_request(
                route.url(target.port),
                route.verb,
                body_params=route.get_body_params(),
                query_params=route.get_query_params(),
                headers=route.headers,
            )
            exceptions = target.rails_exceptions.update()
            keep_snapshot = keep_snapshot or len(exceptions) > 0
            stats.record_stats(route.verb, route.path, status_code, exceptions)
            mutator.on_response(target, status_code)
        # We sent ctl-c, exit now
        except KeyboardInterrupt:
            exit(1)
        # Our fuzzer raised an exception
        except Exception as e:
            if debug_mode:
                raise e
            # Skip this route and pick another one
            mutator.force_next_route()
            target.on_fuzz_exception(route)

        percentage = coverage.calculate_coverage_percentage(
            target.cov.cumulative_coverage
        )
        stats.record_coverage(route.verb, route.path, percentage)
        print("\n\tcumulative cov: %f" % percentage)
        stats.save(target.results_path)
        if should_snapshot and not keep_snapshot:
            state.delete(state_dir)

    counts = json.dumps(stats.get_code_counts(), sort_keys=True)
    print("Code Counts: {}".format(counts))
    print("Final Coverage: {}".format(stats.final_coverage()))
    print("Success Ratio: {}".format(stats.get_success_ratio()))
    print("Total requests: {}".format(len(stats.get_results())))


def fuzz(
    snapshot=None,
    route=None,
    load_db=False,
    route_prefix=None,
    any_route=None,
    stop_after_har=False,
    stop_after_all_routes=False,
):
    random.seed(a=0)
    init_logger()

    pg = postgres2.Postgres()
    state = fuzz_state.FuzzState(pg, FUZZ_DB)

    if snapshot:
        clear_rails_connections(hostname=netutils.target_hostname(), port=PORT)
        logger.info("Loading all state from %s" % snapshot)
        state.load(snapshot)

    # open a connection with the server (need this to keep track of cookies)
    conn = netutils.Connection(state.cookies)
    conn.is_alive()

    # TODO: Get rid of this or move it to postgres2
    postgres.connect_to_db(FUZZ_DB)

    # Wait until server is up then read pluralizations dumped
    init_pluralization()

    target = fuzz_target.Target(RESULTS_PATH, PORT, FUZZ_DB, snapshot=snapshot)

    run(
        conn,
        target,
        state,
        target_route=route,
        stop_after_har=stop_after_har,
        stop_after_all_routes=stop_after_all_routes,
    )


def run_parser():
    parser = argparse.ArgumentParser(description="Fuzzing client")
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
        snapshot=args.snapshot,
        route=args.route,
        load_db=args.load_db,
        any_route=args.any_route,
        stop_after_har=args.stop_after_har,
        stop_after_all_routes=args.stop_after_all_routes,
    )


if __name__ == "__main__":
    main()
