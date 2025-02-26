#!/bin/bash

set +e
set +u

function create_user_and_database() {
	local database=$1
    local database_user="${1}_user"
    local user=$database_user
    if [ -n "$POSTGRES_USER" ]; then
        user=$POSTGRES_USER
    fi
	echo "  Creating user '$user' and database '$database'"
	psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
        SELECT 'CREATE USER $user'
        WHERE NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '$user')\gexec
EOSQL
	psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
        SELECT 'CREATE DATABASE $database'
        WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '$database')\gexec
EOSQL
	psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
	    GRANT ALL PRIVILEGES ON DATABASE $database TO $user;
EOSQL
}

if [ -n "$POSTGRES_MULTIPLE_DATABASES" ]; then
	echo "Multiple database creation requested: $POSTGRES_MULTIPLE_DATABASES"
	for db in $(echo $POSTGRES_MULTIPLE_DATABASES | tr ',' ' '); do
		create_user_and_database $db
	done
	echo "Multiple databases created"
fi