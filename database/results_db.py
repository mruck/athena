import fuzzer.database.mongodb as db
import fuzzer.lib.util as util

EXCEPTIONS_TABLE = "exceptions"


class ResultsDb(object):
    def __init__(self, db_name):
        self.db = db.Connection(db_name)

    def write_one(self, table, payload):
        self.db.write(table, payload)

    def write_exceptions(self, exceptions):
        for e in exceptions:
            exn_dict = e.to_dict()
            exn_dict["TargetID"] = util.get_target_id()
            self.write_one(EXCEPTIONS_TABLE, exn_dict)

    def write_sql_inj(self):
        pass

    def write_xss(self):
        pass

    def write_coverage(self):
        pass

    def find_exception_by_key(self, key):
        return self.db.read(EXCEPTIONS_TABLE, key)

    def print_all_exceptions(self):
        print(self.db.read_all(EXCEPTIONS_TABLE))
