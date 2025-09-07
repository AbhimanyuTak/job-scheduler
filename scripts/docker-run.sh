#!/bin/bash

# Docker run script for job scheduler

set -e

echo "🐳 Running Job Scheduler with Docker..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

# Build and start all services
echo "🏗️  Building and starting all services..."
docker compose up --build -d

# Wait for services to be ready
echo "⏳ Waiting for services to be ready..."

# Wait for PostgreSQL
until docker compose exec postgres pg_isready -U postgres; do
    echo "Waiting for PostgreSQL..."
    sleep 2
done

# Wait for application health check
echo "Waiting for application to be ready..."
until curl -f http://localhost:8080/health > /dev/null 2>&1; do
    echo "Waiting for application..."
    sleep 2
done

echo "✅ All services are ready!"
echo ""
echo "🌐 Application:"
echo "   URL: http://localhost:8080"
echo "   Health: http://localhost:8080/health"
echo ""
echo "📊 Database:"
echo "   Host: localhost:5432"
echo "   Database: job_scheduler"
echo ""
echo "📝 To view logs:"
echo "   docker compose logs -f"
echo ""
echo "🛑 To stop all services:"
echo "   docker compose down"
