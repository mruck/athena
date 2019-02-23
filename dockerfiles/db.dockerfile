# Create and populate a db
from debian:buster

RUN apt-get update
RUN apt-get install -y postgresql-client \
sudo \
vim 
RUN mkdir /db
ADD create_db.sh /db
ADD discourse.dump /db/db.pgdump
WORKDIR /db
ENTRYPOINT ./create_db.sh
