# Fuzzer

## Getting the fuzzer up and running
`mkdir athena && cd athena`
`git clone git@github.com:mruck/fuzzer.git`
`git clone git@github.com:mruck/discourse-fork.git`
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
