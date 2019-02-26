#!/bin/python3
import logging
import os
import sys

from fuzzer.database.postgres import connect_to_db, load_db_dump
from fuzzer.database.db import find_text_cols, inject_xss_payload

# Logger for general debugging
logger = logging.getLogger("xss.log")

try:
    with open("./cookie", "r") as cookie_file:
        COOKIE = cookie_file.read().strip()
except:
    raise Exception("Cookie not found!")
    sys.exit(1)


def load_routes():
    target_app_path = os.getenv("TARGET_APP_PATH")
    if target_app_path is None:
        raise Exception("Specify TARGET_APP_PATH env var")
        sys.exit(1)

    if target_app_path.endswith("/"):
        target_app_path = target_app_path[:-1]

    route_path = os.path.join(target_app_path, ".meta", "final_routes.json")
    if not os.path.isfile(route_path):
        logger.error("Missing final_routes.json to load routes from")


def main():
    # Connect to db
    load_db_dump()
    connect_to_db()
    # Identify text cols and inject xss
    text_cols = find_text_cols()
    inject_xss_payload(text_cols)
    # Spin up server
    # Begin headless browsing
    # Clean up


if __name__ == "__main__":
    main()
