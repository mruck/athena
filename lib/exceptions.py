import json

import fuzzer.lib.util as util

BENIGN_EXCEPTIONS = [
    #    "ActionController::RoutingError",
    #    "ActionController::ParameterMissing",
    "ActiveRecord::RecordNotFound"
]


class Exception(object):
    # TODO: Add backtrace!!!!
    def __init__(self, verb, path, cls, message):
        self.verb = verb
        self.path = path
        self.cls = cls
        self.message = message

    @classmethod
    def from_dict(cls, exn_dict, route):
        # TODO: pull out more stuff from rpoute, like params
        cls(route.verb, route.path, exn_dict["class"], exn_dict["msg"])


# Keep track of unique exceptions as well as pointer to exceptions log dumped
# by rails
class ExceptionTracker(object):
    def __init__(self, exceptions_file):
        self.exceptions_file_pointer = util.open_wrapper(exceptions_file, "r")

    def merge(self, new_exns):
        delta_exns = []
        for new_exn in new_exns:
            found = False
            for unique_exn in self.unique_exceptions:
                if (
                    new_exn.verb == unique_exn.verb
                    and new_exn.path == unique_exn.path
                    and new_exn.cls == unique_exn.cls
                ):
                    found = True
                    break
            if not found:
                self.unique_exns.append(new_exn)
                delta_exns.append(new_exn)
                with open("/tmp/exn", "a") as f:
                    f.write("%s %s\n" % (new_exn.path, new_exn.cls))
        return delta_exns

    # Read exceptions from the exception log and update the list of unique
    # exceptions
    def update(self, route):
        exns = []
        print("**************************")
        for line in self.exceptions_file_pointer:
            print(line)
            x = json.loads(line.strip())
            print(x)
            exns.append(x)
        print("???????????????????????????????")
        # Filter out benign exceptions
        malign_exns = [e for e in exns if e["class"] not in BENIGN_EXCEPTIONS]

        # Instantiate Exception objs
        exn_objs = [Exception.from_dict(e, route) for e in malign_exns]
        with open("/tmp/exn", "a") as f:
            f.write("unfiltered exns:\n")
            for e in exns:
                f.write("%s\n" % (e["class"]))
            f.write("malign exns:\n")
            for e in malign_exns:
                f.write("%s\n" % (e["class"]))
            f.write("malign exns objs:\n")
            for e in exn_objs:
                f.write("%s %s\n" % (e.path, e.cls))
        # Merge with the unique exceptions
        delta_exns = self.merge(exn_objs)
        return []
