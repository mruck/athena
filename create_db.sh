#!/bin/bash

set -x
dropdb $DB_NAME
createdb -T template0 $DB_NAME && psql -d $DB_NAME < db.pgdump
