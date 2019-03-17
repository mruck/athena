import json

import fuzzer.lib.util as util

BENIGN_EXCEPTIONS = [
    "ActionController::RoutingError",
    "ActionController::ParameterMissing",
    "ActiveRecord::RecordNotFound",
]


# Class for an exception raised by the target
class TargetException(object):
    # TODO: Add backtrace!!!!
    def __init__(self, verb, path, cls, message):
        self.verb = verb
        self.path = path
        self.cls = cls
        self.message = message

    @classmethod
    def from_dict(cls, exn_dict, route):
        # TODO: pull out more stuff from route, like params
        return cls(route.verb, route.path, exn_dict["class"], exn_dict["msg"])

    def to_dict(self):
        return {
            "verb": self.verb,
            "path": self.path,
            "class": self.cls,
            "message": self.message,
        }


def is_equal(exn1: TargetException, exn2: TargetException):
    # Compare 2 exceptions for equality
    return (
        exn1.verb == exn2.verb
        and exn1.path == exn2.path
        and exn1.cls == exn2.cls
        and exn1.message == exn2.message
    )


# Keep track of unique exceptions as well as pointer to exceptions log dumped
# by rails
class ExceptionTracker(object):
    def __init__(self, exceptions_file):
        self.exceptions_file_pointer = util.open_wrapper(exceptions_file, "r")
        self.unique_exceptions = []

    # Merge new exceptions with global list of unique exceptions
    def merge(self, new_exns):
        delta_exns = []
        for new_exn in new_exns:
            found = False
            for unique_exn in self.unique_exceptions:
                if is_equal(new_exn, unique_exn):
                    found = True
                    break
            if not found:
                self.unique_exceptions.append(new_exn)
                delta_exns.append(new_exn)
                with open("/tmp/exn2", "a") as f:
                    f.write("%s %s %s\n" % (new_exn.verb, new_exn.path, new_exn.cls))
        return delta_exns

    # Read exceptions from the exception log and update the list of unique
    # exceptions
    def update(self, route):
        exns = []
        # Read raw dump of exceptions
        for line in self.exceptions_file_pointer:
            exns.append(json.loads(line.strip()))
        # Filter out benign exceptions
        malign_exns = [e for e in exns if e["class"] not in BENIGN_EXCEPTIONS]
        # Instantiate Exception objs
        exn_objs = [TargetException.from_dict(e, route) for e in malign_exns]
        # Merge with the unique exceptions
        delta_exns = self.merge(exn_objs)
        return delta_exns
