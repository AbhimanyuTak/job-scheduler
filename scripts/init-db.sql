-- Database initialization script for job scheduler
-- This script runs when the PostgreSQL container starts for the first time

-- Create database if it doesn't exist (already created by POSTGRES_DB env var)
-- But we can add any additional setup here

-- Create extensions if needed
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Set timezone
SET timezone = 'UTC';

-- Create a read-only user for monitoring (optional)
-- CREATE USER scheduler_readonly WITH PASSWORD 'readonly_password';
-- GRANT CONNECT ON DATABASE job_scheduler TO scheduler_readonly;
-- GRANT USAGE ON SCHEMA public TO scheduler_readonly;
-- GRANT SELECT ON ALL TABLES IN SCHEMA public TO scheduler_readonly;
-- ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO scheduler_readonly;

-- Log successful initialization
\echo 'Database initialization completed successfully!'
