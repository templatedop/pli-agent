# PLI Agent Management System - Developer Guide

## Table of Contents

1. [Introduction](#introduction)
2. [Getting Started](#getting-started)
3. [Project Architecture](#project-architecture)
4. [Code Organization](#code-organization)
5. [Key Concepts](#key-concepts)
6. [Development Workflow](#development-workflow)
7. [API Development](#api-development)
8. [Database Development](#database-development)
9. [Testing Guide](#testing-guide)
10. [Common Tasks](#common-tasks)
11. [Troubleshooting](#troubleshooting)
12. [Best Practices](#best-practices)
13. [Contributing](#contributing)

---

## Introduction

### What is PLI Agent Management System?

The PLI (Postal Life Insurance) Agent Management System is a comprehensive Go-based REST API for managing the complete lifecycle of insurance agents, including:

- Agent profile and contact management
- License management with renewal workflows
- Hierarchical organization structure
- Audit trails and notifications
- Advanced search and dashboard capabilities
- Batch operations and HRMS integration
- Asynchronous workflow orchestration

### Tech Stack

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **Language** | Go 1.22 | Backend programming |
| **Database** | PostgreSQL 13+ | Data persistence |
| **Workflow Engine** | Temporal | Async orchestration |
| **DI Framework** | Uber FX | Dependency injection |
| **API Framework** | n-api-server | REST API handling |
| **Testing** | testify | Unit/integration tests |
| **CI/CD** | GitLab CI, GitHub Actions | Automation |

### Project Goals

- **Reliability**: Single source of truth for agent data
- **Auditability**: Complete audit trail of all changes
- **Scalability**: Handle thousands of agents efficiently
- **Maintainability**: Clean code with comprehensive tests
- **Integration**: Seamless HRMS and external system integration

---

## Getting Started

### Prerequisites

Before you begin, ensure you have the following installed:

```bash
# Required
- Go 1.22 or higher
- PostgreSQL 13 or higher
- Git
- Make (optional but recommended)

# Optional (for development)
- Docker & Docker Compose
- golangci-lint (for linting)
- air (for hot reload)
- Temporal server (for workflow testing)
```

### Installation

#### 1. Clone the Repository

```bash
git clone https://gitlab.cept.gov.in/templatedop/pli-agent.git
cd pli-agent
```

#### 2. Install Dependencies

```bash
# Download Go dependencies
go mod download

# Verify dependencies
go mod verify

# Optional: Install development tools
make install-tools
```

#### 3. Configure Environment

Create a `.env` file in the project root:

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=pli_agent_db
DB_SSLMODE=disable

# Server Configuration
SERVER_PORT=8080
SERVER_HOST=localhost

# Temporal Configuration
TEMPORAL_HOST=localhost:7233
TEMPORAL_NAMESPACE=default

# Logging
LOG_LEVEL=info

# HRMS Webhook
WEBHOOK_SECRET=your-webhook-secret-key
```

#### 4. Set Up Database

```bash
# Start PostgreSQL (if using Docker)
docker run --name postgres-pli \
  -e POSTGRES_PASSWORD=your_password \
  -e POSTGRES_DB=pli_agent_db \
  -p 5432:5432 \
  -d postgres:13

# Run migrations
make migrate-up
```

#### 5. Run the Application

```bash
# Standard run
make run

# Or with hot reload (requires air)
make run-dev

# Or directly with go
go run main.go
```

#### 6. Verify Installation

```bash
# Check if server is running
curl http://localhost:8080/health

# Expected response: {"status":"ok"}
```

### Quick Start Checklist

- [ ] Go 1.22+ installed
- [ ] PostgreSQL running
- [ ] Dependencies downloaded (`go mod download`)
- [ ] Environment configured (`.env` file)
- [ ] Migrations applied (`make migrate-up`)
- [ ] Application running (`make run`)
- [ ] Tests passing (`make test`)

---

## Project Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        API Layer                             │
│  (Handlers - REST endpoints, validation, serialization)     │
└───────────────────────┬─────────────────────────────────────┘
                        │
┌───────────────────────┴─────────────────────────────────────┐
│                    Service Layer                             │
│  (Business logic, orchestration, validation)                │
└───────────────────────┬─────────────────────────────────────┘
                        │
┌───────────────────────┴─────────────────────────────────────┐
│                  Repository Layer                            │
│  (Data access, queries, database operations)                │
└───────────────────────┬─────────────────────────────────────┘
                        │
┌───────────────────────┴─────────────────────────────────────┐
│                   Database Layer                             │
│  (PostgreSQL - tables, indexes, constraints)                │
└─────────────────────────────────────────────────────────────┘

                 Async Operations
┌─────────────────────────────────────────────────────────────┐
│                  Temporal Workflows                          │
│  (Long-running processes, orchestration, retries)           │
└─────────────────────────────────────────────────────────────┘
```

### Request Flow

```
1. HTTP Request → API Handler (handler/)
   ├─ Request validation
   ├─ Authentication/Authorization (future)
   └─ Parameter binding

2. Handler → Repository (repo/postgres/)
   ├─ Business logic execution
   ├─ Data validation
   └─ Database queries

3. Repository → Database
   ├─ SQL query execution
   ├─ Transaction management
   └─ Error handling

4. Response → Client
   ├─ Data transformation
   ├─ Response serialization
   └─ HTTP status codes
```

### Design Patterns

| Pattern | Usage | Location |
|---------|-------|----------|
| **Repository** | Data access abstraction | `repo/postgres/` |
| **Dependency Injection** | Component wiring | `bootstrap/` |
| **Factory** | Test data creation | `testutil/fixtures.go` |
| **Strategy** | Query building | `repo/postgres/*_repository.go` |
| **Decorator** | Logging, metrics | Middleware (future) |

---

## Code Organization

### Directory Structure

```
pli-agent/
├── bootstrap/              # Application bootstrapping and DI
│   └── bootstrapper.go     # Main application initialization
│
├── core/
│   └── domain/             # Domain models and business entities
│       ├── agent_profile.go
│       ├── agent_license.go
│       ├── agent_notification.go
│       └── agent_export.go
│
├── db/
│   ├── migrations/         # Database migration files
│   │   ├── 001_agent_profiles_schema.sql
│   │   ├── 002_agent_licenses_schema.sql
│   │   └── ...
│   └── utility.go          # Database utility functions
│
├── handler/                # HTTP request handlers (controllers)
│   ├── profile_creation.go
│   ├── license_management.go
│   ├── search_dashboard.go
│   ├── batch_webhook.go
│   ├── request.go          # Request DTOs
│   └── response/           # Response DTOs
│       ├── profile.go
│       ├── license.go
│       └── search_dashboard.go
│
├── repo/
│   └── postgres/           # PostgreSQL repository implementations
│       ├── agent_profile.go
│       ├── agent_license.go
│       ├── agent_export.go
│       └── hrms_webhook.go
│
├── workflows/              # Temporal workflow definitions
│   ├── agent_onboarding_workflow.go
│   ├── profile_update_workflow.go
│   └── activities/         # Workflow activities
│       ├── agent_onboarding_activities.go
│       └── profile_update_activities.go
│
├── testutil/               # Testing utilities
│   ├── fixtures.go         # Test data factories
│   ├── mocks.go            # Mock implementations
│   └── helpers.go          # Test helpers
│
├── scripts/                # Build and utility scripts
│   └── run-tests.sh        # Test execution script
│
├── .github/
│   └── workflows/          # GitHub Actions CI/CD
│       └── tests.yml
│
├── docs/                   # Additional documentation
│
├── main.go                 # Application entry point
├── go.mod                  # Go module definition
├── go.sum                  # Go module checksums
├── Makefile                # Build automation
└── .gitlab-ci.yml          # GitLab CI/CD configuration
```

### Package Overview

#### `core/domain`
**Purpose**: Domain models and business entities

**Key Files**:
- `agent_profile.go` - Agent profile entity
- `agent_license.go` - License entity
- `agent_notification.go` - Notification entity
- `agent_export.go` - Export and webhook entities

**Responsibilities**:
- Define data structures
- Define constants
- No business logic (pure data)

#### `handler`
**Purpose**: HTTP request handling and routing

**Key Files**:
- `profile_creation.go` - Profile CRUD operations
- `license_management.go` - License operations
- `search_dashboard.go` - Search and dashboard
- `batch_webhook.go` - Batch and webhook operations

**Responsibilities**:
- Request validation
- Response serialization
- Error handling
- Route registration

#### `repo/postgres`
**Purpose**: Data access layer

**Key Files**:
- `agent_profile.go` - Profile repository
- `agent_license.go` - License repository
- `agent_audit_log.go` - Audit repository
- `agent_export.go` - Export repository

**Responsibilities**:
- SQL query execution
- Transaction management
- Data mapping
- Query optimization

#### `workflows`
**Purpose**: Asynchronous workflow orchestration

**Key Files**:
- `agent_onboarding_workflow.go` - Onboarding flow
- `profile_update_workflow.go` - Update approval flow
- `license_renewal_workflow.go` - License renewal

**Responsibilities**:
- Long-running processes
- Activity coordination
- Error handling and retries
- State management

#### `testutil`
**Purpose**: Testing infrastructure

**Key Files**:
- `fixtures.go` - Test data factories
- `mocks.go` - Mock repositories
- `helpers.go` - Test utilities

**Responsibilities**:
- Consistent test data
- Mock implementations
- Test assertions
- Test utilities

---

## Key Concepts

### 1. Agent Profile

An **Agent Profile** represents an insurance agent in the system.

**Key Fields**:
```go
type AgentProfile struct {
    AgentID       string    // Unique identifier (UUID)
    AgentCode     string    // Business identifier (e.g., "AGT-2024-001")
    AgentType     string    // ADVISOR, COORDINATOR, BRANCH_MANAGER
    FirstName     string    // Agent first name
    LastName      string    // Agent last name
    Status        string    // ACTIVE, SUSPENDED, TERMINATED
    OfficeCode    string    // Office/branch code
    PANNumber     string    // Tax identification
}
```

**Relationships**:
- Has many: Addresses, Contacts, Emails, Licenses
- Has audit logs, notifications
- Can have coordinator (hierarchical)

### 2. Agent License

An **Agent License** represents an insurance license held by an agent.

**Key Fields**:
```go
type AgentLicense struct {
    LicenseID     string    // Unique identifier
    AgentID       string    // Foreign key to agent
    LicenseType   string    // LIFE_INSURANCE, HEALTH_INSURANCE, etc.
    LicenseNumber string    // License number
    IssueDate     time.Time // Issue date
    ExpiryDate    time.Time // Expiry date
    Status        string    // ACTIVE, EXPIRED, REVOKED
}
```

**Business Rules**:
- Auto-deactivation when expired
- Renewal workflow with validation
- Cannot delete active licenses
- Audit trail for all changes

### 3. Audit Logging

All changes to agent data are logged for compliance.

**Key Fields**:
```go
type AgentAuditLog struct {
    AuditID      string    // Unique identifier
    AgentID      string    // Agent being modified
    ActionType   string    // CREATE, UPDATE, DELETE, STATUS_CHANGE
    FieldName    string    // Field that changed
    OldValue     string    // Previous value
    NewValue     string    // New value
    PerformedBy  string    // User who made change
    PerformedAt  time.Time // When change occurred
}
```

**Use Cases**:
- Compliance reporting
- Change history
- Dispute resolution
- Security auditing

### 4. Notifications

Multi-channel notifications for agents and administrators.

**Types**:
- `EMAIL` - Email notifications
- `SMS` - SMS text messages
- `INTERNAL` - In-app notifications

**States**:
- `SENT` - Successfully sent
- `DELIVERED` - Confirmed delivery
- `FAILED` - Send failed
- `READ` - User has read notification

### 5. Workflows

Temporal workflows orchestrate complex, long-running processes.

**Common Workflows**:
1. **Agent Onboarding** - Multi-step onboarding with validations
2. **Profile Update** - Approval workflow for changes
3. **License Renewal** - Automated renewal process
4. **Agent Termination** - Cleanup and deactivation

**Workflow Benefits**:
- Reliability (automatic retries)
- Visibility (state tracking)
- Scalability (distributed execution)
- Maintainability (clear steps)

---

## Development Workflow

### 1. Feature Development Process

```bash
# 1. Create feature branch
git checkout -b feature/add-agent-search

# 2. Make changes
# - Add domain models
# - Implement repository methods
# - Create handler functions
# - Write tests

# 3. Run tests
make test

# 4. Format and lint
make fmt
make vet
make lint

# 5. Commit changes
git add .
git commit -m "feat: Add agent search functionality"

# 6. Push and create PR
git push origin feature/add-agent-search
```

### 2. Adding a New API Endpoint

**Step 1: Define Domain Model** (`core/domain/`)

```go
// core/domain/agent_search.go
package domain

type AgentSearchRequest struct {
    Name       *string `json:"name,omitempty"`
    Status     *string `json:"status,omitempty"`
    OfficeCode *string `json:"office_code,omitempty"`
    Page       int     `json:"page" validate:"min=1"`
    Limit      int     `json:"limit" validate:"min=1,max=100"`
}

type AgentSearchResult struct {
    AgentID    string `json:"agent_id"`
    Name       string `json:"name"`
    Status     string `json:"status"`
    OfficeCode string `json:"office_code"`
}
```

**Step 2: Add Repository Method** (`repo/postgres/`)

```go
// repo/postgres/agent_profile.go
func (r *AgentProfileRepository) Search(
    ctx context.Context,
    filters AgentSearchRequest,
    page, limit int,
) ([]domain.AgentSearchResult, int64, error) {
    // Build query with filters
    query := dblib.Psql.Select("*").From("agent_profiles")

    if filters.Name != nil {
        query = query.Where(sq.ILike{"name": "%" + *filters.Name + "%"})
    }
    if filters.Status != nil {
        query = query.Where(sq.Eq{"status": *filters.Status})
    }

    // Execute query
    sql, args, _ := query.ToSql()
    var results []domain.AgentSearchResult
    err := r.db.Select(ctx, &results, sql, args...)

    return results, total, err
}
```

**Step 3: Create Handler** (`handler/`)

```go
// handler/agent_search.go
package handler

func (h *AgentSearchHandler) SearchAgents(
    sctx *serverRoute.Context,
    request req.SearchAgentsRequest,
) (*resp.SearchAgentsResponse, error) {
    // Call repository
    results, total, err := h.repo.Search(
        sctx.Ctx,
        request.Filters,
        request.Page,
        request.Limit,
    )
    if err != nil {
        return nil, fmt.Errorf("search failed: %w", err)
    }

    // Build response
    return &resp.SearchAgentsResponse{
        Results: results,
        Total:   total,
    }, nil
}
```

**Step 4: Register Route** (`handler/routes.go`)

```go
func (h *AgentSearchHandler) RegisterRoutes() []serverRoute.Route {
    return []serverRoute.Route{
        serverRoute.NewRoute("POST", "/agents/search", h.SearchAgents),
    }
}
```

**Step 5: Write Tests** (`handler/agent_search_test.go`)

```go
func TestSearchAgents_Success(t *testing.T) {
    mockRepo := new(testutil.MockProfileRepository)
    handler := NewAgentSearchHandler(mockRepo)

    mockRepo.On("Search", mock.Anything, mock.Anything, 1, 20).
        Return([]domain.AgentSearchResult{{AgentID: "123"}}, int64(1), nil)

    response, err := handler.SearchAgents(ctx, request)

    assert.NoError(t, err)
    assert.Len(t, response.Results, 1)
    mockRepo.AssertExpectations(t)
}
```

### 3. Database Migration Process

**Step 1: Create Migration File**

```bash
# Create new migration (use sequential numbering)
touch db/migrations/009_add_agent_tags.sql
```

**Step 2: Write Migration**

```sql
-- db/migrations/009_add_agent_tags.sql

-- Up Migration
CREATE TABLE IF NOT EXISTS agent_tags (
    tag_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id TEXT NOT NULL REFERENCES agent_profiles(agent_id) ON DELETE CASCADE,
    tag_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by TEXT NOT NULL,

    UNIQUE(agent_id, tag_name)
);

CREATE INDEX IF NOT EXISTS idx_agent_tags_agent_id ON agent_tags(agent_id);
CREATE INDEX IF NOT EXISTS idx_agent_tags_tag_name ON agent_tags(tag_name);

-- Down Migration (optional, for rollback)
-- DROP TABLE IF EXISTS agent_tags;
```

**Step 3: Run Migration**

```bash
# Apply migration
make migrate-up

# Check status
make migrate-status

# Rollback if needed
make migrate-down
```

### 4. Testing Workflow

```bash
# Run all tests
make test

# Run specific package
go test ./handler -v

# Run specific test
go test ./handler -run TestSearchAgents_Success -v

# Run with coverage
make test-coverage

# Run with race detection
make test-race

# Run benchmarks
make test-bench
```

---

## API Development

### Request/Response Pattern

#### Request DTOs (`handler/request.go`)

```go
package handler

// SearchAgentsRequest for AGT-022
type SearchAgentsRequest struct {
    AgentID    *string `json:"agent_id,omitempty"`
    Name       *string `json:"name,omitempty"`
    Status     *string `json:"status,omitempty" validate:"omitempty,oneof=ACTIVE SUSPENDED TERMINATED"`
    OfficeCode *string `json:"office_code,omitempty"`
    Page       *int    `json:"page,omitempty" validate:"omitempty,min=1"`
    Limit      *int    `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}
```

**Key Points**:
- Use pointers for optional fields
- Add validation tags
- Document with comments

#### Response DTOs (`handler/response/`)

```go
package response

// SearchAgentsResponse for AGT-022
type SearchAgentsResponse struct {
    Results    []domain.AgentSearchResult `json:"results"`
    Pagination domain.PaginationMetadata  `json:"pagination"`
}
```

### Error Handling

```go
// Good: Wrap errors with context
if err != nil {
    return nil, fmt.Errorf("failed to search agents: %w", err)
}

// Bad: Lose error context
if err != nil {
    return nil, err
}
```

### Common HTTP Status Codes

| Code | Usage | Example |
|------|-------|---------|
| 200 | Success | GET, PUT successful |
| 201 | Created | POST created resource |
| 400 | Bad Request | Invalid input |
| 404 | Not Found | Resource doesn't exist |
| 409 | Conflict | Duplicate entry |
| 500 | Server Error | Database error |

---

## Database Development

### Query Optimization Guidelines

#### 1. Single Round-Trip Queries

**Good** - Single query with JSON aggregation:
```go
sql := `
    SELECT
        p.*,
        COALESCE(json_agg(DISTINCT a.*) FILTER (WHERE a.address_id IS NOT NULL), '[]') AS addresses,
        COALESCE(json_agg(DISTINCT c.*) FILTER (WHERE c.contact_id IS NOT NULL), '[]') AS contacts
    FROM agent_profiles p
    LEFT JOIN agent_addresses a ON p.agent_id = a.agent_id
    LEFT JOIN agent_contacts c ON p.agent_id = c.agent_id
    WHERE p.agent_id = $1
    GROUP BY p.agent_id
`
```

**Bad** - Multiple queries (N+1 problem):
```go
// Query 1: Get profile
profile, _ := repo.GetProfile(agentID)

// Query 2: Get addresses (separate query)
addresses, _ := repo.GetAddresses(agentID)

// Query 3: Get contacts (separate query)
contacts, _ := repo.GetContacts(agentID)
```

#### 2. Indexing Strategy

```sql
-- Primary key index (automatic)
CREATE TABLE agent_profiles (
    agent_id TEXT PRIMARY KEY
);

-- Foreign key index (for joins)
CREATE INDEX idx_agent_addresses_agent_id ON agent_addresses(agent_id);

-- Composite index (for common query patterns)
CREATE INDEX idx_agent_profiles_status_office
ON agent_profiles(status, office_code);

-- Partial index (for specific conditions)
CREATE INDEX idx_agent_profiles_active
ON agent_profiles(agent_id)
WHERE status = 'ACTIVE' AND deleted_at IS NULL;

-- Trigram index (for fuzzy search)
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_agent_name_trgm
ON agent_profiles USING gin ((first_name || ' ' || last_name) gin_trgm_ops);
```

#### 3. Recursive Queries (Hierarchies)

```sql
-- Get agent hierarchy chain
WITH RECURSIVE hierarchy AS (
    -- Base case: Start with target agent
    SELECT agent_id, agent_code, name, agent_type,
           advisor_coordinator_id, 1 AS level
    FROM agent_profiles
    WHERE agent_id = $1

    UNION ALL

    -- Recursive case: Get parent coordinator
    SELECT p.agent_id, p.agent_code, p.name, p.agent_type,
           p.advisor_coordinator_id, h.level + 1
    FROM agent_profiles p
    INNER JOIN hierarchy h ON p.agent_id = h.advisor_coordinator_id
)
SELECT * FROM hierarchy ORDER BY level ASC;
```

### Transaction Management

```go
// Start transaction
tx, err := r.db.Begin(ctx)
if err != nil {
    return err
}
defer tx.Rollback() // Rollback if not committed

// Execute operations
_, err = tx.Exec(ctx, sql1, args1...)
if err != nil {
    return err
}

_, err = tx.Exec(ctx, sql2, args2...)
if err != nil {
    return err
}

// Commit transaction
return tx.Commit()
```

---

## Testing Guide

### Unit Testing

#### Testing Handlers

```go
func TestSearchAgents_Success(t *testing.T) {
    // Setup: Create mocks
    mockRepo := new(testutil.MockProfileRepository)
    handler := NewAgentSearchHandler(mockRepo)

    // Setup: Define expected behavior
    expectedResults := []domain.AgentSearchResult{
        {AgentID: "123", Name: "John Doe"},
    }
    mockRepo.On("Search", mock.Anything, mock.Anything, 1, 20).
        Return(expectedResults, int64(1), nil)

    // Execute: Call handler
    request := req.SearchAgentsRequest{
        Name: stringPtr("John"),
        Page: intPtr(1),
        Limit: intPtr(20),
    }
    response, err := handler.SearchAgents(ctx, request)

    // Assert: Verify results
    assert.NoError(t, err)
    assert.NotNil(t, response)
    assert.Len(t, response.Results, 1)
    assert.Equal(t, "123", response.Results[0].AgentID)

    // Assert: Verify mock expectations
    mockRepo.AssertExpectations(t)
}
```

#### Testing Repository Logic

```go
func TestValidateExportConfig(t *testing.T) {
    tests := []struct {
        name     string
        config   *domain.AgentExportConfig
        wantErr  bool
        errMsg   string
    }{
        {
            name: "valid config",
            config: &domain.AgentExportConfig{
                ExportName:   "Test Export",
                OutputFormat: "EXCEL",
                CreatedBy:    "user123",
            },
            wantErr: false,
        },
        {
            name: "empty export name",
            config: &domain.AgentExportConfig{
                ExportName:   "",
                OutputFormat: "EXCEL",
                CreatedBy:    "user123",
            },
            wantErr: true,
            errMsg:  "export name is required",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateExportConfig(tt.config)
            if tt.wantErr {
                assert.Error(t, err)
                if tt.errMsg != "" {
                    assert.Contains(t, err.Error(), tt.errMsg)
                }
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Test Data Fixtures

```go
// Use fixtures for consistent test data
func TestExample(t *testing.T) {
    // Create test profile
    profile := testutil.CreateTestAgentProfile("test-id")

    // Customize as needed
    profile.Status = "ACTIVE"
    profile.OfficeCode = "OFF-001"

    // Use in test
    mockRepo.On("FindByID", mock.Anything, "test-id").
        Return(profile, nil)
}
```

### Running Tests

```bash
# All tests
make test

# Specific package
go test ./handler -v

# Specific test
go test ./handler -run TestSearchAgents -v

# With coverage
make test-coverage
open coverage.html

# With race detection
make test-race

# Benchmarks
make test-bench
```

---

## Common Tasks

### 1. Adding a New Field to Agent Profile

**Step 1**: Update domain model
```go
// core/domain/agent_profile.go
type AgentProfile struct {
    // ... existing fields ...
    NewField string `json:"new_field" db:"new_field"`
}
```

**Step 2**: Create migration
```sql
-- db/migrations/010_add_new_field.sql
ALTER TABLE agent_profiles ADD COLUMN new_field TEXT;
```

**Step 3**: Update repository queries
```go
// repo/postgres/agent_profile.go
// Update INSERT and UPDATE queries to include new field
```

**Step 4**: Run migration
```bash
make migrate-up
```

### 2. Adding Validation Rules

```go
// handler/request.go
type CreateAgentRequest struct {
    FirstName string `json:"first_name" validate:"required,min=2,max=100"`
    LastName  string `json:"last_name" validate:"required,min=2,max=100"`
    PANNumber string `json:"pan_number" validate:"required,len=10,alphanum"`
    Email     string `json:"email" validate:"required,email"`
}
```

### 3. Adding a New Workflow

**Step 1**: Define workflow
```go
// workflows/new_workflow.go
func NewWorkflow(ctx workflow.Context, input NewWorkflowInput) error {
    // Step 1: Validate
    err := workflow.ExecuteActivity(ctx, ValidateActivity, input).Get(ctx, nil)
    if err != nil {
        return err
    }

    // Step 2: Process
    var result ProcessResult
    err = workflow.ExecuteActivity(ctx, ProcessActivity, input).Get(ctx, &result)
    if err != nil {
        return err
    }

    // Step 3: Notify
    return workflow.ExecuteActivity(ctx, NotifyActivity, result).Get(ctx, nil)
}
```

**Step 2**: Register workflow
```go
// bootstrap/bootstrapper.go
func (b *Bootstrapper) RegisterWorkflows() {
    b.worker.RegisterWorkflow(workflows.NewWorkflow)
}
```

### 4. Debugging Database Queries

```go
// Enable query logging in development
sql, args, _ := query.ToSql()
log.Debug("SQL: %s, Args: %v", sql, args)

// Use EXPLAIN ANALYZE in PostgreSQL
EXPLAIN ANALYZE
SELECT * FROM agent_profiles WHERE status = 'ACTIVE';
```

---

## Troubleshooting

### Common Issues

#### 1. "Module not found" errors

```bash
# Solution: Download dependencies
go mod download
go mod tidy
```

#### 2. Database connection errors

```bash
# Check PostgreSQL is running
docker ps | grep postgres

# Verify connection settings in .env
DB_HOST=localhost
DB_PORT=5432

# Test connection
psql -h localhost -U postgres -d pli_agent_db
```

#### 3. Migration failures

```bash
# Check migration status
make migrate-status

# Rollback last migration
make migrate-down

# Force migration version
# (Use with caution!)
```

#### 4. Test failures

```bash
# Run with verbose output
go test ./handler -v -run TestSearchAgents

# Check for race conditions
go test ./handler -race

# Clear test cache
go clean -testcache
```

#### 5. Port already in use

```bash
# Find process using port 8080
lsof -i :8080

# Kill process
kill -9 <PID>

# Or change port in .env
SERVER_PORT=8081
```

---

## Best Practices

### Code Style

1. **Follow Go conventions**
   - Use `gofmt` for formatting
   - Use `golangci-lint` for linting
   - Follow [Effective Go](https://golang.org/doc/effective_go)

2. **Naming conventions**
   ```go
   // Good
   var agentID string
   func GetAgentProfile() {}
   type AgentRepository interface {}

   // Bad
   var agent_id string
   func get_agent_profile() {}
   type agentRepository interface {}
   ```

3. **Error handling**
   ```go
   // Good: Always check errors
   result, err := repo.GetAgent(id)
   if err != nil {
       return fmt.Errorf("failed to get agent: %w", err)
   }

   // Bad: Ignoring errors
   result, _ := repo.GetAgent(id)
   ```

### Database Best Practices

1. **Always use parameterized queries**
   ```go
   // Good
   query := "SELECT * FROM agents WHERE agent_id = $1"
   db.Query(query, agentID)

   // Bad (SQL injection risk)
   query := fmt.Sprintf("SELECT * FROM agents WHERE agent_id = '%s'", agentID)
   ```

2. **Use transactions for multi-step operations**
3. **Add indexes for frequently queried columns**
4. **Use EXPLAIN ANALYZE to optimize queries**

### Security Best Practices

1. **Validate all user input**
2. **Use HTTPS in production**
3. **Sanitize database outputs**
4. **Implement rate limiting**
5. **Log security events**

### Performance Best Practices

1. **Minimize database round trips**
2. **Use connection pooling**
3. **Implement caching where appropriate**
4. **Profile performance bottlenecks**
5. **Use indexes strategically**

---

## Contributing

### Git Workflow

```bash
# 1. Create feature branch
git checkout -b feature/your-feature-name

# 2. Make changes and commit
git add .
git commit -m "feat: Add your feature"

# 3. Push to remote
git push origin feature/your-feature-name

# 4. Create Pull Request
# - Fill out PR template
# - Request reviews
# - Address feedback

# 5. Merge after approval
```

### Commit Message Format

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types**:
- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `test:` - Test additions/changes
- `refactor:` - Code refactoring
- `perf:` - Performance improvements
- `chore:` - Build/tooling changes

**Examples**:
```
feat(api): Add agent search endpoint

- Implement search with filters
- Add pagination support
- Include test coverage

Closes #123
```

### Pull Request Checklist

- [ ] Code follows Go conventions
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] No compiler warnings
- [ ] All tests passing
- [ ] No merge conflicts

---

## Additional Resources

### Documentation
- [Go Documentation](https://golang.org/doc/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Temporal Documentation](https://docs.temporal.io/)
- [testify Documentation](https://pkg.go.dev/github.com/stretchr/testify)

### Internal Documentation
- [TESTING_DOCUMENTATION.md](./TESTING_DOCUMENTATION.md)
- [PHASE_11_TESTING_SUMMARY.md](./PHASE_11_TESTING_SUMMARY.md)
- [ENHANCEMENT_SUGGESTIONS.md](./ENHANCEMENT_SUGGESTIONS.md)
- [PROJECT_COMPLETION_SUMMARY.md](./PROJECT_COMPLETION_SUMMARY.md)

### Code Examples
- See `testutil/` for test patterns
- See `handler/` for API patterns
- See `repo/postgres/` for database patterns
- See `workflows/` for workflow patterns

---

## Getting Help

### Internal Resources
1. **Documentation** - Check docs/ directory
2. **Code Examples** - Review existing implementations
3. **Tests** - Look at test files for usage examples

### External Resources
1. **Go Forum** - https://forum.golangbridge.org/
2. **Stack Overflow** - Tag: `go`
3. **PostgreSQL** - https://www.postgresql.org/support/

### Team Communication
- **Slack/Teams** - #pli-agent-dev channel
- **Email** - dev-team@example.com
- **Issue Tracker** - GitLab Issues

---

## Appendix

### Environment Variables Reference

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DB_HOST` | Database host | localhost | Yes |
| `DB_PORT` | Database port | 5432 | Yes |
| `DB_USER` | Database user | postgres | Yes |
| `DB_PASSWORD` | Database password | - | Yes |
| `DB_NAME` | Database name | pli_agent_db | Yes |
| `SERVER_PORT` | API server port | 8080 | Yes |
| `TEMPORAL_HOST` | Temporal server | localhost:7233 | Yes |
| `LOG_LEVEL` | Logging level | info | No |
| `WEBHOOK_SECRET` | HRMS webhook secret | - | Yes |

### Useful Commands Reference

```bash
# Development
make run                  # Run application
make test                 # Run tests
make fmt                  # Format code
make lint                 # Lint code

# Database
make migrate-up           # Apply migrations
make migrate-down         # Rollback migrations
make migrate-status       # Check status

# Quality
make test-coverage        # Coverage report
make test-race            # Race detection
make quick-check          # Quick validation
make full-check           # Full validation

# Build
make build                # Build binary
make docker-build         # Build Docker image
make clean                # Clean artifacts
```

---

**Last Updated**: January 27, 2026
**Version**: 1.0.0
**Maintainer**: Development Team
