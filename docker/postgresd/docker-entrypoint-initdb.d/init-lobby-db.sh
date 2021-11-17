#!/bin/bash
set -e
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE DATABASE lobby;
    GRANT ALL PRIVILEGES ON DATABASE lobby TO $POSTGRES_USER;
    \c lobby;
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
EOSQL
