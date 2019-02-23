#! /usr/local/bin/python3

import argparse
import collections
import json
import os.path
import pickle


class FuzzStats(object):
    def __init__(self):
        self.covs = []
        self.results = []
        self.exceptions = []

        self._2xx = [0, set()]
        self._3xx = [0, set()]
        self._4xx = [0, set()]
        self._5xx = [0, set()]
        self._timeout = [0, set()]

    def _record_result(self, verb, path, code, exns, snapshot_name):
        self.results.append((verb, path, code, exns, snapshot_name))

    def _record_code(self, verb, path, code):
        entry = self._2xx
        if code is None:
            entry = self._timeout
        elif code >= 300 and code < 400:
            entry = self._3xx
        elif code >= 400 and code < 500:
            entry = self._4xx
        elif code >= 500:
            entry = self._5xx

        count, routes = entry
        count += 1
        routes.add(verb + ":" + path)
        entry[0], entry[1] = count, routes

    def _read_exceptions(self, exn_fp):
        # Load exceptions from json
        exns = []
        for line in exn_fp:
            exns.append(json.loads(line.strip()))
        return exns

    def record_stats(self, verb, path, code, exn_fp, snapshot_name):
        exns = self._read_exceptions(exn_fp)
        self._record_result(verb, path, code, exns, snapshot_name)
        self._record_code(verb, path, code)

    def record_coverage(self, verb, path, cov):
        self.covs.append((verb, path, cov))

    def final_coverage(self):
        return self.covs[-1][2]

    def get_code_counts(self):
        cnt = collections.Counter()
        for result in self.results:
            code = result[2]
            cnt[code] += 1
        return cnt

    def get_results(self, verb=None, path=None, code=None):
        output = []
        for result in self.results:
            if verb is not None and result[0] != verb:
                continue
            elif path is not None and result[1] != path:
                continue
            elif code is not None and result[2] != code:
                continue
            output.append(result)
        return output

    def get_exceptions(self, verb=None, path=None):
        output = []
        for exn in self.exceptions:
            if verb is not None and exn[0] != verb:
                continue
            elif path is not None and exn[1] != path:
                continue
            output.append(exn)
        return output

    def get_success_ratio(self):
        return float(self._2xx[0]) / float(len(self.results))

    def save(self, dest):
        dest = os.path.join(dest, "fuzz_stats")
        fh = open(dest, "wb")
        pickle.dump(self, fh)
        fh.close()


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Fuzzing client")
    parser.add_argument("f", help="Fuzz stats pickle file")
    parser.add_argument("--verb")
    parser.add_argument("--path")
    parser.add_argument("--code", type=int)
    args = parser.parse_args()

    fh = open(args.f, "rb")
    stats = pickle.load(fh)
    fh.close()

    results = stats.get_results(verb=args.verb, path=args.path, code=args.code)
    print(json.dumps(results, indent=2))
