#!/bin/bash/python3

import datetime
import itertools
import logging
import json
import os
from pathlib import Path
import pprint
import random
import string
import time
import uuid
import urllib

# Keep track of how many times we've called rand
# so we can repro rand's behavior when loading a snapshot.
# TODO: eventually this will overflow
rand_counter = 0

logger = logging.getLogger("debug")


def target_app_port():
    port = os.getenv("TARGET_APP_PORT")
    if port is None:
        print("Error: TARGET_APP_PORT not specified")
        exit(1)
    return int(port)


def target_db_port():
    port = os.getenv("TARGET_DB_PORT")
    if port is None:
        print("Error: TARGET_DB_PORT not specified")
        exit(1)
    return port


def target_dbname():
    db = os.getenv("TARGET_DBNAME")
    if db is None:
        print("Error: TARGET_DBNAME not specified")
        exit(1)
    return db


def target_db_host():
    host = os.getenv("TARGET_DB_HOST")
    if host is None:
        print("Error: TARGET_DB_HOST not specified")
        exit(1)
    return host


def target_db_user():
    user = os.getenv("TARGET_DB_USER")
    if user is None:
        print("Error: TARGET_DB_USER not specified")
        exit(1)
    return user


def target_db_password():
    password = os.getenv("TARGET_DB_PASSWORD")
    if password is None:
        print("Error: TARGET_DB_PASSWORD not specified")
        exit(1)
    return password


def get_target_id():
    target_id = os.getenv("TARGET_ID")
    # Use a dummy id
    if target_id is None:
        return uuid.uuid4().hex
    else:
        return target_id


# Record state of the pseudo rng before hitting a route. This allows us
# to control for randomness.
def store_random_state(results_path, route):
    rand_state = random.getstate()
    rand_dict = {"verb": route.verb, "path": route.path, "rand_state": rand_state}
    with open(os.path.join(results_path, "random_store.json"), "a") as f:
        f.write(json.dumps(rand_dict) + "\n")


def load_random_state(rng_file, verb, path):
    rand_state_file = open(rng_file, "r")
    print("Loading pseudo rng state from %s" % rng_file)
    for line in rand_state_file:
        data = json.loads(line.strip())
        if (
            data["verb"].upper() == verb.upper()
            and data["path"].lower() == path.lower()
        ):
            state = data["rand_state"]
            # Format state into tuples cause that's what random wants and
            # json dumping messes this up
            # TODO: Not sure if I should be recursing into lists and converting
            # to tuples
            rand_state = []
            for elem in state:
                if isinstance(elem, list):
                    elem = tuple(elem)
                rand_state.append(elem)
            rand_state = tuple(rand_state)
            random.setstate(rand_state)
            return


def get_counter():
    return rand_counter


# Remove all files from results_path except routes.json
def clean_dir(results_path):
    pass


def random_str(length=10):
    seq = string.ascii_uppercase + string.digits
    mystr = "".join(random.choice(seq) for _ in range(length))
    return "BIP_" + mystr


def random_int(max=1000000):
    return random.randint(1, max)


def all_permutations_for_options(options):
    return [list(option) for option in itertools.product(*options)]


def num_permutations_for_options(options):
    """
    Where `options` is a list of lists of options, returns the number of possible permutations
    for those options.

    e.g., [(1, 2), (3, 4), (5, 6, 7)] => 2 * 2 * 3 == 12
    """

    return len(all_permutations_for_options(options))


def next_permutation_for_permutation_with_options(permutation_and_options):
    """
    Given a list of 2-tuples, where each 2-tuple is a (state, list of possible states), return
    a list of the states in the next permutation, or None if there is no next permutation.
    """

    all_permutations = all_permutations_for_options(
        [options for (_, options) in permutation_and_options]
    )

    current_permutation = [state for (state, _) in permutation_and_options]
    # Index into the list of all permutations with the current state
    current_permutation_index = all_permutations.index(current_permutation)

    # The current permutation is the last permutation
    if current_permutation_index + 1 >= len(all_permutations):
        return None
    # Return the next permutation
    else:
        return all_permutations[current_permutation_index + 1]


def is_last_permutation_for_permutation_with_options(permutation_with_options):
    return (
        next_permutation_for_permutation_with_options(permutation_with_options) is None
    )


def current_permutation_number_for_permutation_with_options(permutation_and_options):
    """
    Given a list of 2-tuples, where each 2-tuple is a (state, list of possible states), returns
    the number of the current permutation.
    """

    all_permutations = all_permutations_for_options(
        [options for (_, options) in permutation_and_options]
    )
    current_permutation = [state for (state, _) in permutation_and_options]

    return all_permutations.index(current_permutation) + 1


def run_startup_checks():
    pass
    # Make sure Elasticsearch is running.
    # try:
    #    requests.get("http://localhost:9200")
    # except requests.exceptions.ConnectionError:
    #    print("Elasticsearch isn't running!")
    #    print("\t sudo service elasticsearch start")
    #    sys.exit(1)


def parse_status_codes(routes):
    global logger
    status_codes = {}
    for r in routes:
        code = r.requests[0].status_code
        if code in status_codes:
            status_codes[code].append(r)
        else:
            status_codes[code] = [r]
    logger.info("\n#########################################################\n")
    logger.info("Status code results:\n")
    for code, routes in sorted(status_codes.items()):
        logger.info("Status code: {}".format(code))
        for r in routes:
            logger.info("\t{} {}".format(r.verb, r.path))


def get_results_path():
    return os.getenv("RESULTS_PATH")


def case_insensitive_contains(needle, haystack):
    return str(haystack).upper().find(str(needle).upper()) >= 0


def print_dict(d):
    pprint.PrettyPrinter(indent=2).pprint(d)


# Query dispatcher for route
def dispatch():
    try:
        r = urllib.request.urlopen("http://localhost:8080").read()
        return eval(r.decode("utf-8"))
    # No routes left!
    except urllib.error.HTTPError:
        return None


def get_open_port():
    """
    Returns an open port on the local machine.
    """
    import socket

    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    s.bind(("", 0))
    s.listen(1)
    port = s.getsockname()[1]
    s.close()
    return port


def mk_results_path():
    results_path = "/tmp/results_" + timestamp()
    os.mkdir(results_path)
    os.mkdir(os.path.join(results_path, "db_snapshots"))
    return results_path


def timestamp():
    ts = time.time()
    return datetime.datetime.fromtimestamp(ts).strftime("%m_%d_%H_%M_%S")


def touch_wrapper(filename):
    if os.path.exists(filename):
        os.remove(filename)
    Path(filename).touch()


# Remove files if they exist, touch, and open in mode
def open_wrapper(filename, mode):
    touch_wrapper(filename)
    return open(filename, mode)
