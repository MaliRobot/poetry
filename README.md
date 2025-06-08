# Poetry Management System

A distributed poetry management system with separate API and worker processes for handling large-scale poem processing.

## Architecture

The system is now split into separate services:

- **API Server** (`server/server.go`): Handles HTTP requests and forwards bulk operations to the worker
- **Worker Service** (`worker/worker.go`): Processes bulk poem insertions asynchronously
- **MongoDB**: Stores poem data
- **Elasticsearch**: Provides search functionality

## Components

### API Server
- Handles individual poem submissions synchronously
- Forwards bulk poem uploads to the worker service via HTTP
- Provides search and collection listing functionality
- Runs on port 8080 (or configured port)

### Worker Service
- Runs as a separate process/container
- Processes bulk poem insertions asynchronously
- Configurable buffer size and worker count
- Provides HTTP endpoints for job submission and health checks
- Runs on port 8082 (or configured port)

## Worker Features

### Queue Management
- Configurable buffer size to handle load spikes
- Non-blocking job submission with overflow protection
- Queue size monitoring

### Concurrency
- Multiple worker goroutines for parallel processing
- Configurable worker count based on system resources

### HTTP API
- `POST /jobs` - Submit poems for processing
- `GET /health` - Health check
- `GET /status` - Queue status and metrics

## Environment Variables

### Worker Service
- `WORKER_PORT` - Port for worker HTTP server (default: 8081)
- `WORKER_BUFFER_SIZE` - Job queue buffer size (default: 10)
- `WORKER_MAX_WORKERS` - Number of worker goroutines (default: 3)

### API Server
- `WORKER_HOST` - Worker service hostname (default: localhost)
- `WORKER_PORT` - Worker service port (default: 8081)

### Database
- `DB_HOST` - MongoDB hostname
- `DB_PORT` - MongoDB port
- `DB_USER` - MongoDB username
- `DB_PASS` - MongoDB password
- `DB_NAME` - Database name

## Running the System

### With Docker Compose (Recommended)

```bash
# Start all services
docker-compose up --build

# Start in background
docker-compose up -d --build

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

Services will be available at:
- API Server: http://localhost:8080
- Worker Service: http://localhost:8082
- MongoDB Express: http://localhost:8081
- Elasticsearch: http://localhost:9200
- Kibana: http://localhost:5601

### Manual Setup

1. Start MongoDB and Elasticsearch
2. Set environment variables
3. Start the worker service:
   ```bash
   go run cmd/worker/main.go
   ```
4. Start the API server:
   ```bash
   go run main.go
   ```

## API Endpoints

### Individual Poem
```bash
POST /poem
Content-Type: application/json

{
  "title": "Sample Poem",
  "poem": "This is a sample poem content",
  "poet": "Author Name",
  "language": "en",
  "dataset": "sample_dataset",
  "tags": "nature,love"
}
```

### Bulk Poems
```bash
POST /poems
Content-Type: multipart/form-data

# Upload a JSON file containing an array of poems
```

### Search
```bash
GET /search?q=your_search_term
```

### Collections
```bash
GET /collections
```

## Testing

### Unit Tests
```bash
# Test worker functionality
go test ./worker -v

# Test all packages
go test ./... -v
```

### Integration Tests
```bash
# Start services first
docker-compose up -d

# Test API endpoints
curl -X GET http://localhost:8080/ping
curl -X GET http://localhost:8082/health

# Test worker status
curl -X GET http://localhost:8082/status
```

## Worker Overload Protection

The worker implements several mechanisms to prevent API blocking during overload:

1. **Buffered Channels**: Jobs are queued in a buffered channel
2. **Non-blocking Submission**: API returns immediately if worker queue is full
3. **Graceful Degradation**: API returns 503 Service Unavailable when worker is overloaded
4. **Multiple Workers**: Configurable number of worker goroutines for parallel processing

## Monitoring

### Worker Metrics
- Queue size via `/status` endpoint
- Health status via `/health` endpoint
- Processing logs with timing information

### API Metrics
- HTTP status codes for success/failure tracking
- Error logging for worker communication issues

## Scaling

### Horizontal Scaling
- Multiple worker instances can be deployed
- Load balancer can distribute jobs across workers
- Each worker maintains its own queue

### Vertical Scaling
- Increase `WORKER_MAX_WORKERS` for more concurrent processing
- Increase `WORKER_BUFFER_SIZE` for larger queues
- Allocate more memory/CPU to worker containers

## Error Handling

### Worker Failures
- Individual job failures are logged but don't stop processing
- Worker restart automatically resumes operation
- Jobs in progress during restart may need resubmission

### Database Failures
- Worker logs database connection errors
- Failed insertions are logged for debugging
- Consider implementing retry logic for production use

## Development

### Adding New Features
1. Add functionality to worker package
2. Update HTTP handlers in cmd/worker/main.go
3. Add corresponding API client code in server package
4. Add comprehensive tests

### Testing Strategy
- Unit tests for worker logic
- Integration tests for HTTP communication
- Load tests for performance validation
- Concurrency tests for race condition detection
