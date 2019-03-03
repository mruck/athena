#!/bin/bash
set -xe
# Create DB
dropdb $FUZZ_DB || true
createdb -T template0 $FUZZ_DB && psql -d $FUZZ_DB < state/db.pgdump

# Discourse specific
pkill -9 redis || true
redis-server --daemonize yes

# Medium specific
sudo service elasticsearch stop || true
sudo service elasticsearch start

# Run app
TARGET_APP_PATH=`pwd` RAILS_ENV=development RAILS_MASTER=1 bundle exec bin/rails s -b 0.0.0.0 -p $PORT
