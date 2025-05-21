#!/bin/bash
# This script checks the status of the PostgreSQL container

echo "==== PostgreSQL Container Status ===="
docker ps -a | grep postgres

echo -e "\n==== PostgreSQL Container Logs ===="
docker logs project_maker-postgres-1

echo -e "\n==== Testing PostgreSQL Connection ===="
docker exec -it project_maker-postgres-1 pg_isready -U postgres || echo "Connection failed"

echo -e "\n==== Checking Docker Network ===="
docker network inspect project_maker_default

echo -e "\n==== PostgreSQL Volume ===="
docker volume inspect project_maker_postgres_data

echo -e "\n==== Host Port Status (5432) ===="
if command -v lsof > /dev/null; then
  lsof -i :5432 || echo "No process using port 5432"
else
  echo "lsof not installed, skipping port check"
fi 