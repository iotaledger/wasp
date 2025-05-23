#!/bin/bash

docker run --name iota-postgres -e POSTGRES_PASSWORD=postgres -p:5432:5432 -d postgres

iota start --graphql-port=9001 --graphql-host="0.0.0.0"  --with-indexer --pg-password="postgres" --pg-db-name="postgres"
