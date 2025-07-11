#!/bin/bash

# Docker production environment setup script for Habbr
set -e

echo "🚀 Starting Habbr production environment..."

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

# Check for production environment variables
if [ -z "$DATABASE_PASSWORD" ]; then
    print_warning "DATABASE_PASSWORD not set. Using default from docker-compose.yml"
fi

# Build production image
print_status "Building production Docker image..."
docker compose build habbr-api

# Start infrastructure services
print_status "Starting PostgreSQL and Redis..."
docker compose up -d postgres redis

# Wait for services to be ready
print_status "Waiting for services to be ready..."
timeout=120
counter=0

# Wait for PostgreSQL
until docker compose exec postgres pg_isready -U habbr_user -d habbr > /dev/null 2>&1; do
    if [ $counter -ge $timeout ]; then
        print_error "PostgreSQL failed to start within $timeout seconds"
        docker compose logs postgres
        exit 1
    fi
    sleep 2
    counter=$((counter + 2))
    echo -n "."
done
echo ""

# Wait for Redis
until docker compose exec redis redis-cli ping > /dev/null 2>&1; do
    if [ $counter -ge $timeout ]; then
        print_error "Redis failed to start within $timeout seconds"
        docker compose logs redis
        exit 1
    fi
    sleep 1
    counter=$((counter + 1))
    echo -n "."
done
echo ""

# Run database migrations
print_status "Running database migrations..."
docker compose exec postgres psql -U habbr_user -d habbr -f /docker-entrypoint-initdb.d/001_initial_schema.sql > /dev/null 2>&1 || {
    print_warning "Initial migration might already be applied"
}

docker compose exec postgres psql -U habbr_user -d habbr -f /docker-entrypoint-initdb.d/003_performance_indexes.sql > /dev/null 2>&1 || {
    print_warning "Performance migration might already be applied"
}

# Start the application
print_status "Starting Habbr API in production mode..."
docker compose up -d habbr-api

# Wait for application to be ready
print_status "Waiting for application to be ready..."
timeout=60
counter=0
until curl -f http://localhost:8080/health > /dev/null 2>&1; do
    if [ $counter -ge $timeout ]; then
        print_error "Application failed to start within $timeout seconds"
        docker compose logs habbr-api
        exit 1
    fi
    sleep 2
    counter=$((counter + 2))
    echo -n "."
done
echo ""

print_status "Production environment is ready!"
print_status "API available at: http://localhost:8080"
print_status "Health check: http://localhost:8080/health"

# Show running containers
print_status "Running containers:"
docker compose ps

# Show logs if requested
if [ "$1" = "--logs" ]; then
    print_status "Showing application logs..."
    docker compose logs -f habbr-api
fi
