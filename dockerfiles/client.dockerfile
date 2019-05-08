from golang:1.12

RUN apt-get update && apt-get install -y bpython3 \
    less \
    jq \
    postgresql-client \
    python3 \
    python3-pip \
    sudo \
    watch \
    vim

RUN mkdir /fuzz
WORKDIR /fuzz

ADD goFuzz/ /fuzz

WORKDIR /fuzz
