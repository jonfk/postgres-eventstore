# postgres-eventstore

This is an experiment on using postgres as an event store for the event sourcing pattern.

# Instructions
```bash
$ source .env

# First deploy postgres to docker
$ ./deploy_postgres.sh

# run migrations
$ diesel migration run

# install dependencies
$ glide install

# build
$ make
```