#! /usr/local/bin/bash

if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
fi


# -a -f myInsertFile
# PGPASSWORD=$POSTGRES_PASSWORD psql -h 127.0.0.1 -p 32432 -U $PG_AUTH_USER 
PGPASSWORD=$POSTGRES_PASSWORD psql -h 127.0.0.1 -p $POSTGRES_PORT -d $POSTGRES_DB -U $POSTGRES_USER 

