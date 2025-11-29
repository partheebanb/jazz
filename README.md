# Jazz - Self-Hosted Logging Platform

A lightweight, self-hosted logging platform built with Go, PostgreSQL, and Docker. Jazz provides a simple alternative to enterprise logging services like Sentry or Datadog for small teams and hobby projects.

## Features

- 🚀 **Fast Full-Text Search** - PostgreSQL GIN indexes provide sub-100ms search across millions of logs
- 🔒 **Multi-Tenant** - Isolated projects with API key authentication
- 📊 **Advanced Filtering** - Filter by level, source, timestamp, and search queries
- 🐳 **Docker-First** - Production-ready containerized deployment
- 📈 **Performant** - Batch log ingestion handles 1000+ logs/second
- 🧪 **Well-Tested** - 86%+ code coverage with comprehensive test suite

## Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.25+ (for development)

### Run with Docker

```bash
# Clone the repository
git clone https://github.com/yourusername/jazz.git
cd jazz

# Start the services
docker-compose up -d

# Run database migrations
docker exec jazz-api ./jazz-migrate

# Verify it's running
curl http://localhost:8080/health
```

Jazz API is now running on `http://localhost:8080`!

## Usage

### 1. Create a Project

```bash
curl -X POST http://localhost:8080/projects \
  -H "Content-Type: application/json" \
  -d '{"name":"My Application"}'
```

Response:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "My Application",
  "api_key": "jazz_a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "created_at": "2024-11-22T10:30:00Z",
  "updated_at": "2024-11-22T10:30:00Z"
}
```

**Save your API key** - you'll need it to send logs!

### 2. Send Logs

```bash
curl -X POST http://localhost:8080/logs \
  -H "Authorization: Bearer jazz_YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "logs": [
      {
        "level": "error",
        "message": "Database connection failed",
        "source": "backend"
      },
      {
        "level": "info",
        "message": "User logged in successfully",
        "source": "auth"
      }
    ]
  }'
```

### 3. Query Logs

**Get recent logs:**
```bash
curl "http://localhost:8080/logs?limit=10" \
  -H "Authorization: Bearer jazz_YOUR_API_KEY"
```

**Filter by level:**
```bash
curl "http://localhost:8080/logs?level=error&limit=20" \
  -H "Authorization: Bearer jazz_YOUR_API_KEY"
```

**Filter by time range:**
```bash
curl "http://localhost:8080/logs?start_time=2024-11-01T00:00:00Z&end_time=2024-11-22T23:59:59Z" \
  -H "Authorization: Bearer jazz_YOUR_API_KEY"
```

### 4. Search Logs

**Full-text search:**
```bash
curl -X POST http://localhost:8080/search \
  -H "Authorization: Bearer jazz_YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "database timeout",
    "limit": 50
  }'
```

**Search with filters:**
```bash
curl -X POST http://localhost:8080/search \
  -H "Authorization: Bearer jazz_YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "connection failed",
    "level": "error",
    "source": "backend",
    "start_time": "2024-11-01T00:00:00Z"
  }'
```

## API Reference

### Projects

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/projects` | POST | Create a new project |
| `/projects` | GET | List all projects |
| `/projects/:id` | GET | Get project details |
| `/projects/:id` | DELETE | Delete a project |

### Logs (Requires API Key)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/logs` | POST | Ingest logs (batch up to 1000) |
| `/logs` | GET | Query logs with filters |
| `/search` | POST | Full-text search logs |
| `/health` | GET | Health check |

### Request/Response Examples

**Query Logs Response:**
```json
{
  "logs": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "project_id": "550e8400-e29b-41d4-a716-446655440000",
      "level": "error",
      "message": "Database connection timeout",
      "source": "backend",
      "timestamp": "2024-11-22T10:30:00Z"
    }
  ],
  "total": 1542,
  "limit": 50,
  "offset": 0,
  "has_more": true
}
```

**Search Response:**
```json
{
  "logs": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "level": "error",
      "message": "Database connection timeout",
      "source": "backend",
      "timestamp": "2024-11-22T10:30:00Z",
      "rank": 0.8574
    }
  ],
  "total": 45,
  "query_time_ms": 42
}
```

## Development

### Setup Local Development

```bash
# Clone repository
git clone https://github.com/yourusername/jazz.git
cd jazz

# Install dependencies
go mod download

# Start PostgreSQL
docker-compose up -d postgres

# Run migrations
go run cmd/migrate/main.go

# Run API
go run main.go
```

### Project Structure

```
jazz/
├── cmd/
│   └── migrate/          # Database migration tool
├── database/             # Database layer
│   ├── db.go            # Connection management
│   ├── projects.go      # Project CRUD
│   ├── logs.go          # Log operations
│   ├── search.go        # Full-text search
│   └── query_builder.go # SQL query builder
├── handlers/             # HTTP handlers
│   ├── logs.go          # Log endpoints
│   └── projects.go      # Project endpoints
├── middleware/           # HTTP middleware
│   └── auth.go          # API key authentication
├── models/               # Data models
│   ├── log.go
│   └── project.go
├── docker-compose.yml    # Docker services
├── Dockerfile           # Multi-stage build
└── main.go              # Application entry point
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run only unit tests (fast)
go test -short ./...

# View coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Test database is automatically created and destroyed** - no manual setup required!

### Code Quality

```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run

# Vet code
go vet ./...
```

## Deployment

### Environment Variables

```bash
# Required
DATABASE_URL=postgres://user:password@host:5432/jazz?sslmode=disable

# Optional
PORT=8080                    # API port (default: 8080)
GIN_MODE=release            # Gin mode (debug/release)
```

### Deploy to Fly.io

```bash
# Install flyctl
curl -L https://fly.io/install.sh | sh

# Login
fly auth login

# Launch app
fly launch

# Deploy
fly deploy
```

### Deploy to Railway

1. Connect your GitHub repository
2. Add PostgreSQL service
3. Set `DATABASE_URL` environment variable
4. Deploy!

### Production Checklist

- [ ] Set `GIN_MODE=release`
- [ ] Use strong database password
- [ ] Enable HTTPS
- [ ] Set up database backups
- [ ] Configure log retention policies
- [ ] Monitor disk usage
- [ ] Set up alerts

## Architecture

### Database Schema

**Projects Table:**
```sql
CREATE TABLE projects (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    api_key VARCHAR(64) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

**Logs Table:**
```sql
CREATE TABLE logs (
    id UUID PRIMARY KEY,
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    level VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    source VARCHAR(100),
    timestamp TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
```

### Performance

- **Search Speed**: <100ms for 1M logs (PostgreSQL GIN indexes)
- **Ingestion**: 1000+ logs/second (batch inserts)
- **Connection Pooling**: 25 max connections, 5 min connections
- **Query Optimization**: COUNT(*) OVER() for single-query pagination

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests for your changes
4. Ensure tests pass (`go test ./...`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## License

MIT License

## Acknowledgments

- Built with [Gin](https://github.com/gin-gonic/gin) web framework
- Uses [pgx](https://github.com/jackc/pgx) for PostgreSQL connectivity
- Inspired by Sentry, Datadog, and LogRocket