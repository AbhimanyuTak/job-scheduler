#!/bin/bash

# Test Runner Script for Job Scheduler
# Provides a simple interface to run different types of tests

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if services are running
check_services() {
    log_info "Checking if services are running..."
    
    # Check if Docker Compose services are up
    if ! docker compose ps | grep -q "Up"; then
        log_error "Docker Compose services are not running"
        log_info "Please start services with: docker compose up -d"
        exit 1
    fi
    
    # Check if API is responding
    if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
        log_error "API service is not responding"
        log_info "Please ensure the scheduler service is running"
        exit 1
    fi
    
    # Check if Redis is responding
    if ! docker compose exec redis redis-cli ping > /dev/null 2>&1; then
        log_error "Redis service is not responding"
        exit 1
    fi
    
    log_success "All services are running"
}

# Run tests based on argument
run_tests() {
    case "$1" in
        "unit")
            log_info "Running unit tests..."
            make -f Makefile.test test-unit
            ;;
        "integration")
            log_info "Running integration tests..."
            check_services
            make -f Makefile.test test-integration
            ;;
        "e2e")
            log_info "Running end-to-end tests..."
            check_services
            make -f Makefile.test test-e2e
            ;;
        "redis")
            log_info "Running Redis tests..."
            check_services
            make -f Makefile.test test-redis
            ;;
        "all")
            log_info "Running all tests..."
            make -f Makefile.test test-unit
            check_services
            make -f Makefile.test test-integration
            ;;
        "quick")
            log_info "Running quick system check..."
            check_services
            make -f Makefile.test test-quick
            ;;
        "coverage")
            log_info "Running tests with coverage..."
            check_services
            make -f Makefile.test test-coverage
            ;;
        "race")
            log_info "Running tests with race detection..."
            check_services
            make -f Makefile.test test-race
            ;;
        "benchmark")
            log_info "Running benchmark tests..."
            check_services
            make -f Makefile.test test-benchmark
            ;;
        "clean")
            log_info "Cleaning test artifacts..."
            make -f Makefile.test test-clean
            ;;
        "help"|"--help"|"-h")
            show_help
            ;;
        "")
            log_info "No test type specified. Running unit tests..."
            make -f Makefile.test test-unit
            ;;
        *)
            log_error "Unknown test type: $1"
            show_help
            exit 1
            ;;
    esac
}

# Show help
show_help() {
    echo "Job Scheduler Test Runner"
    echo ""
    echo "Usage: $0 [TEST_TYPE]"
    echo ""
    echo "Available test types:"
    echo "  unit        - Run unit tests (fast, no external dependencies)"
    echo "  integration - Run integration tests (require running services)"
    echo "  e2e         - Run end-to-end tests (full system testing)"
    echo "  redis       - Run Redis-specific tests"
    echo "  all         - Run unit and integration tests"
    echo "  quick       - Run quick system check"
    echo "  coverage    - Run tests with coverage report"
    echo "  race        - Run tests with race detection"
    echo "  benchmark   - Run benchmark tests"
    echo "  clean       - Clean test artifacts"
    echo "  help        - Show this help"
    echo ""
    echo "Examples:"
    echo "  $0 unit        # Run unit tests"
    echo "  $0 e2e         # Run end-to-end tests"
    echo "  $0 coverage    # Run tests with coverage"
    echo ""
    echo "Environment variables:"
    echo "  INTEGRATION_TESTS=true  - Enable integration tests"
    echo "  TEST_BASE_URL          - Override API base URL"
    echo "  TEST_API_KEY           - Override API key"
    echo "  TEST_REDIS_HOST        - Override Redis host"
    echo "  TEST_REDIS_PORT        - Override Redis port"
    echo "  TEST_TIMEOUT           - Override test timeout (seconds)"
    echo "  TEST_MAX_RETRIES       - Override max retries"
    echo "  TEST_WAIT_INTERVAL     - Override wait interval (seconds)"
}

# Main execution
main() {
    echo "=========================================="
    echo "Job Scheduler Test Runner"
    echo "=========================================="
    echo ""
    
    # Check if Go is available
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    # Check if Make is available
    if ! command -v make &> /dev/null; then
        log_error "Make is not installed or not in PATH"
        exit 1
    fi
    
    # Run tests
    run_tests "$1"
    
    echo ""
    log_success "Test execution completed!"
}

# Run main function
main "$@"
