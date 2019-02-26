import logging
import os
import subprocess
import time

logger = logging.getLogger("debug")


class Postgres(object):
    def __init__(self, hostname=None, user="root"):
        self.hostname = hostname
        self.user = user

    def _extend_args(self, args):
        if self.hostname:
            args.extend(["-h", self.hostname])
        if self.user:
            args.extend(["-U", self.user])
        return args

    def create_db(self, db_name=None):
        if not db_name:
            db_name = "db_{}".format(int(time.time() * 1000))

        args = self._extend_args(["createdb", "-T", "template0"])
        args.append(db_name)

        p = subprocess.run(args)
        assert p.returncode == 0
        return db_name

    def ensure_db_dropped(self, db_name):
        for i in range(2):
            if self.try_db_drop(db_name):
                return

            self.kill_db_conns(db_name)

            # Need to sleep for a sec before dropping db again because otherwise
            # we get a "db in recovery mode" error
            time.sleep(1)

        raise Exception("Failed to drop database {}".format(db_name))

    def try_db_drop(self, db_name):
        logger.info("dropping db {}".format(db_name))
        args = self._extend_args(["dropdb"])
        args.append(db_name)

        dropdb_process = subprocess.run(
            args, stdout=subprocess.PIPE, stderr=subprocess.PIPE
        )
        # Success
        if dropdb_process.returncode == 0:
            return True

        output = dropdb_process.stderr.decode()
        # DB was already dropped
        if 'database "{}" does not exist'.format(db_name) in output:
            return True

        logger.error("couldn't drop '{}' database\n".format(db_name))
        logger.error("\t" + "\n\t".join(output.split("\n")))
        return False

    def kill_db_conns(self, db_name):
        logger.info("killing db conns to {}".format(db_name))
        subprocess.run(
            ["./kill.sh", db_name], stdout=subprocess.PIPE, stderr=subprocess.PIPE
        )

    def snapshot(self, db_name, filepath):
        with open(filepath, "w") as dumpfile:
            args = self._extend_args(["pg_dump"])
            args.append(db_name)
            p = subprocess.run(args, stdout=dumpfile, stderr=subprocess.PIPE)
            if p.returncode != 0:
                print(p.stderr)
                assert False

    def load_snapshot(self, db_name, filepath):
        self.ensure_db_dropped(db_name)
        self.create_db(db_name)

        with open(filepath, "r") as f:
            args = self._extend_args(["psql"])
            args.append(db_name)
            p = subprocess.run(
                args,
                stdin=f,
                stdout=open(os.devnull, "w"),
                stderr=open(os.devnull, "w"),
            )
            assert p.returncode == 0
