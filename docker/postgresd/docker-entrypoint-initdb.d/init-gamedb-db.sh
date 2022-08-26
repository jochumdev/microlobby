#!/bin/bash
set -e
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE DATABASE gamedb;
    GRANT ALL PRIVILEGES ON DATABASE gamedb TO $POSTGRES_USER;
    \c gamedb;
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
EOSQL
