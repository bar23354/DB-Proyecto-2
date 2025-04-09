#!/bin/sh

set -e

echo "Esperando a que PostgreSQL esté listo..."

until pg_isready -U postgres; do
  echo "PostgreSQL no está disponible aún - esperando..."
  sleep 2
done

echo "Aplicando esquema DDL..."
psql -U postgres -d airline -f /app/ddl.sql

echo "Cargando datos iniciales..."
psql -U postgres -d airline -f /app/data.sql

echo "¡Base de datos inicializada con éxito!"