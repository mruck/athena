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

RUN mkdir /athena
WORKDIR /athena
ADD . /athena
WORKDIR /athena/goFuzz
CMD go run main.go
