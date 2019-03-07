# DB API.  Let's abstract away the db and swap out as necessary.
# Currently we are importing from postgres.py but we can trivially
# swap out to sqlite.py

import json
import logging
import os
import requests
import shlex
import subprocess

import fuzzer.database.postgres as postgres
from fuzzer.lib.util import random_int, get_results_path, timestamp

logger = logging.getLogger("debug")

pluralizations = None
RESULTS_PATH = "/tmp/results"


def init_pluralization():
    global pluralizations
    with open(os.path.join(RESULTS_PATH, "pluralizations.json")) as pluralizations_file:
        pluralizations = json.loads(pluralizations_file.read())


def pluralize(table):
    global pluralizations
    return pluralizations[table] if table in pluralizations else table


# Tell rails to drop all connections to the db so we can reset
def clear_rails_connections(hostname="localhost", port=3000):
    target = "http://{}:{}".format(hostname, port) + "/rails/info/clear_all_connections"
    requests.get(target)


# Check if a tablel and col exist.  Called from fuzzy_match
def table_col_exists(table, col):
    table = pluralize(table)
    query = (
        "SELECT column_name FROM information_schema.columns WHERE "
        "table_name='%s' AND column_name='%s'" % (table, col)
    )
    return postgres.run_query(query)


def build_count_query(table, query_str):
    table = pluralize(table)
    query = "SELECT COUNT(*) FROM %s " % table
    if query_str:
        query += "WHERE %s" % query_str
    return query


def build_query(table, query_str, count):
    table = pluralize(table)
    random_offset = random_int(count) - 1
    query = "SELECT * FROM %s " % (table)
    if query_str:
        query += "WHERE %s " % query_str
    query += "OFFSET %s LIMIT 1" % random_offset
    return query


def impose_constraints(val, constraints):
    constraints = constraints or []
    # TODO: As soon as we hit a route with constraints, implement this
    # functionality. Otherwise hold off for now
    return val


class EmptyTableError(Exception):
    """
    Raised when we a query an empty table. Specifically, We mapped a
    table and col to a param, but the table is empty. See no_records.log
    for details.
    """

    pass


# Get number of rows in a table
def count_rows(table, query_str=""):
    count_query = build_count_query(table, query_str)
    count = postgres.run_query(count_query)
    if count is None or count["count"] == 0:
        with open(os.path.join(get_results_path(), "no_records.log"), "a") as f:
            f.write(count_query + "\n")
        return 0
    else:
        return count["count"]


# Lookup a record in a db.  If no records satisfy the query, record the
# query in "no_record.log" and run a simple query to pop a random
# record from the table (ie make query_str = "")
def lookup(table, col, query_str="", constraints=None, can_fail=False):
    count = count_rows(table, query_str=query_str)
    # There are no records that satisfy this query
    if count == 0:
        if query_str == "":
            # The table is empty, we can't make this query any simpler
            return None
        else:
            # Simplify the query and try again
            return lookup(table, col, query_str="")
    query = build_query(table, query_str, count)
    record = postgres.run_query(query)
    # count > 0. There should always be a record unless we are racy.
    assert record
    return impose_constraints(record[col], constraints)


def snapshot_db(results_path, db_name, filename):
    with open(os.path.join(results_path, "db_snapshots", filename), "w") as dumpfile:
        subprocess.run(["pg_dump", db_name], stdout=dumpfile)


# Filter out tables with metadata and only return relevant tables
def find_tables():
    query = "SELECT * FROM information_schema.tables WHERE table_schema NOT IN ('pg_catalog', 'information_schema')"
    results = postgres.run_query(query, return_all_records=True)
    TABLE_INDEX = 2
    tables = [t[TABLE_INDEX] for t in results]
    return tables


XSS_PAYLOAD = "<script>alert('%s')</script>"


def find_text_cols():
    # Get relevant tables
    good_tables = find_tables()
    # Get all column names
    query = "SELECT * FROM information_schema.columns"
    results = postgres.run_query(query, return_all_records=True)
    DATA_TYPE_INDEX = 27
    TABLE_INDEX = 2
    COL_INDEX = 3
    # Tables are keys. Values are a list of column names
    text_cols = {}
    for r in results:
        table = r[TABLE_INDEX]
        # We don't care about this table
        if table not in good_tables:
            continue
        data_type = r[DATA_TYPE_INDEX]
        if data_type in ["char", "_char", "varchar"]:
            col_name = r[COL_INDEX]
            if table in text_cols:
                text_cols[table].add(col_name)
            else:
                text_cols[table] = set([col_name])
    return text_cols


def inject_xss_payload(text_cols):
    for table, cols in text_cols.items():
        cols = list(cols)
        num_cols = len(cols)
        cols = [
            col + """= CONCAT('<script>alert("BIP_',id,'")</script>')""" for col in cols
        ]
        cols = ",".join(cols)
        query = "UPDATE %s SET %s" % (table, cols)
        # Try running a query to update every record with an alert box and the id col
        postgres.run_query(query, can_fail=True)


CREATE_DB = 'docker run --volumes-from my-postgres  %s -e "DB_NAME=%s" --rm fuzzer-db'

# Rails cannot connect to an empty db so provide a dump path
def create_db(dump_path=None, db_name=None):
    if not db_name:
        db_name = "db_" + timestamp()
    # Overwrite db.pgdump with dumpfile from host
    if dump_path:
        host_mount = "-v %s:/db/db.pgdump" % dump_path
    else:
        host_mount = ""
    cmd = CREATE_DB % (host_mount, db_name)
    log = os.path.join("/tmp", db_name + ".log")
    print("Creating db and logging output to %s" % log)
    with open(log, "w") as f:
        f.write(cmd)
        f.write("\n\n")
        p = subprocess.run(shlex.split(cmd), stderr=f, stdout=f)
    assert p.returncode == 0
    return db_name
