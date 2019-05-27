from golang:1.12

RUN apt-get update && apt-get install -y bpython3 \
    less \
    jq \
# For the mongo shell
    mongodb \
    postgresql-client \
    python3 \
    python3-pip \
    sudo \
    watch \
    vim

RUN pip3 install moz-sql-parser
RUN mkdir /athena
WORKDIR /athena
ADD . /athena
WORKDIR /athena/goFuzz
# Build it so it pulls the dependencies
RUN go build -o athena *.go
CMD ./athena
