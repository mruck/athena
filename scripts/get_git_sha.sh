#! /bin/bash

function hash_source () {
    (find $1 -type f -print0  | sort -z | xargs -0 sha1sum; find $1 \( -type f -o -type d \) -print0 | sort -z |xargs -0 gstat -c '%n %a') | \
        sha1sum | cut -d ' ' -f 1 | head -c 10
}

repo_root=$(git rev-parse --show-toplevel)
git_sha=$(git log | head -n 1 | cut -f 2 -d ' ' | head -c 10)
suffix=$([ -z "$(git status --porcelain)" ] || hash_source $repo_root)

echo "$git_sha""$suffix"
