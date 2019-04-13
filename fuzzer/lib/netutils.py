import datetime
import json
import os
import socket
import time
import urllib.request
import urllib.parse


CONTENT_TYPE_FORM = "application/x-www-form-urlencoded; charset=UTF-8"
CONTENT_TYPE_JSON = "application/json"


def json_hook(obj):
    if isinstance(obj, (datetime.date, datetime.datetime)):
        return obj.isoformat()


def target_hostname():
    return os.getenv("TARGET_HOSTNAME", "localhost")


class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        return None


class Connection:
    def __init__(self, old_cookie_jar, timeout=None):
        self.cookie_jar = old_cookie_jar
        self.timeout = timeout or 60
        # Don't follow redirects
        opener = urllib.request.build_opener(NoRedirect)
        urllib.request.install_opener(opener)

    def log_params(self, url, body_params=None, query_params=None):
        print(url)
        if len(body_params) > 0:
            print(body_params)
        if len(query_params) > 0:
            print(query_params)

    def is_alive(self, port):
        ret_code = None
        while ret_code is None:
            time.sleep(5)
            print("Polling...")
            try:
                ret_code = self.send_request(
                    "http://localhost:%s/rails/info/pluralizations" % port, "GET"
                )
            except urllib.error.URLError:
                pass
        # if ret_code != 200:
        #    print("Error: ret code is %d" % ret_code)
        #    assert False

    def _format_body_params(self, body, headers):
        # The HAR sent JSON so lets stick to that
        if "Content-Type" in headers and headers["Content-Type"] == CONTENT_TYPE_JSON:
            print("Sending JSON")
            return json.dumps(body, default=json_hook)
        # Otherwise default to url encoding
        headers["Content-Type"] = CONTENT_TYPE_FORM
        return urllib.parse.urlencode(body, doseq=True)

    def _clean_headers(self, headers):
        headers.pop("Content-Length", None)
        headers.pop("Cookies", None)
        return headers

    def _build_request(
        self, url, verb, body_params=None, headers=None, query_params=None
    ):
        if verb == "GET":
            assert len(body_params) == 0

        self.log_params(url, body_params=body_params, query_params=query_params)

        if body_params:
            body_params = self._format_body_params(body_params, headers).encode("utf-8")

        if query_params:
            url += "?" + urllib.parse.urlencode(query_params, doseq=True)

        headers = self._clean_headers(headers)

        request = urllib.request.Request(
            url, data=body_params, headers=headers, method=verb
        )
        return request

    # query_params are of the form {'p': [1], 'q': [2, 3]}
    def send_request(
        self, url, verb, body_params=None, headers=None, query_params=None
    ):
        body_params = body_params or {}
        query_params = query_params or {}
        headers = headers or {}

        request = self._build_request(
            url,
            verb,
            body_params=body_params,
            headers=headers,
            query_params=query_params,
        )
        self.cookie_jar.add_cookie_header(request)
        try:
            response = urllib.request.urlopen(request, timeout=self.timeout)
        except socket.timeout:
            print("timeout!")
            with open("/tmp/har_timeouts", "a") as f:
                f.write(url)
                f.write("\n")
            return None
        except urllib.error.HTTPError as e:
            return e.code
        self.cookie_jar.extract_cookies(response, request)
        return response.code
