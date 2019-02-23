#!/usr/bin/env python2
"""
Very simple HTTP server in python.
Usage::
    python dispatcher.py <?dispatcher items> <?port>
    ie: dispatcher.py ~/discourse_state/routes.json 8080
Send a GET request::
    curl http://localhost:8080
Code borrowed from: https://gist.github.com/bradmontgomery/2219997)
"""

from BaseHTTPServer import BaseHTTPRequestHandler, HTTPServer
from threading import Lock
import json

queue_lock = Lock()
queue = []


class S(BaseHTTPRequestHandler):
    def _set_headers(self, code=200):
        self.send_response(code)
        self.send_header("Content-type", "application/json")
        self.end_headers()

    def do_GET(self):
        with queue_lock:
            if len(queue) == 0:
                self._set_headers(404)
                self.wfile.write("No more elements")
                return
            self._set_headers(200)
            el = queue.pop(0)
            self.wfile.write(el)

    def do_HEAD(self):
        with queue_lock:
            if len(queue) == 0:
                self._set_headers(404)
                return
            self._set_headers()


def run(server_class=HTTPServer, handler_class=S, port=8080):
    server_address = ("", port)
    httpd = server_class(server_address, handler_class)
    print("Starting httpd...")
    httpd.serve_forever()


def load_queue(fname):
    contents = open(fname, "r").read()
    return json.loads(contents)


if __name__ == "__main__":
    from sys import argv

    port = 8080
    payloads_file = "dispatcher_inputs.json"
    if len(argv) >= 2:
        payloads_file = argv[1]
    if len(argv) >= 3:
        port = int(argv[2])

    queue = load_queue(payloads_file)
    run(port=port)
