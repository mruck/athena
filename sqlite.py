#!/bin/bash/python3

# Connect to database and execute queries
import sqlite3

######################################## DB setup
# Hardcode for now. Assume each model exists and is prepopulated
# Note that names are canonicalized (pluralized and lowercase)
POP_QUERY = "SELECT * from %s limit 1"
DATABASE = "/vagrant/my-test-app/db/development.sqlite3"


def gen_pop(table):
    return POP_QUERY % table


def dict_factory(cursor, row):
    d = {}
    for idx, col in enumerate(cursor.description):
        d[col[0]] = row[idx]
    return d


# Connect to database
def connect_to_db():
    global cursor
    conn = sqlite3.connect(DATABASE)
    conn.row_factory = dict_factory
    cursor = conn.cursor()


# Execute query
def run_query(query):
    global cursor
    cursor.execute(query)
    return cursor.fetchone()


# Retrieve a record from table
def pop_from_table(table):
    query = gen_pop(table)
    return run_query(query)
