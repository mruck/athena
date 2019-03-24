#!/bin/bash

# Kill all pids connected to db
psql -d $1 -c "SELECT pid, datname FROM pg_stat_activity;" | grep $1 | cut -d '|' -f 1 | tr -d ' ' | xargs sudo kill -9
