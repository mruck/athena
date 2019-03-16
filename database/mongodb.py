import pymongo


def new_db(db_name):
    client = pymongo.MongoClient()
    return client[db_name]


def new_table(db, table_name):
    return db[table_name]


# TODO: Pinging is broken
def is_alive(db):
    print("pinging...")
    result = db.command("ping")
    return result == {u"ok": 1.0}


def find_one(table, needle):
    return table.find_one(needle)
