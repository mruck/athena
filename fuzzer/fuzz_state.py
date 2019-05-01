#! /usr/local/bin/python3

import http.cookiejar
import os.path
import pathlib
import pickle
import random
import shutil

from fuzzer.database.db import clear_rails_connections
import fuzzer.lib.util as util


class FuzzState(object):
    def __init__(self, postgres, db_name):
        self.cookies = http.cookiejar.LWPCookieJar()
        self.postgres = postgres
        self.db_name = db_name

    def save(self, dest):
        self._ensure_path_exists(dest)

        cookies_file = FuzzState.cookies_file(dest)
        self.cookies.save(cookies_file)

        fh = open(FuzzState.rng_file(dest), "wb")
        random_state = random.getstate()
        pickle.dump(random_state, fh)
        fh.close()

        db_fname = FuzzState.db_file(dest)
        self.postgres.snapshot(self.db_name, db_fname)
        return self

    def load(self, src):
        dirname = FuzzState.dirname(src)
        if not os.path.exists(dirname):
            raise Exception(
                "Failed to find the fuzz state directory: {}".format(dirname)
            )

        cookies_file = FuzzState.cookies_file(src)
        self.cookies.revert(cookies_file)

        fh = open(FuzzState.rng_file(src), "rb")
        random.setstate(pickle.load(fh))
        fh.close()

        # There was a snapshot provided but we already loaded db so don't do it again,
        # just use the other state
        if not os.getenv("SKIP_DB_LOAD"):
            db_fname = FuzzState.db_file(src)
            clear_rails_connections(port=util.target_app_port())
            self.postgres.load_snapshot(self.db_name, db_fname)
        return self

    def delete(self, path):
        path = FuzzState.dirname(path)
        shutil.rmtree(path, ignore_errors=True)

    def _ensure_path_exists(self, dest):
        path = FuzzState.dirname(dest)
        pathlib.Path(path).mkdir(parents=True, exist_ok=True)

    @classmethod
    def dirname(cls, dest):
        return os.path.join(dest, "fuzz_state")

    @classmethod
    def cookies_file(cls, directory):
        return os.path.join(cls.dirname(directory), "cookiejar")

    @classmethod
    def rng_file(cls, directory):
        return os.path.join(cls.dirname(directory), "rng")

    @classmethod
    def db_file(cls, directory):
        return os.path.join(cls.dirname(directory), "dbdump")
