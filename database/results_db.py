import fuzzer.database.mongodb as db_lib

EXCEPTIONS_TABLE = "exceptions_table"


class ResultsDb(object):
    def __init__(self, db_name):
        self.db = db_lib.new_db(db_name)
        assert db_lib.is_alive(self.db)
        self.exceptions_table = db_lib.new_table(self.db, EXCEPTIONS_TABLE)

    def write_one(self, table, payload):
        table.insert_one(payload)

    def write_exceptions(self, exceptions):
        for e in exceptions:
            self.write_one(self.exceptions_table, e.to_dict())

    def write_sql_inj(self):
        pass

    def write_xss(self):
        pass

    def write_coverage(self):
        pass

    def find_exception(self, exception_class):
        return self.exceptions_table.find_one(exception_class)

    def print_all_exceptions(self):
        for exn in self.exceptions_table.find():
            print(exn)
