# Configuration Guide

## Environment Variables

The application uses environment variables for configuration. You can set these in two ways:

1. **Using config.env file** (recommended for development)
2. **Using system environment variables** (recommended for production)

## Configuration File

Create a `config.env` file in the project root with the following variables:

```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=job_scheduler
DB_SSLMODE=disable

# Application Configuration
GIN_MODE=debug
PORT=8080

# API Configuration
API_KEY=your-secret-api-key-here
```

## Environment Variables Reference

### Database Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | Database username |
| `DB_PASSWORD` | `password` | Database password |
| `DB_NAME` | `job_scheduler` | Database name |
| `DB_SSLMODE` | `disable` | SSL mode for connection |

### Application Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `GIN_MODE` | `debug` | Gin framework mode (debug/release) |
| `PORT` | `8080` | HTTP server port |

### API Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `API_KEY` | `your-secret-api-key-here` | API key for authentication |

## Setup Instructions

### Development Setup

1. **Copy the configuration template:**
   ```bash
   cp config.env.example config.env
   ```

2. **Edit the configuration:**
   ```bash
   nano config.env
   ```

3. **Update the values as needed:**
   - Change `DB_PASSWORD` to your database password
   - Set a secure `API_KEY`
   - Adjust other settings as needed

### Production Setup

1. **Set environment variables:**
   ```bash
   export DB_HOST=your-db-host
   export DB_PASSWORD=your-secure-password
   export API_KEY=your-secure-api-key
   export GIN_MODE=release
   ```

2. **Or use a process manager like systemd:**
   ```ini
   [Service]
   Environment=DB_HOST=your-db-host
   Environment=DB_PASSWORD=your-secure-password
   Environment=API_KEY=your-secure-api-key
   Environment=GIN_MODE=release
   ```

## Security Notes

- **Never commit `config.env` to version control**
- Use strong passwords for database connections
- Generate secure API keys for production
- Use `GIN_MODE=release` in production
- Consider using SSL (`DB_SSLMODE=require`) for production databases

## Docker Configuration

When using Docker, environment variables can be passed through:

```bash
# Using docker-compose
docker-compose up -e DB_PASSWORD=your-password

# Using docker run
docker run -e DB_PASSWORD=your-password job-scheduler
```

## Configuration Loading

The application loads configuration in the following order:

1. **config.env file** (if present)
2. **System environment variables** (if config.env not found)
3. **Default values** (if neither is set)

This allows for flexible configuration management across different environments.
