#!/bin/bash
set -e
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE DATABASE settings;
    GRANT ALL PRIVILEGES ON DATABASE settings TO $POSTGRES_USER;
    \c settings;
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
EOSQL
