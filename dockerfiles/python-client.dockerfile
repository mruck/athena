from debian:buster

RUN apt-get update && apt-get install -y bpython3 \
    jq \
    postgresql-client \
    python3 \
    python3-pip \
    sudo \
    watch \
    vim
RUN pip3 install psycopg2 requests
RUN pip3 install --upgrade virtualenv

RUN mkdir /client
WORKDIR /client

ADD ./pip-reqs.txt pip-reqs.txt
ADD ./Makefile Makefile
ENV VENV_LOCATION=/venv
RUN make venv

ADD . /client

WORKDIR /client/fuzzer
ENTRYPOINT ./run_client.sh
