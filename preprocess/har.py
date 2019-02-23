#!/usr/bin/env python3
#
# Utility for replaying HAR entries against a chosen server.
#
import json
import http.cookiejar
import socket
import urllib.request
import urllib.parse


def patch_url(entry, netloc):
    url = entry["request"]["url"]
    if netloc is not None:
        url = urllib.parse.urlparse(url)._replace(netloc=netloc).geturl()
    return url


class HarReplayer:
    def __init__(self, netloc):
        self.netloc = netloc
        self.cookie_jar = http.cookiejar.CookieJar()

    def replay_har(self, entry):
        expected_status = entry["response"]["status"]
        if expected_status == 0:
            return  # e.g. this can happen if an ad blocker stops the request.

        url = entry["request"]["url"]
        # TODO: We get a weird urlError on login
        if "login" in url:
            return
        # We accidentally browsed to a different site
        if "localhost" not in url and "127.0.0.1" not in url:
            return
        if self.netloc is not None:
            url = (
                urllib.parse.urlparse(url)
                ._replace(netloc=self.netloc, scheme="http")
                .geturl()
            )
        # print("\n\n\n\n")
        print(url)
        headers = get_har_headers(entry)

        params = get_body_params(entry)
        data = urllib.parse.urlencode(params).encode("utf-8")
        request = urllib.request.Request(
            url, method=entry["request"]["method"], headers=headers, data=data
        )
        self.cookie_jar.add_cookie_header(request)

        response = None
        try:
            response = urllib.request.urlopen(request, timeout=5)
        except socket.timeout:
            with open("/tmp/har_timeouts", "a") as f:
                f.write(url)
                f.write("\n")
            return None
        except urllib.error.HTTPError as e:
            if e.code != expected_status:
                log_status(entry, e.code)
            return e.code
        self.cookie_jar.extract_cookies(response, request)
        if response.status != expected_status:
            log_status(entry, response.status)
        return response.status


def log_status(entry, actual_status):
    if "mini-profiler" in entry["request"]["url"]:
        return
    with open("/tmp/replay_har.errors", "a") as f:
        entry["replay_status"] = actual_status
        json.dump(entry, f)
        f.write("\n")


def get_body_params(entry):
    post_data = entry.get("request", {}).get("postData", {})
    if "params" not in post_data:
        return {}
    params = {
        urllib.parse.unquote(param["name"]): urllib.parse.unquote(param["value"])
        for param in post_data["params"]
    }
    return params


def get_har_headers(entry):
    headers = {
        header["name"]: header["value"]
        for header in entry["request"]["headers"]
        if not header["name"].startswith(":")
    }
    if "Cookie" in headers:
        del headers["Cookie"]
    return headers


def main(argv):
    import json

    if len(argv) < 3:
        print("Usage:", argv[0], "hostname:port", "har_file_paths...")
        return 64

    netloc = argv[1] or None
    replayer = HarReplayer(netloc)
    for i in range(2, len(argv)):
        har = json.load(open(argv[i]))
        for entry in har["log"]["entries"]:
            replayer.replay_har(entry)
    return 0


if __name__ == "__main__":
    import sys

    sys.exit(main(sys.argv))
