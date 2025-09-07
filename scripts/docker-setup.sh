#!/bin/bash

# Docker setup script for job scheduler

set -e

echo "🐳 Setting up Job Scheduler with Docker..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

# Check if docker compose is available
if ! docker compose version &> /dev/null; then
    echo "❌ docker compose is not available. Please install Docker Compose and try again."
    exit 1
fi

echo "✅ Docker is running"

# Start PostgreSQL only for development
echo "🚀 Starting PostgreSQL database..."
docker compose -f docker-compose.dev.yml up -d postgres

# Wait for PostgreSQL to be ready
echo "⏳ Waiting for PostgreSQL to be ready..."
until docker compose -f docker-compose.dev.yml exec postgres pg_isready -U postgres; do
    echo "Waiting for PostgreSQL..."
    sleep 2
done

echo "✅ PostgreSQL is ready!"

# Display connection information
echo ""
echo "📊 Database Connection Information:"
echo "   Host: localhost"
echo "   Port: 5432"
echo "   Database: job_scheduler"
echo "   Username: postgres"
echo "   Password: password"
echo ""
echo "🚀 You can now run the application with:"
echo "   go run ./cmd/scheduler"
echo ""
echo "📝 To stop the database:"
echo "   docker compose -f docker-compose.dev.yml down"
