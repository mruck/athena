#!/bin/bash
set -x
source $VENV_LOCATION/bin/activate
python3 fuzz.py $RESULTS_PATH  $DB_NAME $PORT /discourse-fork --fuzzer_number $FUZZER_NUMBER --instances $INSTANCES $ROUTE $SNAPSHOT $ANY_ROUTE $LOAD_DB $STOP_AFTER_ALL_ROUTES $STOP_AFTER_HAR
