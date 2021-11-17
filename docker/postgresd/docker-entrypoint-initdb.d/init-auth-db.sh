#!/bin/bash
set -e
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE DATABASE auth;
    GRANT ALL PRIVILEGES ON DATABASE auth TO $POSTGRES_USER;
    \c auth;
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
EOSQL
