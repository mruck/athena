import glob
import json
import os

import fuzzer.lib.util as util


class Coverage(object):
    def __init__(self, src_cov_file):
        self.cumulative_coverage = {}
        self.src_cov_file = util.open_wrapper(src_cov_file, "r")

    def update_cumulative_cov(self, new_coverage):
        """
        Given new coverage, updates the cumulative cov and returns the delta
        """
        deltas = {}
        for filepath, line_counts in new_coverage.items():
            if filepath not in self.cumulative_coverage:
                # We hit a new file, update cumulative cov
                self.cumulative_coverage[filepath] = line_counts
                deltas[filepath] = line_counts
            else:
                deltas[filepath] = line_counts
                for index, count in enumerate(line_counts):
                    if count is None or count == 0:
                        deltas[filepath][index] = count
                        continue
                    # We've seen this file before, but we hit a new line
                    if self.cumulative_coverage[filepath][index] == 0:
                        self.cumulative_coverage[filepath][index] = count
                        deltas[filepath][index] = count
                    else:
                        # This doesn't count as a change
                        deltas[filepath][index] = 0
                        # Still add to the cumulative count
                        self.cumulative_coverage[filepath][index] += count
            # Check if things changed
            delta_prime = [count for count in deltas[filepath] if count]
            if delta_prime == []:
                # This file contains no deltas
                del deltas[filepath]
        return deltas

    def read_src_cov(self):
        """
        Read source cov dumped by rails and return new coverage
        """
        new_coverage = {}
        for json_line in self.src_cov_file:
            file_coverages = json.loads(json_line)

            for filepath, line_counts in file_coverages.items():
                if filepath not in new_coverage:
                    new_coverage[filepath] = line_counts
                else:
                    for index, count in enumerate(line_counts):
                        if count is None or count == 0:
                            continue
                        new_coverage[filepath][index] += count

        return new_coverage

    def update(self):
        """
        Read the coverage dumped by rails, update the cumulative coverage, and return the delta
        """
        new_cov = self.read_src_cov()
        delta = self.update_cumulative_cov(new_cov)
        return delta


# TODO: new vs deltas
def log_src_deltas(deltas, results_path):
    percentages = {}
    for filepath, line_counts in deltas.items():
        runnable_count = [count for count in line_counts if count is not None]
        run_count = len([count for count in runnable_count if count > 0])
        percentages[filepath] = round(float(run_count) / len(runnable_count) * 100, 2)
    with open(os.path.join(results_path, "coverage_deltas.json"), "a") as f:
        json.dump(percentages, f)
        f.write("\n")


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
    try:
        coverage_percentage = (
            float(total_number_lines_run) / total_number_runnable_lines * 100
        )
    except ZeroDivisionError:
        print("Error collecting coverage")
        return 0
    return coverage_percentage


def coverage_is_equal(coverage1, coverage2):
    if (
        len(set(coverage1.keys()) - set(coverage2.keys())) != 0
        or len(set(coverage2.keys()) - set(coverage1.keys())) != 0
    ):
        return False

    for filepath, line_counts1 in coverage1.items():
        line_counts2 = coverage2[filepath]

        for (count1, count2) in zip(line_counts1, line_counts2):
            if count1 != count2:
                return False

    return True


def merge_source_coverages(coverage1, coverage2):
    new_coverage = dict(coverage1)

    for filepath, line_counts in coverage2.items():
        if filepath not in new_coverage:
            new_coverage[filepath] = list(line_counts)
        else:
            for index, count in enumerate(line_counts):
                if count is None:
                    continue
                new_coverage[filepath][index] += count

    return new_coverage


# total_coverage is a dictionary where the key is the filename and the
# value is the array of file lines hit
# Returns a new dictionary with the run count and runnable count
def calculate_source_coverage_stats(total_coverage):
    all_line_counts = [count for counts in total_coverage.values() for count in counts]
    runnable_counts = [count for count in all_line_counts if count is not None]
    run_count = len([count for count in runnable_counts if count > 0])

    return {
        "run_count": run_count,
        "runnable_count": len(runnable_counts),
        "filepaths": total_coverage.keys(),
        "total_coverage": total_coverage,
    }


def filter_ruby_files(filepaths):
    pass
    # app_directory_path = os.path.join(target_app, "app")

    # ruby_filepaths = glob.glob("{}/**/*.rb".format(app_directory_path), recursive=True)

    # filepaths_not_run = [
    #    filepath for filepath in ruby_filepaths if filepath not in filepaths
    # ]


def dump_coverage_to_js(filepaths, line_coverage):
    # Write out a data file for use by the line highlighter app.
    file_contents = {}
    for filepath in filepaths:
        with open(filepath, "r") as f:
            file_contents[filepath] = f.read()

    with open(os.path.join("/tmp/coverage-data.js"), "w") as data_file:
        data_file.write(
            "export const fileContents = {};".format(json.dumps(file_contents))
        )
        data_file.write("\n")
        data_file.write(
            "export const lineCountData = {};".format(json.dumps(line_coverage))
        )


# Given a coverage dictionary mapping file paths to an array of lines hit,
# filter out only the ruby files and dump to JS so coverage-visualizer can run
def process_coverage(cumulative_coverage, target_app=None):
    coverage_stats = calculate_source_coverage_stats(cumulative_coverage)

    filepaths = coverage_stats["filepaths"]
    total_coverage = coverage_stats["total_coverage"]

    # ruby_filepaths = filer_ruby_files(filepaths)
    dump_coverage_to_js(filepaths, total_coverage)


def union_cov(cov_files):
    """
    Union given coverage files
    """
    unioned_cov = {}
    for cov_file in cov_files:
        with open(cov_file, "r") as f:
            cov = json.loads(f.read())
            unioned_cov = merge_source_coverages(unioned_cov, cov)
    coverage_percentage = calculate_coverage_percentage(unioned_cov)
    print("coverage percentage: %f" % coverage_percentage)
