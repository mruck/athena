import pymongo
import sys


def get_host():
    # We are in a pod
    if "linux" in sys.platform:
        return "mongodb-service"
    # We are running locally on osx
    elif "darwin" in sys.platform:
        return "localhost"
    else:
        print("Untested architecture %s" % sys.platform)
        assert False


class Connection(object):
    def __init__(self, db_name):
        target = "mongodb://" + get_host() + ":27017/"
        client = pymongo.MongoClient(target)
        self.db = client[db_name]
        self.is_alive()

    def is_alive(self):
        print("pinging...")
        result = self.db.command("ping")
        assert result == {u"ok": 1.0}

    def write(self, table, payload):
        self.db[table].insert_one(payload)

    def read(self, table, key):
        return self.db[table].find_one(key)

    def read_all(self, table):
        entries = []
        for entry in self.db[table].find():
            entries.append(entry)
        return entries
