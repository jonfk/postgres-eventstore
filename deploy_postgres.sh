#!/usr/bin/env sh

set -x
set -e

docker pull postgres:9.5
docker run --name postgres-eventstore -e POSTGRES_PASSWORD=${DATABASE_PASSWORD} \
                                        -e POSTGRES_USER=${DATABASE_USER} \
                                        -p 5432:5432 \
                                        -d postgres:9.5
