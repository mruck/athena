import logging
import os
import psycopg2
import subprocess
import sys
import time

import fuzzer.lib.util as util

logger = logging.getLogger("debug")


def create_db(dumpfile, name=None):
    """
    Create and populate db
    """
    if not name:
        name = "db_" + util.timestamp()
    p = subprocess.run(["createdb", "-T", "template0", name])
    assert p.returncode == 0
    with open(dumpfile, "r") as f:
        p = subprocess.run(
            ["psql", name],
            stdin=f,
            stdout=open(os.devnull, "w"),
            stderr=open(os.devnull, "w"),
        )
        assert p.returncode == 0
    return name


def load_db_dump(db_name, filepath):
    try:
        # If dropping fails, kill all attached pids and try again
        for i in range(0, 2):
            dropdb_process = subprocess.run(
                ["dropdb", db_name], stdout=subprocess.PIPE, stderr=subprocess.PIPE
            )
            # Success
            if dropdb_process.returncode == 0:
                break
            output = dropdb_process.stderr.decode()
            # DB was already dropped
            if 'database "{}" does not exist'.format(db_name) in output:
                break
            # Kill pids attached to db and try again
            if i == 0:
                print("Uh oh can't dropdb.  Killing connected pids")
                subprocess.run(
                    ["./kill.sh", db_name],
                    stdout=subprocess.PIPE,
                    stderr=subprocess.PIPE,
                )
                # Need to sleep for a sec before dropping db again because otherwise
                # we get a "db in recovery mode" error
                time.sleep(1)
            else:
                logger.error("couldn't drop '{}' database\n".format(db_name))
                logger.error("\t" + "\n\t".join(output.split("\n")))
                sys.exit(1)
        create_db(filepath, name=db_name)
        logger.info("Done.")
    except FileNotFoundError:
        logger.error("Error: couldn't find database dump file {}".format(filepath))
        sys.exit(1)
    except Exception as e:
        logger.error("Unknown error when trying to load db dump: {}".format(e))
        sys.exit(1)


def connect_to_db(db_name):
    global cursor
    global conn

    # conn_str = "dbname='%s' user='root' host='localhost' password='mysecretpassword'" % db_name
    conn_str = (
        "dbname='%s' user='root' host='localhost' password='mysecretpassword'" % db_name
    )
    conn = psycopg2.connect(conn_str)
    cursor = conn.cursor()


def run_query(query, return_all_records=False, can_fail=False, first_time=True):
    global cursor
    global conn
    try:
        cursor.execute(query)
    except psycopg2.ProgrammingError as e:
        logger.error(e)
        logger.error("query ran: %s" % query)
        exn = "Programming Error" + "\nquery_ran: %s" % query
        raise Exception(exn)
    except psycopg2.DataError as e:
        logger.error("query ran: %s" % query)
        raise e
    except psycopg2.InternalError as e:
        # This seems to be a psycopg bug rather than user bug (although I'm
        # only 90% confident of that).  Let's rollback and try again.
        if first_time:
            conn.rollback()
            run_query(
                query,
                return_all_records=return_all_records,
                can_fail=can_fail,
                first_time=False,
            )
        else:
            logger.error(e)
            logger.error("query ran: %s" % query)
            exn = "Internal error" + "\nquery_ran: %s" % query
            raise Exception(exn)

    # return a single record
    if return_all_records:
        return cursor.fetchall()
    else:
        rows = cursor.fetchone()
        cols = [desc[0] for desc in cursor.description]
        try:
            # Zip col names and row
            return {k: v for k, v in zip(cols, rows)}
        except TypeError:
            # The row returned is empty
            return None


def snapshot_db(result_path, db_name, filename):
    with open(os.path.join(result_path, "db_snapshots", filename), "w") as dumpfile:
        p = subprocess.run(
            ["pg_dump", db_name], stdout=dumpfile, stderr=subprocess.PIPE
        )
        if p.returncode != 0:
            print(p.stderr)
            assert False
