#!/bin/bash
set -xe
./create_db.sh
pkill -9 redis || true
redis-server --daemonize yes
TARGET_APP_PATH=`pwd` RAILS_ENV=development RAILS_MASTER=1 bundle exec bin/rails s -b 0.0.0.0 -p $PORT
