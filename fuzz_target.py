import json
import os
import sys
import traceback

import fuzzer.lib.coverage as coverage
import fuzzer.lib.util as util

STATE = "/state"


class Target(object):
    """
    Get file pointers to files written to by rails. Only open the file once
    so we can keep track of our position.  Open in a+ in case the file doesn't
    exist yet. This is a bit confusing because these files should be read only
    by the fuzzer!
    """

    def __init__(self, results_path, port, db, snapshot=None):
        self.results_path = results_path
        self.state = STATE
        self.port = port
        self.db = db
        # Remove files if they exist, touch, and open as RO
        self.params_file = util.open_wrapper(os.path.join(results_path, "params"), "r")
        self.queries_file = util.open_wrapper(
            os.path.join(results_path, "queries"), "r"
        )
        self.cov = coverage.Coverage(os.path.join(results_path, "src_line_coverage"))
        # Exceptions dumped by rails
        self.rails_exceptions = util.open_wrapper(
            os.path.join(results_path, "rails_exception_log.json"), "r"
        )
        # Exceptions dumped by fuzzer
        self.fuzzer_exceptions = util.open_wrapper(
            os.path.join(results_path, "fuzzer_exceptions"), "a"
        )
        self.snapshot = snapshot
        os.environ["RESULTS_PATH"] = results_path

    def latest_exns(self):
        return [json.loads(line.strip()) for line in self.rails_exceptions]

    def on_fuzz_exception(self, route, state_dir):
        etype, val, tb = sys.exc_info()
        self.fuzzer_exceptions.write("***%s %s***\n" % (route.verb, route.path))
        self.fuzzer_exceptions.write("State saved at %s\n" % (state_dir))
        traceback.print_exception(etype, val, tb, file=self.fuzzer_exceptions)
        self.fuzzer_exceptions.write("\n")
