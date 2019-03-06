from debian:buster

RUN apt-get update && apt-get install -y bpython3 \
    jq \
    postgresql-client \
    python3 \
    python3-pip \
    sudo \
    vim 
RUN pip3 install psycopg2 requests 
RUN pip3 install --upgrade virtualenv

RUN mkdir /client
ADD . /client
WORKDIR /client
ENV VENV_LOCATION=/venv
RUN make venv

ENTRYPOINT ./run_client.sh
