# Fuzzer

## Getting the fuzzer up and running
`mkdir athena && cd athena`
`git clone git@github.com:mruck/fuzzer.git`
`git clone git@github.com:mruck/discourse-fork.git`
`git clone git@github.com:mruck/coverage-visualizer.git``
`git clone git@github.com:mruck/rails-fork.git``
`cd rails-fork && git checkout 5-2-1-bip`
`cd fuzzer && make all`.

### Python Environment
To install the required python packages, make sure you have virtualenv installed:
```
pip install --upgrade virtualenv
```
Then you can create the python virtual environment with:
```
make venv
source ./venv/bin/activate
```

### Discourse Clone
For now, the fuzzer is tuned to work with Discourse. This means we need to spawn a
postgres container for the server to interact with:
```
make postgres-start
```
### Postgres
Currently db instrumentation is done by hooking the rails fork.  Instead, we
should either 1) stick a shim between the app and the db or 2) tail postgres
queries.  These solutions are framework agnostic.  The latter is easier, so
eventually we should do that.  Specifically it requires the config to look like
`/athena/postgres/postgresql.conf` and that should be copies to
`/var/lib/postgresql/data/postgresql.conf` in the postgres container. I wanted to create
a custom image with this set up, but that value can't be set at container start up,
so I need to do something more complicated which is why I'm deferring it.  If I
set the conf file after postgres start up, connect to any db via psql then run
` SELECT pg_reload_conf(); `.
Note: rails does db caching, not sure if that will hurt us because it won't make new
queries.  To toggle off, see:
https://stackoverflow.com/questions/3599875/disable-sql-cache-temporary-in-rails
