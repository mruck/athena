#!/bin/bash
# Arg1: db to drop
# Arg2: directory with db dump
set -x
dropdb $1
createdb -T template0 $1
psql $1 < $HOME/discourse-fork/state/db.pgdump
