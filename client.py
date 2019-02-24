import os
import shlex
import stat
import subprocess


# Hacky way to get relative path to discourse-fork/state
PWD = os.path.dirname(os.path.realpath(__file__))
STATE = os.path.join(PWD, "..", "discourse-fork", "state")

# TODO: Replace popen with docker python library
RUN_DOCKER_CLIENT = (
    "docker run %(interactive)s --name=%(name)s --net=host --volumes-from my-postgres -v "
    '%(state)s:/state -e "RESULTS_PATH=%(results_path)s" -e "DB_NAME=%(db)s" '
    '-e "FUZZER_NUMBER=%(fuzzer_number)s" -e "INSTANCES=%(instances)s" '
    '-e "PORT=%(port)s" -v %(results_path)s:%(results_path)s -v %(pwd)s:/client '
    "%(snapshot_mount)s %(snapshot_env_var)s %(route_env_var)s %(any_route_env_var)s "
    "%(load_db_env_var)s %(stop_after_har_env_var)s %(shell)s fuzzer-client"
)

TIMEOUT = 60 * 25
TARGET_APP = "/discourse-fork"


class Client(object):
    def __init__(
        self, port, db, results_path, snapshot=None, fuzzer_number=None, instances=None
    ):
        self.db = db
        self.port = port
        self.results_path = results_path
        self.fuzzer_number = fuzzer_number or 0
        self.instances = instances or 1
        self.proc = None
        # stdout/stderr filename if spawned in background
        self.log = None
        self.name = "client_" + str(port)
        self.background = True

    def run(
        self,
        snapshot=None,
        route=None,
        any_route=False,
        background=True,
        load_db=False,
        shell=False,
        stop_after_har=False,
    ):
        if snapshot:
            # Mount the path to the snapshot
            snapshot_mount = "-v %s:%s" % (snapshot, snapshot)
            snapshot_env_var = "-e 'SNAPSHOT=--snapshot %s'" % snapshot
        else:
            snapshot_mount = ""
            snapshot_env_var = ""
        if route:
            route_env_var = "-e 'ROUTE=--route %s'" % route
        else:
            route_env_var = ""
        if any_route:
            any_route_env_var = "-e 'ANY_ROUTE=--any-route'"
        else:
            any_route_env_var = ""
        if load_db:
            load_db_env_var = "-e 'LOAD_DB=--load_db'"
        else:
            load_db_env_var = ""
        if shell:
            shell = "--entrypoint=bash"
        else:
            shell = ""
        if stop_after_har:
            stop_after_har_env_var = "-e 'STOP_AFTER_HAR=--stop_after_har'"
        else:
            stop_after_har_env_var = ""

        # Interactive cmd
        docker_cmd_interactive = RUN_DOCKER_CLIENT % {
            "results_path": self.results_path,
            "state": STATE,
            "db": self.db,
            "port": str(self.port),
            "fuzzer_number": self.fuzzer_number,
            "instances": self.instances,
            "snapshot_mount": snapshot_mount,
            "snapshot_env_var": snapshot_env_var,
            "stop_after_har_env_var": stop_after_har_env_var,
            "route_env_var": route_env_var,
            "pwd": PWD,
            "any_route_env_var": any_route_env_var,
            "load_db_env_var": load_db_env_var,
            "name": self.name,
            "shell": shell,
            "interactive": "-it",
        }
        run_docker = os.path.join(self.results_path, "run_client_docker.sh")
        with open(run_docker, "w") as f:
            f.write(docker_cmd_interactive)
        st = os.stat(run_docker)
        os.chmod(run_docker, st.st_mode | stat.S_IEXEC)
        print("docker run command for client at %s" % run_docker)

        # Non interactive cmd
        docker_cmd = RUN_DOCKER_CLIENT % {
            "results_path": self.results_path,
            "state": STATE,
            "db": self.db,
            "port": str(self.port),
            "fuzzer_number": self.fuzzer_number,
            "instances": self.instances,
            "name": self.name,
            "snapshot_mount": snapshot_mount,
            "snapshot_env_var": snapshot_env_var,
            "stop_after_har_env_var": stop_after_har_env_var,
            "route_env_var": route_env_var,
            "route_env_var": route_env_var,
            "pwd": PWD,
            "any_route_env_var": any_route_env_var,
            "load_db_env_var": load_db_env_var,
            "shell": shell,
            "interactive": "",
        }

        if not background or shell:
            self.background = False
            # We use `pwd` in our docker cmd so we need shell=True? unclear...
            subprocess.run(docker_cmd_interactive, shell=True)
        else:
            log = os.path.join(self.results_path, "client.stdout")
            fp = open(log, "w")
            self.proc = subprocess.Popen(
                shlex.split(docker_cmd), stdout=fp, stderr=fp, stdin=subprocess.PIPE
            )
            print("Started fuzzer %d/%d" % (self.fuzzer_number, self.instances))
            print("Logs can be viewed at %s" % log)

    def check_errors(self):
        """
        Read stout/stderr logged by client
        """
        if self.log:
            return open(self.log, "r").read()
        return None

    def rm_container(self):
        cmd = "docker rm -f %s" % self.name
        subprocess.run(shlex.split(cmd))

    def check(self):
        """
        Check on the client we spawned
        """
        try:
            self.proc.communicate(timeout=TIMEOUT)
        except subprocess.TimeoutExpired:
            print("Error: Fuzzer timed out!")
            self.proc.kill()
            self.proc.communicate()
            print(self.check_errors())
        if self.proc.returncode != 0:
            print("Fuzzer process returned non-zero exit code")
            print(self.check_errors())
        # Clean up
        self.rm_container()
