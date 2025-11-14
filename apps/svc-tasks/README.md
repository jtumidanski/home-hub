# svc-tasks - Task Management Service

Microservice for managing user tasks within the Home Hub platform.

## Overview

The `svc-tasks` service provides REST APIs for CRUD operations on tasks. Tasks are associated with users and households, with support for daily task management, status tracking, and completion timestamps.

## Features

- **Task CRUD Operations**: Create, read, update, and delete tasks
- **Status Management**: Track task status (incomplete/complete)
- **Date-based Organization**: Associate tasks with specific days
- **User Authorization**: Users can only access their own tasks
- **Multi-tenancy Ready**: Household-level isolation prepared for future gateway integration
- **Time Tracking**: Automatic timestamps for creation and completion

## API Endpoints

All endpoints follow JSON:API specification.

### List Tasks
```
GET /api/tasks
GET /api/tasks?day=2025-11-14
GET /api/tasks?status=incomplete
```

Returns all tasks for the authenticated user, optionally filtered by day or status.

**Example Response:**
```json
{
  "data": [
    {
      "type": "tasks",
      "id": "uuid",
      "attributes": {
        "userId": "uuid",
        "householdId": "uuid",
        "day": "2025-11-14",
        "title": "Buy groceries",
        "description": "Milk, eggs, bread",
        "status": "incomplete",
        "createdAt": "2025-11-14T10:00:00Z",
        "completedAt": null,
        "updatedAt": "2025-11-14T10:00:00Z"
      }
    }
  ]
}
```

### Create Task
```
POST /api/tasks
```

**Request Body:**
```json
{
  "data": {
    "type": "tasks",
    "attributes": {
      "day": "2025-11-14",
      "title": "Buy groceries",
      "description": "Milk, eggs, bread"
    }
  }
}
```

**Response:** 201 Created with task object

### Get Task
```
GET /api/tasks/{id}
```

Returns a single task by ID. Returns 404 if not found or 403 if user doesn't own the task.

### Update Task
```
PATCH /api/tasks/{id}
```

**Request Body:**
```json
{
  "data": {
    "type": "tasks",
    "id": "uuid",
    "attributes": {
      "title": "Buy groceries and cook dinner",
      "status": "complete"
    }
  }
}
```

**Response:** 200 OK with updated task object

### Delete Task
```
DELETE /api/tasks/{id}
```

**Response:** 204 No Content

### Complete Task
```
POST /api/tasks/{id}/complete
```

Convenience endpoint to mark a task as complete. Sets status to "complete" and records completion timestamp.

**Response:** 200 OK with updated task object

## Architecture

### Domain-Driven Design

The service follows strict DDD patterns:

```
task/
├── model.go       # Immutable domain model
├── entity.go      # GORM database entity
├── builder.go     # Fluent builder with validation
├── state.go       # Status enum (incomplete, complete)
├── provider.go    # Data access layer (lazy evaluation)
├── processor.go   # Business logic (pure functions)
├── rest.go        # JSON:API request/response models
└── resource.go    # HTTP handlers and route registration
```

### Database Schema

```sql
CREATE TABLE tasks (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    household_id UUID NOT NULL,
    day DATE NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'incomplete',
    created_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    updated_at TIMESTAMP NOT NULL,
    FOREIGN KEY (household_id) REFERENCES households(id),
    INDEX idx_tasks_user_day (user_id, day),
    INDEX idx_tasks_household_status (household_id, status)
);
```

## Development

### Prerequisites

- Go 1.24+
- PostgreSQL 16
- Docker & Docker Compose (for containerized development)

### Local Development

1. **Using Docker Compose:**
   ```bash
   docker-compose up svc-tasks
   ```

2. **Using Go workspace:**
   ```bash
   go run ./apps/svc-tasks
   ```

### Environment Variables

```bash
DB_HOST=localhost           # Database host
DB_PORT=5432                # Database port
DB_NAME=svc-tasks           # Database name
DB_USER=postgres            # Database user
DB_PASSWORD=postgres        # Database password
SERVICE_PORT=8080           # Service HTTP port
SERVICE_PREFIX=/api/        # API path prefix
LOG_LEVEL=info              # Logging level
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317  # Tracing endpoint
```

### Build

```bash
# Build with Go workspace
go build -o bin/svc-tasks ./apps/svc-tasks

# Build Docker image
docker-compose build svc-tasks
```

### Testing

```bash
# Run unit tests
go test ./apps/svc-tasks/task/...

# Run with coverage
go test -cover ./apps/svc-tasks/task/...

# Run integration tests (requires test database)
go test -tags=integration ./apps/svc-tasks/task/...
```

## Implementation Details

### Validation Rules

- **Title**: Required, 1-255 characters
- **Day**: Required, valid date
- **UserId**: Required, valid UUID
- **HouseholdId**: Required, valid UUID
- **Status**: Must be "incomplete" or "complete"
- **CompletedAt**: Can only be set when status is "complete"

### Authorization

All endpoints require authentication. The service:
- Extracts user ID from auth context (injected by nginx/gateway)
- Verifies user ownership before any read/update/delete operation
- Returns 403 Forbidden if user doesn't own the requested task

### Multi-Tenancy

Tasks are scoped to households via the `household_id` foreign key. When the gateway is implemented, household context will be injected via `X-Household-ID` header for additional isolation.

## Future Enhancements

- **Daily Rollover Worker**: Automatically roll incomplete tasks to the next day
- **Kafka Event Emission**: Publish task lifecycle events (created, completed, deleted)
- **Audit Log**: Track all changes to tasks for accountability
- **Task Recurrence**: Support recurring tasks (daily, weekly, monthly)
- **Task Priority**: Add priority levels and sorting
- **Task Categories**: Organize tasks with tags/categories
- **Task Sharing**: Allow household members to share tasks

## Dependencies

- **GORM**: PostgreSQL ORM
- **gorilla/mux**: HTTP routing
- **api2go**: JSON:API serialization
- **logrus**: Structured logging
- **OpenTelemetry**: Distributed tracing
- **shared-go**: Internal shared libraries (database, auth, server)

## Architecture Compliance

This service follows the architecture specified in `/docs/PROJECT_KNOWLEDGE.md`:
- ✅ Domain-driven design with immutable models
- ✅ Separate database entity from domain model
- ✅ Builder pattern with validation
- ✅ Provider pattern for lazy evaluation
- ✅ Processor pattern for business logic
- ✅ JSON:API compliant REST endpoints
- ✅ Multi-tenancy ready
- ✅ OpenTelemetry tracing integration

## Troubleshooting

### Common Issues

1. **"user is not associated with a household"**
   - Ensure the authenticated user has a household_id set in the users table
   - Create a household and associate the user with it first

2. **"auth context not found"**
   - Ensure nginx or oauth2-proxy is injecting auth headers
   - Check that auth middleware is properly configured

3. **"task not found" or 403 Forbidden**
   - Verify the task exists
   - Ensure the requesting user owns the task

### Database Migrations

Migrations run automatically on service startup using GORM AutoMigrate. To manually trigger:

```bash
# Connect to database and verify schema
docker exec -it hh-db psql -U postgres -d svc-tasks -c "\d tasks"
```

## License

Internal project - Home Hub Platform
