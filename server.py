import os
import shlex
import stat
import subprocess

# Requires postgres and redis container to be running
DISCOURSE_SERVER_CMD = (
    'docker run %(interactive)s -p %(port)s:%(port)s --name=%(name)s --rm -e "DISCOURSE_DEV_DB=%(db)s" -v '
    "%(results_path)s:%(results_path)s %(shell)s "
    "%(extra_mount_flags)s "
    '--volumes-from my-postgres -e "RESULTS_PATH=%(results_path)s" '
    '-e "PORT=%(port)s"  target-server'
)


class Server(object):
    def __init__(self, port, db, results_path, extra_mounts=None):
        self.port = port
        self.db = db
        self.results_path = results_path
        # Name of docker container with server
        self.name = "server_%d" % port
        self.background = True
        self.extra_mounts = extra_mounts

    def _get_mount_flags(self):
        if self.extra_mounts is None or len(self.extra_mounts) == 0:
            return ""
        flag = ""
        for local, dest in self.extra_mounts:
            flag += "-v {}:{}:delegated".format(local, dest)
        return flag

    def build_run_cmd(self, shell=False):
        if shell:
            shell = "--entrypoint=/bin/bash"
        else:
            shell = ""
        cmd = DISCOURSE_SERVER_CMD % {
            "db": self.db,
            "port": str(self.port),
            "name": self.name,
            "results_path": self.results_path,
            "shell": shell,
            "interactive": "",
            "extra_mount_flags": self._get_mount_flags(),
        }
        interactive_cmd = DISCOURSE_SERVER_CMD % {
            "db": self.db,
            "port": str(self.port),
            "name": self.name,
            "results_path": self.results_path,
            "shell": shell,
            "interactive": "-it",
            "extra_mount_flags": self._get_mount_flags(),
        }
        run_script = os.path.join(self.results_path, "run_server.sh")
        with open(run_script, "w") as f:
            f.write(interactive_cmd)
        st = os.stat(run_script)
        os.chmod(run_script, st.st_mode | stat.S_IEXEC)
        print("Run command available at: %s" % run_script)
        if shell:
            return interactive_cmd
        else:
            return cmd

    def run(self, background=True, shell=False):
        print("\nStarting %s..." % self.name)
        cmd = self.build_run_cmd(shell=shell)
        if background:
            # Run in background and log stdout to file
            log = os.path.join(self.results_path, "server.stdout")
            print("stdout available at: %s" % log)
            fp = open(log, "w")
            proc = subprocess.Popen(
                shlex.split(cmd),
                stderr=fp,
                stdout=fp,
                stdin=subprocess.PIPE,
                shell=False,
            )
            # Wait for server to spin up
            try:
                out, err = proc.communicate(timeout=20)
                fp.close()
                print("Server exited!")
                print(open(log, "r").read())
                exit(1)
            except subprocess.TimeoutExpired:
                pass
        else:
            print("results_path: %s" % self.results_path)
            print("db: %s" % self.db)
            print("port: %d\n" % self.port)
            subprocess.run(cmd, shell=True)

    def rm_container(self):
        cmd = "docker rm -f %s" % self.name
        subprocess.run(shlex.split(cmd))
