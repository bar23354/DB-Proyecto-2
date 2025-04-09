#!/bin/bash

set -e

# Wait for PostgreSQL to be ready
until psql -U postgres -c '\q'; do
  >&2 echo "PostgreSQL is unavailable - sleeping"
  sleep 1
done

# Create the database
psql -U postgres -c "CREATE DATABASE airline;"

# Initialize the database schema
psql -U postgres -d airline -f /app/ddl.sql

# Load the initial data
psql -U postgres -d airline -f /app/data.sql

>&2 echo "PostgreSQL is ready - database initialized"