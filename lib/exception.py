import json

import fuzzer.lib.util as util

BENIGN_EXCEPTIONS = [
    "ActionController::RoutingError",
    "ActionController::ParameterMissing",
    "ActiveRecord::RecordNotFound",
]


# Keep track of unique exceptions as well as pointer to exceptions log dumped
# by rails
class ExceptionTracker(object):
    def __init__(self, exceptions_file):
        self.exceptions_file_pointer = util.open_wrapper(exceptions_file, "r")

    # Read exceptions from the exception log and update the list of unique
    # exceptions
    def update(self):
        exns = [json.loads(line.strip()) for line in self.exceptions_file_pointer]
        # Filter out benign exceptions
        malign_exns = [e for e in exns if e["class"] not in BENIGN_EXCEPTIONS]
        return malign_exns
