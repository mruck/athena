import json
import pprint
import shutil
import subprocess

import fuzzer.client as client
import fuzzer.database.db as db
import fuzzer.server as server
import fuzzer.lib.util as util


def load():
    """
    Load file with global client-server pairs
    """
    try:
        contents = open("/tmp/global_ids", "r").read()
        return json.loads(contents)
    except FileNotFoundError:
        return []


def dump_global_id_file(global_ids):
    """
    Write global id metadata file to disc
    """
    with open("/tmp/global_ids", "w") as f:
        json.dump(global_ids, f)


def update_config(config):
    """
    Update the existing config
    """
    # Load global id file
    metadata = load()
    # Figure out the index of our config
    index = next(
        (index for (index, d) in enumerate(metadata) if d["port"] == config["port"]),
        None,
    )
    if index is None:
        print("Config not found!")
        assert False
    # Update that index
    metadata[index].update(config)
    dump_global_id_file(metadata)


def get_config(id):
    """
    Retrieve the config from the global id
    file for a given client server pair
    """
    # List of dicts with metadata about each client-server pair
    metadata = load()
    # Search for our id
    index = next((index for (index, d) in enumerate(metadata) if d["port"] == id), None)
    if index is None:
        print("ID %d does not exist!" % id)
        assert False
    config = metadata[index]
    return config


def check_server_alive(id, config=None):
    """
    Ensure server is still running by checking whether or not container has exited.
    """
    if config is None:
        config = get_config(id)

    cmd = "docker ps | tail -n+2 | grep %s" % config["server_name"]
    p = subprocess.run(cmd, shell=True)
    return p.returncode == 0


def show_logs(id, config=None):
    if config is None:
        config = get_config(id)
    pprint.PrettyPrinter(indent=4).pprint(config)


class FuzzDuo(object):
    def __init__(self, client, server, config=None):
        self.client = client
        self.server = server
        self.config = config

    def get_id(self):
        """
        Return unique id assigned to fuzz duo.  Use the port for now.
        """
        return self.client.port

    @classmethod
    def from_config(cls, config):
        """
        Initalize fuzz duo pair from config
        """
        s = server.Server(config["port"], config["db"], config["results_path"])
        c = client.Client(
            config["port"],
            config["db"],
            config["results_path"],
            fuzzer_number=config["fuzzer_number"],
            instances=config["instances"],
        )
        return cls(c, s, config=config)

    @classmethod
    def new(cls, client_background=True, server_background=True):
        """
        Allocate port, db and results directory.
        Spin up a client and server.
        """
        port = util.get_open_port()
        # Rely on createdb containerized utility to get a db
        db_name = db.create_db()
        results_path = util.mk_results_path()

        s = server.Server(port, db_name, results_path)
        c = client.Client(port, db_name, results_path)

        return cls(c, s)

    @classmethod
    def run(
        cls,
        client_background=True,
        server_background=True,
        fuzzer_number=0,
        instances=1,
        extra_mounts=None,
    ):
        """
        Initialize new client and server, allocating port/db, then run
        """

        port = util.get_open_port()
        # Rely on createdb containerized utility to get a db
        db_name = db.create_db()
        results_path = util.mk_results_path()

        s = server.Server(port, db_name, results_path, extra_mounts=extra_mounts)
        c = client.Client(
            port,
            db_name,
            results_path,
            fuzzer_number=fuzzer_number,
            instances=instances,
        )
        duo = cls(c, s)
        duo.run_server(background=server_background)
        duo.run_client(background=client_background)
        return duo

    def save(self):
        """
        Store client-server metadata to global id file
        """
        global_dict = load()
        fuzz_duo_dict = {
            "port": self.client.port,
            "db": self.client.db,
            "results_path": self.client.results_path,
            "fuzzer_number": self.client.fuzzer_number,
            "snapshot": None,
            "instances": self.client.instances,
            # server container name
            "server_name": self.server.name,
            "client_": self.client.name,
        }
        global_dict.append(fuzz_duo_dict)
        with open("/tmp/global_ids", "w") as f:
            json.dump(global_dict, f)

        duo_id = self.client.port
        print("Saved with id: %d" % duo_id)
        return duo_id

    def run_server(self, background=True, new_server=True, shell=False):
        """
        Wrapper for server class.  Ensures we store metadata first.
        """
        if new_server:
            duo_id = self.save()
        self.server.run(background=background, shell=shell)
        return duo_id

    def run_client(
        self,
        background=True,
        route=None,
        any_route=None,
        load_db=False,
        shell=False,
        stop_after_har=False,
        stop_after_all_routes=False,
    ):
        """
        Wrapper for client class. Do any necessary setup including
        saving the results directory if the user wants to restore a
        snapshot.
        """
        # User wants to run a specific route
        if route:
            # Allocate a snapshot directory and copy results_path there because
            # our call to the fuzzer will clobber the results_path
            if self.config["snapshot"] is None:
                self.config["snapshot"] = "/tmp/snapshot_" + util.timestamp()
                shutil.copytree(self.config["results_path"], self.config["snapshot"])
                # Dump updated config
                update_config(self.config)
                self.client.run(
                    background=background,
                    route=route,
                    any_route=any_route,
                    snapshot=self.config["snapshot"],
                    load_db=load_db,
                    stop_after_har=stop_after_har,
                    stop_after_all_routes=stop_after_all_routes,
                )
        else:
            self.client.run(
                background=background,
                route=route,
                any_route=any_route,
                load_db=load_db,
                shell=shell,
                stop_after_har=stop_after_har,
                stop_after_all_routes=stop_after_all_routes,
            )
