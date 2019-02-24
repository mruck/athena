#!/usr/bin/env python3

# Orchestration framework to spawn containerized target server app
# and client fuzzer on osx. A CLI for the fuzz duo class.
#
# TODO:
# If a new fuzz duo is spawned, the db is created from the db container.
# Otherwise, the db snapshot is restored from the client container
# (cause that functionality already exists).

import argparse

import db
import coverage
import fuzz_duo
import server


def repro(args):
    """
    Repro behavior of fuzz duo with given ID.
    Recycle ports and dbs.
    """
    # Get config and check that duo id exists
    config = fuzz_duo.get_config(args.id)
    if config is None:
        return
    # Print logs then we are done
    if args.logs:
        fuzz_duo.show_logs(args.id, config=config)
        return
    # Initialize fuzz duo pair from config
    duo = fuzz_duo.FuzzDuo.from_config(config)

    if args.mount_discourse:
        duo.server.extra_mounts = [("/tmp/discourse-fork", "/discourse-fork")]

    # Spawn only client
    if args.client:
        # Remove the client container
        duo.client.rm_container()
        duo.run_client(
            background=False,
            route=args.route,
            any_route=args.any_route,
            shell=args.shell,
            stop_after_har=args.stop_after_har,
            stop_after_all_routes=args.stop_after_all_routes,
        )
        return

    # Spawn only server in foreground
    if args.server:
        duo.server.rm_container()
        duo.server.run(background=False, shell=args.shell)
        return

    # Spawn both a client and server
    # The user wants to restart the server or the server is down.
    if args.restart_server or not fuzz_duo.check_server_alive(args.id, config=config):
        duo.server.rm_container()
        duo.server.run()

    # Remove the client container
    duo.client.rm_container()
    # Should client run in foreground or background
    background = not args.foreground
    duo.run_client(
        background=background,
        route=args.route,
        any_route=args.any_route,
        stop_after_har=args.stop_after_har,
        stop_after_all_routes=args.stop_after_all_routes,
    )


def spawn_fuzz_duo(args):
    """
    Spawn new containerized server-client pairs.
    Allocate ports and dbs.
    We do not kill the servers so that we can
    1) reconnect to them for reproing
    2) cut down on start up time (it takes about 20s for a server to start up and to
    restore the db)
    """
    duo = fuzz_duo.FuzzDuo.new()
    duo.save()


def run_parser():
    # create the top-level parser
    parser = argparse.ArgumentParser(
        description="Welcome to Bananas in Pajamas." " Lets fuzz some web apps."
    )
    subparsers = parser.add_subparsers(help="sub-command help")

    parser_server = subparsers.add_parser(
        "server", description="Spawn server in foreground", help="server -h"
    )
    parser_server.add_argument(
        "--shell", action="store_true", help="Spawn a shell into the container"
    )
    parser_server.set_defaults(func=server)

    parser_fuzz_duo = subparsers.add_parser(
        "fuzz_duo", description="Spawn a client server fuzz duo", help="fuzz_duo -h"
    )
    parser_fuzz_duo.add_argument(
        "instances", type=int, help="Number of client server instances to spawn"
    )
    parser_fuzz_duo.add_argument(
        "--foreground", action="store_true", help="Run client in foreground"
    )
    parser_fuzz_duo.add_argument(
        "--mount-discourse",
        action="store_true",
        help="Mount the discourse fork from /tmp/discourse-fork",
    )
    parser_fuzz_duo.set_defaults(func=spawn_fuzz_duo)

    parser_repro = subparsers.add_parser(
        "repro",
        description="Repro behavior of fuzz duo given a fuzz id",
        help="fuzz_duo -h",
    )
    parser_repro.add_argument("id", type=int, help="Id of fuzz duo")
    parser_repro.add_argument("--logs", action="store_true", help="Show logs for ID")
    parser_repro.add_argument(
        "--stop_after_har",
        action="store_true",
        help="Stop fuzzer after done mutating har requests",
    )
    parser_repro.add_argument(
        "--stop_after_all_routes",
        action="store_true",
        help="Stop fuzzer after we've hit all routes at least once",
    )
    parser_repro.add_argument(
        "--mount-discourse",
        action="store_true",
        help="Mount the discourse fork from /tmp/discourse-fork",
    )
    parser_repro.add_argument(
        "--restart_server", action="store_true", help="Restart server"
    )
    parser_repro.add_argument(
        "--shell", action="store_true", help="Spawn a shell into the container"
    )
    parser_repro.add_argument(
        "--foreground", action="store_true", help="Spawn client in the foreground"
    )
    group1 = parser_repro.add_mutually_exclusive_group()
    group1.add_argument(
        "--client", action="store_true", help="Run client in foreground"
    )
    group1.add_argument(
        "--server", action="store_true", help="Run server in foreground"
    )
    group2 = parser_repro.add_mutually_exclusive_group()
    group2.add_argument(
        "--route",
        help="Rerun ROUTE This restores the db and RNG state to the snapshot before "
        "ROUTE was run. Default behavior runs all routes.",
    )
    group2.add_argument(
        "--any_route",
        action="store_true",
        help="Run a random route. Useful for debugging. Does not restore db/RNG state. "
        "Default behavior runs all routes.",
    )
    parser_repro.set_defaults(func=repro)
    args = parser.parse_args()
    return args


def main():
    """
    Run discourse
    """
    args = run_parser()
    args.func(args)


if __name__ == "__main__":
    main()
