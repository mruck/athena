from golang:1.12

RUN apt-get update && apt-get install -y bpython3 \
    jq \
    postgresql-client \
    python3 \
    python3-pip \
    sudo \
    watch \
    vim

RUN mkdir /client
WORKDIR /client

ADD goFuzz/* /client

WORKDIR /client
