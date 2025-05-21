#!/bin/bash
# This script completely cleans up Docker resources for the project

# Stop all containers
echo "Stopping all containers..."
docker-compose down

# Remove the volume
echo "Removing PostgreSQL volume..."
docker volume rm project_maker_postgres_data

# Remove any orphaned containers
echo "Removing any orphaned containers..."
docker container prune -f

# Clean system
echo "Cleaning up Docker system..."
docker system prune -f

echo "Cleanup complete. Now try running 'make docker-up' again." 