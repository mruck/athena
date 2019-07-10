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

RUN mkdir /athena
WORKDIR /athena
# Add deps so they get cached
ADD go.mod /athena
ADD go.sum /athena
RUN go mod download
ADD . /athena
WORKDIR /athena/goFuzz
RUN mkdir -p /var/log/athena
CMD go run main.go
