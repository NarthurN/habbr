#!/bin/bash

# Docker development environment setup script for Habbr
set -e

echo "🚀 Starting Habbr development environment..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    print_error "Docker is not running. Please start Docker and try again."
    exit 1
fi

# Check if Docker Compose is available
if ! docker compose version > /dev/null 2>&1; then
    print_error "docker compose is not available. Please ensure Docker Compose is installed."
    exit 1
fi

# Build development image
print_status "Building development Docker image..."
docker compose build habbr-api

# Start PostgreSQL and Redis first
print_status "Starting PostgreSQL and Redis..."
docker compose up -d postgres redis

# Wait for PostgreSQL to be ready
print_status "Waiting for PostgreSQL to be ready..."
timeout=60
counter=0
until docker compose exec postgres pg_isready -U habbr_user -d habbr > /dev/null 2>&1; do
    if [ $counter -ge $timeout ]; then
        print_error "PostgreSQL failed to start within $timeout seconds"
        docker compose logs postgres
        exit 1
    fi
    sleep 1
    counter=$((counter + 1))
    echo -n "."
done
echo ""

# Run migrations
print_status "Running database migrations..."
docker compose exec postgres psql -U habbr_user -d habbr -f /docker-entrypoint-initdb.d/001_initial_schema.sql > /dev/null 2>&1 || {
    print_warning "Initial migration might already be applied"
}

# Start development tools (optional)
if [ "$1" = "--with-tools" ]; then
    print_status "Starting development tools (pgAdmin, Redis Insight)..."
    docker compose --profile dev up -d pgadmin redis-insight
fi

# Start the application in development mode
print_status "Starting Habbr API in development mode..."
docker compose up habbr-api

print_status "Development environment is ready!"
print_status "API available at: http://localhost:8080"
print_status "GraphQL Playground: http://localhost:8080"

if [ "$1" = "--with-tools" ]; then
    print_status "pgAdmin available at: http://localhost:5050 (admin@habbr.local / admin)"
    print_status "Redis Insight available at: http://localhost:8001"
fi
