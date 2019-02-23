#! /usr/local/bin/python3

import psycopg2
import random
import tempfile
import urllib.request

import fuzzer.fuzz_state as fuzz_state
import fuzzer.postgres2 as postgres


def _req_res_pair():
    request = urllib.request.Request("https://www.google.com")
    response = urllib.request.urlopen(request)
    return request, response


def do_test():
    state_location = tempfile.mkdtemp()

    pg = postgres.Postgres(hostname="localhost")
    db_name = pg.create_db()
    state = fuzz_state.FuzzState(pg, db_name)

    # Set postgres state
    conn_str = "host=localhost dbname='{}' user='root'".format(db_name)
    conn = psycopg2.connect(conn_str)
    cur = conn.cursor()
    cur.execute("CREATE TABLE test (id serial PRIMARY KEY, num integer, data varchar);")
    cur.execute("INSERT INTO test (id, num, data) VALUES (1, 2, 'asd');")
    conn.commit()
    cur.close()
    conn.close()

    # Set cookie state
    request, response = _req_res_pair()
    state.cookies.extract_cookies(response, request)

    # Save the state.
    state.save(state_location)
    first_rand = random.randint(0, 1000)
    cookies_length = len(state.cookies)

    # Create new state and load it from fs.
    db_name = pg.create_db()
    state = fuzz_state.FuzzState(pg, db_name)
    state.load(state_location)

    # Check the state again to make sure it matches.
    second_rand = random.randint(0, 1000)
    assert first_rand == second_rand
    assert len(state.cookies) == cookies_length

    conn_str = "host=localhost dbname='{}' user='root'".format(db_name)
    conn = psycopg2.connect(conn_str)
    cur = conn.cursor()
    cur.execute("SELECT * FROM test;")
    key, num, data = cur.fetchone()
    assert key == 1
    assert num == 2
    assert data == "asd"
    conn.commit()
    cur.close()
    conn.close()


if __name__ == "__main__":
    do_test()
    print("PASSED")
