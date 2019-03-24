#!/usr/bin/python3

import click
import datetime
import glob
import json
import os
import shlex
import subprocess
import sys
import time
import traceback

from fuzzer.lib.coverage import merge_source_coverages
from fuzzer.query import stringify_ast

BENCHMARKS_DIR_NAME = "benchmarks"
NO_PREFIX_DIR_NAME = "__BIP_NO_PREFIX"


class BIPBenchmarkComparisonException(Exception):
    def __init__(self, message, details):
        self.message = message
        self.details = details


def route_to_unix_filename(route_prefix):
    """
    Transform the given route prefix into a string that can be used as a Unix directory name.

    This works by first converting all instances of '/' to '__'.
    """

    return canonicalize_route_prefix(route_prefix).replace("/", "__")


def unix_filename_to_route(directory_name):
    """
    The inverse operation of `route_to_unix_filename` - turns a directory nameized
    directory name into a route.
    """

    if directory_name == NO_PREFIX_DIR_NAME:
        return "No prefix"

    return directory_name.replace("__", "/")


def canonicalize_route_prefix(route_prefix):
    """
    Canonicalizes the specified route prefix by ensuring that it begins with a / and does not
    end with one.
    """

    route_prefix = route_prefix if route_prefix.startswith("/") else f"/{route_prefix}"
    route_prefix = route_prefix[:-1] if route_prefix.endswith("/") else route_prefix

    return route_prefix


def create_benchmarks_dir_if_not_exists(results_path, dir_name):
    if not os.path.isdir(os.path.join(results_path, "benchmarks")):
        os.mkdir(os.path.join(results_path, "benchmarks"))

    if not os.path.isdir(os.path.join(results_path, "benchmarks", dir_name)):
        os.mkdir(os.path.join(results_path, "benchmarks", dir_name))


def benchmarks_dir_path_for_server_root(results_path):
    return os.path.join(results_path, BENCHMARKS_DIR_NAME)


def datetime_from_benchmark_filename(filename):
    filename = filename.replace(".json", "")
    year, month, day, hour, minute, second = map(int, filename.split("_"))
    return datetime.datetime(year, month, day, hour, minute, second)


def read_prefix_benchmarks_dir(results_path):
    """
    Read the benchmarks directory and return a dict keyed on prefix. Each value in the
    dict is a hash with `date` and `filename` keys, where `date` contains the date parsed from
    the filename.
    """

    benchmarks_dir_path = benchmarks_dir_path_for_server_root(results_path)
    benchmark_file_paths = glob.glob(f"{benchmarks_dir_path}/**/**")

    prefix_benchmarks = {}
    for benchmark_file_path in benchmark_file_paths:
        prefix_directory, filename = benchmark_file_path[
            len(benchmarks_dir_path) + 1 :
        ].split("/")

        filename = filename.replace(".json", "")
        date = datetime_from_benchmark_filename(filename)

        if prefix_directory not in prefix_benchmarks:
            prefix_benchmarks[prefix_directory] = []

        prefix_benchmarks[prefix_directory].append({"date": date, "filename": filename})

    return prefix_benchmarks


def last_two_benchmark_filenames_for_prefix(results_path, route_prefix):
    """
    Returns a tuple of the last two benchmark filenames for the given route prefix, or raises
    an exception if there are fewer than 2 benchmarks.
    """

    prefix_benchmarks = read_prefix_benchmarks_dir(results_path)
    prefix = (
        route_to_unix_filename(route_prefix)
        if route_prefix is not None
        else NO_PREFIX_DIR_NAME
    )
    if prefix not in prefix_benchmarks:
        raise RuntimeError(
            f"Can't find benchmark directory with route prefix '{route_prefix}'"
        )
    benchmarks = sorted(prefix_benchmarks[prefix], key=lambda hash: hash["date"])
    if len(benchmarks) < 2:
        raise RuntimeError(
            f"There aren't enough benchmarks to compare for the prefix '{route_prefix or 'all'}'"
        )

    return (benchmarks[-2]["filename"], benchmarks[-1]["filename"])


def comparison_operator_for_values(value_1, value_2):
    """
    Returns either '<', '>', or '~' depending on how `value_1` compares to `value_2`.
    """

    if abs(value_1 - value_2) < 0.00001:
        return "~"
    elif value_1 < value_2:
        return "<"
    else:
        return ">"


def compare_benchmarks(old_benchmark, new_benchmark):
    sorted_old_route_hashes = sorted(
        old_benchmark["routes"], key=lambda route: route["route"]
    )
    sorted_old_routes = [route_hash["route"] for route_hash in sorted_old_route_hashes]
    sorted_new_route_hashes = sorted(
        new_benchmark["routes"], key=lambda route: route["route"]
    )
    sorted_new_routes = [route_hash["route"] for route_hash in sorted_new_route_hashes]

    report_details = ""

    if sorted_old_routes != sorted_new_routes:
        missing_routes = [
            route for route in sorted_old_routes if route not in sorted_new_routes
        ]

        report_details += f"Old benchmark routes:\n"
        report_details += f"\t {json.dumps(sorted_old_routes)}\n"
        report_details += f"New benchmark routes:\n"
        report_details += f"\t {json.dumps(sorted_new_routes)}\n"
        report_details += f"Routes not present in new benchmark:\n"
        report_details += f"\t {json.dumps(missing_routes)}\n\n"

        raise BIPBenchmarkComparisonException(
            "A different set of routes were run in the old and new benchmarks. Details written to benchmark_deets.txt.",
            report_details,
        )

    regression_checks = [
        {"key": "sorted_parameter_paths", "name": "Parameters"},
        {"key": "sorted_exception_types", "name": "Exceptions"},
        {"key": "sorted_queries", "name": "Queries"},
    ]
    regression_check_results = []
    for regression_check_hash in regression_checks:
        regression_status = "unchanged"
        for old_route_hash, new_route_hash in zip(
            sorted_old_route_hashes, sorted_new_route_hashes
        ):
            old_items = old_route_hash[regression_check_hash["key"]]
            new_items = new_route_hash[regression_check_hash["key"]]
            if old_items != new_items:
                missing_items = [
                    old_item for old_item in old_items if old_item not in new_items
                ]
                if len(missing_items) != 0:
                    report_details += f"{regression_check_hash['name']} regression for route: {old_route_hash['route']}\n"
                    report_details += f"\t Missing {regression_check_hash['name']}: {json.dumps(missing_items)}\n\n"

                    regression_status = "regression"
                else:
                    regression_status = (
                        "improvement"
                        if regression_status != "regression"
                        else "regression"
                    )

        regression_check_results.append(
            {"name": regression_check_hash["name"], "status": regression_status}
        )

    coverage_regression_status = "unchanged"
    old_coverage = old_benchmark["source_coverage"]
    new_coverage = new_benchmark["source_coverage"]
    for filepath, old_line_counts in old_coverage.items():
        if filepath not in new_coverage:
            print(filepath)
            coverage_regression_status = "regression"

            report_details += f"Coverage regression:\n"
            report_details += f"\t File not run in new benchmark: {filepath}"

            continue

        new_line_counts = new_coverage[filepath]

        lines_regressed = []
        for index, (old_count, new_count) in enumerate(
            zip(old_line_counts, new_line_counts)
        ):
            if old_count is None or new_count is None:
                continue

            if old_count > 0 and new_count == 0:
                coverage_regression_status = "regression"
                lines_regressed.append(index + 1)
            elif old_count == 0 and new_count > 0:
                coverage_regression_status = (
                    "improvement"
                    if coverage_regression_status != "regression"
                    else "regression"
                )

        if len(lines_regressed) != 0:
            report_details += f"\nCoverage regression in {filepath}\n\t Lines: {json.dumps(lines_regressed)}\n\n"

    regression_check_results.append(
        {"name": "Coverage", "status": coverage_regression_status}
    )

    return (regression_check_results, report_details)


def print_benchmark_comparison_report(
    results_path, route_prefix, old_benchmark_name, new_benchmark_name
):
    benchmarks_dir_path = benchmarks_dir_path_for_server_root(results_path)
    benchmarks_directory_name = (
        route_to_unix_filename(route_prefix)
        if route_prefix is not None
        else NO_PREFIX_DIR_NAME
    )

    old_benchmark_filepath = os.path.join(
        benchmarks_dir_path,
        benchmarks_directory_name,
        old_benchmark_name.replace(".json", "") + ".json",
    )
    new_benchmark_filepath = os.path.join(
        benchmarks_dir_path,
        benchmarks_directory_name,
        new_benchmark_name.replace(".json", "") + ".json",
    )
    benchmark_deets_filepath = os.path.join(results_path, "benchmark_deets.txt")

    if not os.path.isfile(old_benchmark_filepath):
        print(f"Can't find old benchmark file at {old_benchmark_filepath}")
        sys.exit(1)
    if not os.path.isfile(new_benchmark_filepath):
        print(f"Can't find new benchmark file at {new_benchmark_filepath}")
        sys.exit(1)

    old_benchmark_date = datetime_from_benchmark_filename(old_benchmark_name)
    new_benchmark_date = datetime_from_benchmark_filename(new_benchmark_name)
    print(
        f"\nComparing benchmarks run on '{old_benchmark_date.strftime('%b %d, %Y %-I:%M:%S %p')}' and '{new_benchmark_date.strftime('%b %d, %Y %-I:%M:%S %p')}'...\n"
    )

    with open(old_benchmark_filepath, "r") as old_benchmark_file:
        old_benchmark = json.loads(old_benchmark_file.read())
    with open(new_benchmark_filepath, "r") as new_benchmark_file:
        new_benchmark = json.loads(new_benchmark_file.read())

    try:
        (regression_check_results, report_details) = compare_benchmarks(
            old_benchmark, new_benchmark
        )
    except BIPBenchmarkComparisonException as e:
        print(e.message)
        print("\nAborting!\n")
        with open(benchmark_deets_filepath, "w") as benchmark_deets_file:
            benchmark_deets_file.write(e.details)
        sys.exit(1)
    except RuntimeError as e:
        print(str(e))
        print("\nAborting!\n")
        sys.exit(1)

    for result_hash in regression_check_results:
        status_emoji = {"regression": "🚨", "unchanged": "✅", "improvement": "🎉"}[
            result_hash["status"]
        ]
        print(
            f"{(result_hash['name'] + ':').ljust(12)} {status_emoji} {result_hash['status'].upper()}"
        )

    with open(benchmark_deets_filepath, "w") as benchmark_deets_file:
        benchmark_deets_file.write(report_details)

    print()
    print("             New           Old")

    new_coverage_percentage = new_benchmark["stats"]["coverage_percentage"]
    old_coverage_percentage = old_benchmark["stats"]["coverage_percentage"]
    coverage_operator = comparison_operator_for_values(
        new_coverage_percentage, old_coverage_percentage
    )
    print(
        f"Coverage %:  {(str(round(new_coverage_percentage, 4)) + '%').ljust(10)} {coverage_operator}  {round(old_coverage_percentage, 4)}%"
    )

    new_run_time = new_benchmark["stats"]["fuzzer_run_time"]
    old_run_time = old_benchmark["stats"]["fuzzer_run_time"]
    run_time_operator = comparison_operator_for_values(new_run_time, old_run_time)
    print(
        f"Run time:    {(str(round(new_run_time, 4)) + 's').ljust(10)} {run_time_operator}  {round(old_run_time, 4)}s"
    )

    print()
    print("Deets written to benchmark_deets.txt")
    print()


def calculate_coverage_percentage(coverage_data):
    total_number_runnable_lines = 0
    total_number_lines_run = 0
    for line_counts in coverage_data.values():
        for count in line_counts:
            if count is None:
                continue
            total_number_runnable_lines += 1
            if count > 0:
                total_number_lines_run += 1
    coverage_percentage = (
        float(total_number_lines_run) / total_number_runnable_lines * 100
    )
    return coverage_percentage


def check_fuzzer(fuzzer_process, instance=None):
    print("check fuzzer")
    try:
        fuzzer_process.communicate(timeout=60 * 8)
    except subprocess.TimeoutExpired:
        print("Error: Fuzzer timed out!")
        fuzzer_process.kill()
        fuzzer_process.communicate()
        return -1
    if fuzzer_process.returncode != 0:
        print(
            f"\nFuzzer process returned non-zero exit code: {fuzzer_process.returncode}"
        )
        instance = instance if instance else 0
        log = "/tmp/fuzz_%d" % instance
        print("See %s for stdout/stderr" % log)
    return fuzzer_process.returncode


def run_fuzzer(
    results_path,
    routes_file_name,
    db_name,
    state,
    port=None,
    target_app=None,
    quiet=None,
    fuzzer_number=None,
    instances=None,
):
    """
    Run the fuzzer in a subprocess and return the return code of the process.
    """
    cmds = [
        "python3",
        "fuzz.py",
        "--routes",
        routes_file_name,
        "--output-benchmark-data",
        results_path,
        db_name,
        state,
        "0",
    ]
    if port:
        cmds += ["--port", str(port)]
    if target_app:
        cmds += ["--target_app", target_app]
    if fuzzer_number == 0 or fuzzer_number:
        cmds += ["--fuzzer_number", str(fuzzer_number)]
        cmds += ["--instances", str(instances)]

    # print(" ".join(cmds))
    fuzzer_number = fuzzer_number if fuzzer_number else 0
    log = "/tmp/fuzz_%d" % fuzzer_number
    fp = open(log, "w")
    fuzzer_process = subprocess.Popen(cmds, stdout=fp, stderr=fp, stdin=subprocess.PIPE)
    print("Fuzzer is up")
    return fuzzer_process


def kill_server(name):
    cmd = "docker rm -f %s" % name
    subprocess.run(shlex.split(cmd))


def spawn_server(results_dir, db, port, instance):
    name = "server_%d" % instance
    if "dante" in results_dir:
        cmd = (
            'docker run --net=host --name=%(name)s --rm -e "DEV_DB=%(db)s" -e "QUIET=1" -v '
            "/home/marsbar/dante_results:/dante_results -v /etc/passwd:/etc/passwd -v"
            '/var/run/postgresql:/var/run/postgresql -e "RESULTS_PATH=/%(results_dir)s"'
            '-e "PORT=%(port)s" --entrypoint=/bin/bash test_img5 -c "cd /dante-stories-fork; '
            './run.sh"'
        )
        cmd = cmd % {
            "db": db,
            "results_dir": results_dir,
            "port": str(port),
            "name": name,
        }
    elif "discourse" in results_dir:
        cmd = (
            'docker run --net=host --name=%(name)s --rm -e "DISCOURSE_DEV_DB=%(db)s" -e "QUIET=1" -v '
            "/home/marsbar/discourse_results:/discourse_results -v /etc/passwd:/etc/passwd -v "
            '/var/run/postgresql:/var/run/postgresql -e "RESULTS_PATH=/discourse_results" '
            '-e "PORT=%(port)s" test_img5'
        )
        cmd = cmd % {"db": db, "port": str(port), "name": name}
    else:
        assert False
    print(cmd)
    log = "/tmp/%s.log" % name
    fp = open(log, "w")
    proc = subprocess.Popen(
        shlex.split(cmd), stderr=fp, stdout=fp, stdin=subprocess.PIPE, shell=False
    )
    # Wait for server to spin up
    try:
        out, err = proc.communicate(timeout=5)
        fp.close()
        print(open(log, "r").read())
        return None
    except subprocess.TimeoutExpired:
        pass
    return name


@click.group()
def cli():
    pass


DANTE_DB = ["stories0", "stories1", "stories2", "stories3"]
DISCOURSE_DB = [
    "discourse_development0",
    "discourse_development1",
    "discourse_development2",
    "discourse_development3",
]
PORTS = [4000, 4001, 4002, 4003]


@cli.command()
@click.argument("results_path", type=click.Path(exists=True))
@click.argument("state", type=click.Path(exists=True))
@click.argument("target_app", type=click.Path(exists=True))
@click.argument("instances", type=int)
def parallel(results_path, state, target_app, instances):
    """
    Run dante or discourse in parallel
    """
    if "dante" in results_path:
        db_names = DANTE_DB
    elif "discourse" in results_path:
        db_names = DISCOURSE_DB
    else:
        assert False

    server_names = []
    failed = False
    try:
        for i in range(0, instances):
            print("Starting server number %d" % i)
            name = spawn_server(results_path, db_names[i], PORTS[i], i)
            if name is None:
                failed = True
                break
            server_names.append(name)
            print("Done")
        fuzzer_run_time = 0
        ret = 0
        start_time = time.time()
        procs = []
        for i in range(0, instances):
            fuzz_dir = "fuzz_%d" % PORTS[i]
            results = os.path.join(results_path, fuzz_dir)
            routes_file = os.path.join(state, "routes.json")
            proc = run_fuzzer(
                results,
                routes_file,
                db_names[i],
                state,
                port=PORTS[i],
                target_app=target_app,
                fuzzer_number=i,
                instances=instances,
            )
            procs.append(proc)
        for i in range(0, instances):
            ret = check_fuzzer(procs[i], i)
            if ret != 0:
                failed = True
                break
        fuzzer_run_time = time.time() - start_time
    except Exception as e:
        value, ty, tb = sys.exc_info()
        traceback.print_exception(value, ty, tb)
    for name in server_names:
        print("killing %s" % name)
        kill_server(name)
    if failed:
        exit(1)

    # union test_cov files
    unioned_cov = {}
    for i in range(0, instances):
        fuzz_dir = "fuzz_%d" % PORTS[i]
        results = os.path.join(results_path, fuzz_dir)
        with open(os.path.join(results, "test_coverage.json"), "r") as coverage_file:
            cov = json.loads(coverage_file.read())
            unioned_cov = merge_source_coverages(unioned_cov, cov)
    coverage_percentage = calculate_coverage_percentage(unioned_cov)
    print("coverage percentage: %f" % coverage_percentage)
    print("fuzzer run time: {}".format(fuzzer_run_time))


@cli.command()
@click.argument("results_path", type=click.Path(exists=True))
@click.option("--route-prefix", help="Compare output for this route prefix")
@click.option(
    "--benchmarks",
    nargs=2,
    help="The names of 2 benchmark files to compare, as output by the `list` command."
    "If not specified, this command will compare the last two benchmarks instead.",
)
def compare(results_path, route_prefix, benchmarks):
    """
    Compare two existing benchmarks.
    """

    if len(benchmarks) != 2:
        try:
            (
                old_benchmark_name,
                new_benchmark_name,
            ) = last_two_benchmark_filenames_for_prefix(results_path, route_prefix)
        except RuntimeError as e:
            print(str(e), "- aborting!")
            sys.exit(1)
    else:
        (b1, b2) = benchmarks
        if datetime_from_benchmark_filename(b1) < datetime_from_benchmark_filename(b2):
            old_benchmark_name = b1
            new_benchmark_name = b2
        else:
            old_benchmark_name = b2
            new_benchmark_name = b1

    print_benchmark_comparison_report(
        results_path, route_prefix, old_benchmark_name, new_benchmark_name
    )


@cli.command()
@click.argument("results_path", type=click.Path(exists=True))
@click.argument("state", type=click.Path(exists=True))
@click.argument("target_app", type=click.Path(exists=True))
@click.option("--db_name", help="db to connect to")
@click.option(
    "--route-prefix", help="Only generate benchmark on routes with this prefix"
)
@click.option("--port", help="port to query")
def new(results_path, state, target_app, route_prefix, db_name, port):
    """
    Run a new benchmark, connecting to the given port and db.
    Results are written to RESULTS_PATH.
    """
    route_prefix = (
        canonicalize_route_prefix(route_prefix) if route_prefix is not None else None
    )

    # If a route prefix was specified, first write out a new benchmark_routes.json file with
    # the routes we care about.
    routes_file_name = "routes.json"
    if route_prefix is not None:
        routes_file_name = "benchmark_routes.json"
        routes = []
        routes_file_path = os.path.join(results_path, "routes.json")
        with open(routes_file_path, "r") as routes_file:
            for line in routes_file.readlines():
                if line.strip() == "":
                    continue
                routes.append(json.loads(line))

        matching_routes = [
            route for route in routes if route["path"].startswith(route_prefix)
        ]

        benchmark_routes_file_path = os.path.join(state, routes_file_name)
        with open(benchmark_routes_file_path, "w") as benchmark_routes_file:
            for route in matching_routes:
                benchmark_routes_file.write(json.dumps(route) + "\n")

    num_routes = None
    with open(os.path.join(state, routes_file_name), "r") as routes_file:
        num_routes = len(routes_file.readlines())

    start_time = time.time()
    proc = run_fuzzer(
        results_path,
        routes_file_name,
        db_name,
        state,
        port=port,
        target_app=target_app,
        quiet=quiet,
    )
    ret = check_fuzzer(proc)
    assert ret == 0
    fuzzer_run_time = time.time() - start_time

    coverage_data = {}
    with open(os.path.join(results_path, "test_coverage.json"), "r") as coverage_file:
        coverage_data = json.loads(coverage_file.read())

    coverage_percentage = calculate_coverage_percentage(coverage_data)

    visited_routes = []
    with open(
        os.path.join(results_path, "benchmark_data.json"), "r"
    ) as benchmark_data_file:
        visited_routes = json.loads(benchmark_data_file.read())

    routes = []
    for route in visited_routes:
        parameters = []
        exceptions = []
        queries = []
        for request in route["requests"]:
            for json_path in request["params"].keys():
                if json_path not in parameters:
                    parameters.append(json_path)

            for exception in request["exceptions"]:
                exception_class = exception["class"]
                if exception_class not in exceptions:
                    exceptions.append(exception_class)

            for query in request["queries"]:
                if query["ast"] is not None:
                    stringified_query = stringify_ast(query["ast"], canonicalize=True)
                    if stringified_query not in queries:
                        queries.append(stringified_query)

        routes.append(
            {
                "route": f"{route['verb']} {route['path']}",
                "sorted_parameter_paths": sorted(parameters),
                "sorted_exception_types": sorted(exceptions),
                "sorted_queries": sorted(queries),
            }
        )

    # Grab hash of fuzzer repo
    commit_hash = subprocess.check_output(["git", "rev-parse", "HEAD"])
    # TODO: grab hash of target app repo
    # Right now we are only passing around a results dir.
    # We can optionally specify a path to target app so that we can grab
    # the hash.
    # commit_hash = subprocess.check_output(
    #    ["git", "rev-parse", "HEAD"], cwd=results_path
    # )

    stats = {
        "fuzzer_run_time": fuzzer_run_time,
        "num_routes": num_routes,
        "coverage_percentage": coverage_percentage,
        "fuzzer commit_hash": commit_hash.decode("utf-8").strip(),
    }

    run_data = {"routes": routes, "source_coverage": coverage_data, "stats": stats}

    output_file_name = f"{datetime.datetime.now().strftime('%Y_%m_%d_%H_%M_%S')}.json"
    output_directory = (
        route_to_unix_filename(route_prefix)
        if route_prefix is not None
        else NO_PREFIX_DIR_NAME
    )
    create_benchmarks_dir_if_not_exists(results_path, output_directory)

    with open(
        os.path.join(results_path, "benchmarks", output_directory, output_file_name),
        "w",
    ) as output_file:
        output_file.write(json.dumps(run_data))

    print(
        f"Data dumped to {os.path.join('benchmarks', output_directory, output_file_name)}"
    )

    try:
        (
            old_benchmark_name,
            new_benchmark_name,
        ) = last_two_benchmark_filenames_for_prefix(results_path, route_prefix)
        print_benchmark_comparison_report(
            results_path, route_prefix, old_benchmark_name, new_benchmark_name
        )
    except RuntimeError:
        # There aren't yet 2 benchmarks to compare, so just ignore this exception.
        pass


def stats_for_benchmark_file(results_path, directory_name, filename):
    """
    Read and parse the specified benchmark file and return the stats.
    """

    filepath = os.path.join(
        benchmarks_dir_path_for_server_root(results_path),
        directory_name,
        filename + ".json",
    )

    with open(filepath, "r") as benchmark_file:
        return json.loads(benchmark_file.read())["stats"]


@cli.command()
@click.argument("results_path", type=click.Path(exists=True))
@click.option(
    "--limit",
    type=int,
    help="The number of the most recent benchmarks to list per prefix. Specify 0 to list all.",
    default=5,
)
@click.option(
    "--route-prefix",
    help="Only list routes with this prefix (or specify 'unprefixed' to list benchmarks run on all routes)",
)
def list(results_path, limit, route_prefix):
    """
    List past benchmarks.
    """

    prefix_benchmarks = read_prefix_benchmarks_dir(results_path)

    for prefix_directory, benchmark_dicts in prefix_benchmarks.items():
        if route_prefix is not None:
            if route_prefix == "unprefixed":
                if prefix_directory != NO_PREFIX_DIR_NAME:
                    continue
            elif prefix_directory != route_to_unix_filename(route_prefix):
                continue

        benchmark_dicts = sorted(benchmark_dicts, key=lambda hash: hash["date"])

        if limit != 0:
            benchmark_dicts = benchmark_dicts[-1 * limit :]

        print(unix_filename_to_route(prefix_directory))
        for hash in benchmark_dicts:
            stats = stats_for_benchmark_file(
                results_path, prefix_directory, hash["filename"]
            )

            print(
                "\t",
                hash["date"].strftime("%b %d, %Y %-I:%M:%S %p"),
                f"({round(stats['coverage_percentage'], 2)}% in {round(stats['fuzzer_run_time'], 2)}s | {hash['filename']})",
            )


if __name__ == "__main__":
    cli()
