# PLI Agent Project - Code Patterns Reference

This document captures the key patterns and conventions used in the PLI Agent Management System to be followed in related projects.

## Project Structure Pattern

```
project-root/
├── bootstrap/              # Application initialization
│   └── bootstrapper.go     # DI container setup
├── core/domain/            # Domain models (pure data structures)
├── db/migrations/          # Sequential SQL migrations
├── handler/                # HTTP handlers (controllers)
│   ├── request.go          # Request DTOs
│   └── response/           # Response DTOs
├── repo/postgres/          # Data access layer
├── workflows/              # Temporal workflows
│   └── activities/         # Workflow activities
├── testutil/               # Testing utilities
│   ├── fixtures.go         # Test data factories
│   ├── mocks.go            # Mock implementations
│   └── helpers.go          # Test helpers
├── scripts/                # Utility scripts
├── main.go                 # Entry point
├── Makefile                # Build automation
└── .gitlab-ci.yml          # CI/CD pipeline
```

## Domain Model Pattern

**Location**: `core/domain/*.go`

**Purpose**: Define data structures without business logic

**Example**:
```go
package domain

import (
    "database/sql"
    "time"
)

// Entity represents a business entity
type Entity struct {
    ID          string         `json:"id" db:"id"`
    Name        string         `json:"name" db:"name"`
    Status      string         `json:"status" db:"status"`
    CreatedAt   time.Time      `json:"created_at" db:"created_at"`
    CreatedBy   string         `json:"created_by" db:"created_by"`
    UpdatedAt   sql.NullTime   `json:"updated_at" db:"updated_at"`
    UpdatedBy   sql.NullString `json:"updated_by" db:"updated_by"`
}

// Constants for status values
const (
    StatusActive    = "ACTIVE"
    StatusInactive  = "INACTIVE"
    StatusSuspended = "SUSPENDED"
)
```

**Key Conventions**:
- Use `sql.Null*` types for nullable fields
- Add JSON and DB tags
- Define constants for enums
- No business logic in domain models
- Include audit fields (created_at, created_by, etc.)

## Repository Pattern

**Location**: `repo/postgres/*_repository.go`

**Purpose**: Data access abstraction with single round-trip optimization

**Example**:
```go
package postgres

import (
    "context"
    "fmt"

    sq "github.com/Masterminds/squirrel"
    "pli-agent-api/core/domain"
    dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"
)

const entityTable = "entities"

type EntityRepository struct {
    db dblib.DB
}

func NewEntityRepository(db dblib.DB) *EntityRepository {
    return &EntityRepository{db: db}
}

// FindByID fetches entity with related data in single query
func (r *EntityRepository) FindByID(
    ctx context.Context,
    id string,
) (*domain.Entity, error) {
    // Single query with JSON aggregation for related entities
    sql := `
        SELECT
            e.*,
            COALESCE(
                json_agg(DISTINCT r.*) FILTER (WHERE r.id IS NOT NULL),
                '[]'
            ) AS related
        FROM entities e
        LEFT JOIN related r ON e.id = r.entity_id
        WHERE e.id = $1 AND e.deleted_at IS NULL
        GROUP BY e.id
    `

    var entity domain.Entity
    err := r.db.Get(ctx, &entity, sql, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get entity: %w", err)
    }

    return &entity, nil
}

// Create inserts new entity
func (r *EntityRepository) Create(
    ctx context.Context,
    entity *domain.Entity,
) (*domain.Entity, error) {
    query := dblib.Psql.Insert(entityTable).
        Columns(
            "id", "name", "status",
            "created_at", "created_by",
        ).
        Values(
            entity.ID, entity.Name, entity.Status,
            entity.CreatedAt, entity.CreatedBy,
        ).
        Suffix("RETURNING *")

    sql, args, _ := query.ToSql()

    var created domain.Entity
    err := r.db.Get(ctx, &created, sql, args...)
    if err != nil {
        return nil, fmt.Errorf("failed to create entity: %w", err)
    }

    return &created, nil
}

// Search with filters and pagination
func (r *EntityRepository) Search(
    ctx context.Context,
    filters EntityFilters,
    page, limit int,
) ([]domain.Entity, int64, error) {
    // Build query with filters
    query := dblib.Psql.Select("*").
        From(entityTable).
        Where(sq.Eq{"deleted_at": nil})

    if filters.Name != nil {
        query = query.Where(sq.ILike{"name": "%" + *filters.Name + "%"})
    }
    if filters.Status != nil {
        query = query.Where(sq.Eq{"status": *filters.Status})
    }

    // Add pagination
    offset := (page - 1) * limit
    query = query.Limit(uint64(limit)).Offset(uint64(offset))
    query = query.OrderBy("created_at DESC")

    sql, args, _ := query.ToSql()

    var entities []domain.Entity
    err := r.db.Select(ctx, &entities, sql, args...)
    if err != nil {
        return nil, 0, fmt.Errorf("search failed: %w", err)
    }

    // Get total count
    countSQL := `SELECT COUNT(*) FROM entities WHERE deleted_at IS NULL`
    var total int64
    _ = r.db.Get(ctx, &total, countSQL)

    return entities, total, nil
}
```

**Key Conventions**:
- Single round-trip queries using JSON aggregation
- Always use parameterized queries ($1, $2)
- Wrap errors with context using `fmt.Errorf`
- Use Squirrel for query building
- Soft delete check in WHERE clause
- Return total count for paginated results

## Handler Pattern

**Location**: `handler/*_handler.go`

**Purpose**: HTTP request handling and routing

**Example**:
```go
package handler

import (
    "fmt"

    "pli-agent-api/core/domain"
    req "pli-agent-api/handler/request"
    resp "pli-agent-api/handler/response"
    "pli-agent-api/repo/postgres"

    serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"
)

type EntityHandler struct {
    repo *postgres.EntityRepository
}

func NewEntityHandler(repo *postgres.EntityRepository) *EntityHandler {
    return &EntityHandler{repo: repo}
}

// GetEntity handles GET /entities/:id
func (h *EntityHandler) GetEntity(
    sctx *serverRoute.Context,
    uri req.EntityIDUri,
) (*resp.EntityResponse, error) {
    // Call repository
    entity, err := h.repo.FindByID(sctx.Ctx, uri.EntityID)
    if err != nil {
        return nil, fmt.Errorf("failed to get entity: %w", err)
    }

    // Build response
    return &resp.EntityResponse{
        Entity: entity,
    }, nil
}

// CreateEntity handles POST /entities
func (h *EntityHandler) CreateEntity(
    sctx *serverRoute.Context,
    request req.CreateEntityRequest,
) (*resp.EntityResponse, error) {
    // Validate request
    if request.Name == "" {
        return nil, fmt.Errorf("name is required")
    }

    // Build domain model
    entity := &domain.Entity{
        ID:        uuid.New().String(),
        Name:      request.Name,
        Status:    domain.StatusActive,
        CreatedAt: time.Now(),
        CreatedBy: "system", // TODO: Get from context
    }

    // Call repository
    created, err := h.repo.Create(sctx.Ctx, entity)
    if err != nil {
        return nil, fmt.Errorf("failed to create entity: %w", err)
    }

    return &resp.EntityResponse{
        Entity: created,
    }, nil
}

// SearchEntities handles POST /entities/search
func (h *EntityHandler) SearchEntities(
    sctx *serverRoute.Context,
    request req.SearchEntitiesRequest,
) (*resp.SearchEntitiesResponse, error) {
    // Default pagination
    page := 1
    if request.Page != nil {
        page = *request.Page
    }
    limit := 20
    if request.Limit != nil {
        limit = *request.Limit
    }

    // Call repository
    entities, total, err := h.repo.Search(
        sctx.Ctx,
        request.Filters,
        page,
        limit,
    )
    if err != nil {
        return nil, fmt.Errorf("search failed: %w", err)
    }

    // Build response
    return &resp.SearchEntitiesResponse{
        Results: entities,
        Pagination: domain.PaginationMetadata{
            Page:       page,
            Limit:      limit,
            TotalCount: total,
            TotalPages: (total + int64(limit) - 1) / int64(limit),
        },
    }, nil
}
```

**Key Conventions**:
- One handler per feature area
- Use request/response DTOs (not domain models directly)
- Validate input in handler
- Call repository methods (no direct SQL)
- Wrap errors with context
- Return proper response structures

## Request/Response DTOs

**Request DTOs** (`handler/request.go`):
```go
package handler

// CreateEntityRequest for creating entity
type CreateEntityRequest struct {
    Name   string  `json:"name" validate:"required,min=2,max=200"`
    Status *string `json:"status,omitempty" validate:"omitempty,oneof=ACTIVE INACTIVE"`
}

// SearchEntitiesRequest for searching
type SearchEntitiesRequest struct {
    Name   *string `json:"name,omitempty"`
    Status *string `json:"status,omitempty"`
    Page   *int    `json:"page,omitempty" validate:"omitempty,min=1"`
    Limit  *int    `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// EntityIDUri for URL parameters
type EntityIDUri struct {
    EntityID string `uri:"entity_id" validate:"required,uuid"`
}
```

**Response DTOs** (`handler/response/*.go`):
```go
package response

import "pli-agent-api/core/domain"

// EntityResponse for single entity
type EntityResponse struct {
    Entity *domain.Entity `json:"entity"`
}

// SearchEntitiesResponse for search results
type SearchEntitiesResponse struct {
    Results    []domain.Entity           `json:"results"`
    Pagination domain.PaginationMetadata `json:"pagination"`
}
```

**Key Conventions**:
- Use pointers for optional fields
- Add validation tags
- Separate request and response types
- Document with comments

## Database Migration Pattern

**Location**: `db/migrations/NNN_description.sql`

**Naming**: Sequential numbering (001, 002, 003...)

**Example**:
```sql
-- db/migrations/005_create_entities_table.sql
-- Description: Create entities table with audit fields
-- Author: Developer Name
-- Date: 2026-01-27

-- =============================================================================
-- ENTITIES TABLE
-- =============================================================================

CREATE TABLE IF NOT EXISTS entities (
    -- Primary Key
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Business Fields
    name VARCHAR(200) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    description TEXT,

    -- Foreign Keys
    parent_id UUID REFERENCES entities(id) ON DELETE SET NULL,

    -- Audit Fields
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by TEXT NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE,
    updated_by TEXT,
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- Constraints
    CONSTRAINT chk_entity_status CHECK (status IN ('ACTIVE', 'INACTIVE', 'SUSPENDED'))
);

-- =============================================================================
-- INDEXES
-- =============================================================================

-- Primary lookup index
CREATE INDEX IF NOT EXISTS idx_entities_id ON entities(id) WHERE deleted_at IS NULL;

-- Status filter index
CREATE INDEX IF NOT EXISTS idx_entities_status ON entities(status) WHERE deleted_at IS NULL;

-- Name search index
CREATE INDEX IF NOT EXISTS idx_entities_name ON entities(name) WHERE deleted_at IS NULL;

-- Composite index for common query pattern
CREATE INDEX IF NOT EXISTS idx_entities_status_created
ON entities(status, created_at DESC) WHERE deleted_at IS NULL;

-- Foreign key index
CREATE INDEX IF NOT EXISTS idx_entities_parent_id ON entities(parent_id);

-- Full-text search (if needed)
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX IF NOT EXISTS idx_entities_name_trgm
ON entities USING gin (name gin_trgm_ops);

-- =============================================================================
-- COMMENTS
-- =============================================================================

COMMENT ON TABLE entities IS 'Main entities table for domain management';
COMMENT ON COLUMN entities.status IS 'Entity status: ACTIVE, INACTIVE, SUSPENDED';
COMMENT ON COLUMN entities.deleted_at IS 'Soft delete timestamp';

-- =============================================================================
-- DOWN MIGRATION (Optional - for rollback)
-- =============================================================================

-- DROP TABLE IF EXISTS entities CASCADE;
```

**Key Conventions**:
- Use descriptive file names
- Add header comments with description, author, date
- Group related DDL statements with section headers
- Always use `IF NOT EXISTS` / `IF EXISTS`
- Add indexes for common queries
- Use UUID for primary keys
- Include audit fields (created_at, created_by, updated_at, updated_by, deleted_at)
- Add check constraints for enums
- Comment tables and important columns
- Use `TIMESTAMP WITH TIME ZONE` for timestamps

## Testing Pattern

**Test Fixtures** (`testutil/fixtures.go`):
```go
package testutil

import (
    "database/sql"
    "time"

    "pli-agent-api/core/domain"
    "github.com/google/uuid"
)

// CreateTestEntity creates a test entity with default values
func CreateTestEntity(id string) *domain.Entity {
    if id == "" {
        id = uuid.New().String()
    }

    return &domain.Entity{
        ID:        id,
        Name:      "Test Entity",
        Status:    domain.StatusActive,
        CreatedAt: time.Now(),
        CreatedBy: "test-user",
    }
}
```

**Mock Repositories** (`testutil/mocks.go`):
```go
package testutil

import (
    "context"

    "pli-agent-api/core/domain"
    "github.com/stretchr/testify/mock"
)

type MockEntityRepository struct {
    mock.Mock
}

func (m *MockEntityRepository) FindByID(ctx context.Context, id string) (*domain.Entity, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*domain.Entity), args.Error(1)
}

func (m *MockEntityRepository) Create(ctx context.Context, entity *domain.Entity) (*domain.Entity, error) {
    args := m.Called(ctx, entity)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*domain.Entity), args.Error(1)
}
```

**Unit Tests** (`handler/*_test.go`):
```go
package handler

import (
    "testing"

    "pli-agent-api/testutil"
    req "pli-agent-api/handler/request"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"
)

func TestGetEntity_Success(t *testing.T) {
    // Setup: Create mock repository
    mockRepo := new(testutil.MockEntityRepository)
    handler := NewEntityHandler(mockRepo)

    // Setup: Create test data
    entity := testutil.CreateTestEntity("test-123")

    // Setup: Configure mock expectations
    mockRepo.On("FindByID", mock.Anything, "test-123").
        Return(entity, nil)

    // Execute: Call handler
    ctx := &serverRoute.Context{Ctx: testutil.TestContext(t)}
    uri := req.EntityIDUri{EntityID: "test-123"}
    response, err := handler.GetEntity(ctx, uri)

    // Assert: Verify results
    assert.NoError(t, err)
    assert.NotNil(t, response)
    assert.Equal(t, "test-123", response.Entity.ID)
    assert.Equal(t, "Test Entity", response.Entity.Name)

    // Assert: Verify mock expectations were met
    mockRepo.AssertExpectations(t)
}

func TestGetEntity_NotFound(t *testing.T) {
    mockRepo := new(testutil.MockEntityRepository)
    handler := NewEntityHandler(mockRepo)

    mockRepo.On("FindByID", mock.Anything, "non-existent").
        Return(nil, fmt.Errorf("not found"))

    ctx := &serverRoute.Context{Ctx: testutil.TestContext(t)}
    uri := req.EntityIDUri{EntityID: "non-existent"}
    response, err := handler.GetEntity(ctx, uri)

    assert.Error(t, err)
    assert.Nil(t, response)
    assert.Contains(t, err.Error(), "not found")

    mockRepo.AssertExpectations(t)
}

// Table-driven test example
func TestValidateStatus(t *testing.T) {
    tests := []struct {
        name    string
        status  string
        wantErr bool
    }{
        {"valid active", "ACTIVE", false},
        {"valid inactive", "INACTIVE", false},
        {"invalid status", "INVALID", true},
        {"empty status", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateStatus(tt.status)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

**Key Testing Conventions**:
- Use fixtures for consistent test data
- Mock external dependencies
- Use table-driven tests for multiple scenarios
- Test both success and error cases
- Use descriptive test names: `TestFunction_Scenario_ExpectedResult`
- Assert mock expectations were met
- Aim for 80%+ code coverage

## Temporal Workflow Pattern

**Workflow** (`workflows/entity_workflow.go`):
```go
package workflows

import (
    "time"

    "go.temporal.io/sdk/temporal"
    "go.temporal.io/sdk/workflow"
)

// EntityProcessingWorkflow orchestrates entity processing
func EntityProcessingWorkflow(
    ctx workflow.Context,
    input EntityWorkflowInput,
) error {
    // Set workflow options
    options := workflow.ActivityOptions{
        StartToCloseTimeout: 5 * time.Minute,
        RetryPolicy: &temporal.RetryPolicy{
            MaximumAttempts: 3,
            InitialInterval: 1 * time.Second,
        },
    }
    ctx = workflow.WithActivityOptions(ctx, options)

    // Step 1: Validate
    err := workflow.ExecuteActivity(ctx, ValidateEntityActivity, input).Get(ctx, nil)
    if err != nil {
        return err
    }

    // Step 2: Process
    var result ProcessingResult
    err = workflow.ExecuteActivity(ctx, ProcessEntityActivity, input).Get(ctx, &result)
    if err != nil {
        return err
    }

    // Step 3: Notify
    err = workflow.ExecuteActivity(ctx, NotifyEntityActivity, result).Get(ctx, nil)
    if err != nil {
        return err
    }

    return nil
}
```

**Activities** (`workflows/activities/entity_activities.go`):
```go
package activities

import (
    "context"

    "go.temporal.io/sdk/activity"
)

type EntityActivities struct {
    repo EntityRepository
}

func (a *EntityActivities) ValidateEntityActivity(
    ctx context.Context,
    input EntityWorkflowInput,
) error {
    logger := activity.GetLogger(ctx)
    logger.Info("Validating entity", "entity_id", input.EntityID)

    // Validation logic
    return nil
}
```

## CI/CD Pattern

**Makefile Targets**:
```makefile
test:
    go test ./... -v -coverprofile=coverage.out

test-coverage:
    go test ./... -coverprofile=coverage.out
    go tool cover -html=coverage.out

build:
    go build -o bin/app ./main.go

lint:
    golangci-lint run ./...
```

**GitLab CI** (`.gitlab-ci.yml`):
```yaml
stages:
  - validate
  - test
  - build

unit-tests:
  stage: test
  script:
    - go test ./... -v -coverprofile=coverage.out
    - go tool cover -func=coverage.out
  coverage: '/total:\s+\(statements\)\s+(\d+\.\d+)%/'
```

## Summary

Follow these patterns for consistency:
- ✅ Domain models in `core/domain/`
- ✅ Repositories in `repo/postgres/`
- ✅ Handlers in `handler/`
- ✅ Single round-trip queries
- ✅ Comprehensive testing
- ✅ Clear error handling
- ✅ Audit fields on all tables
- ✅ Soft deletes
- ✅ UUID primary keys
