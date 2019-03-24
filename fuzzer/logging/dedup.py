#! /usr/bin/python
# Example usage:
#       python3 dedup.py class log0,log1,log2,log3 | jq .
import json
import argparse


def aggregate(aggregators, files):
    output = {}

    for fname in files:
        fh = open(fname, "r")
        contents = fh.read().split("\n")

        for line in contents:
            if line == "":
                continue
            obj = json.loads(line)

            keys = []
            for prop in aggregators:
                keys.append(obj[prop])
            key = " -- ".join(keys)

            if key not in output:
                output[key] = 0
            output[key] += 1

    print(json.dumps(output))


def get_args():
    parser = argparse.ArgumentParser(description="")
    parser.add_argument("aggregators")
    parser.add_argument("files")
    return parser.parse_args()


if __name__ == "__main__":
    args = get_args()
    aggregate(args.aggregators.split(","), args.files.split(","))
