# PLI Agent API - Phase 6 Implementation Context

**Last Updated**: 2026-01-27
**Branch**: `claude/develop-policy-apis-golang-BcDD3`
**Latest Commit**: `e640d68` - Final database optimizations

---

## CRITICAL: Database Optimization Requirements

### ⚠️ GOLDEN RULE: MINIMIZE DATABASE ROUND TRIPS

**User's Explicit Feedback**: "Can you remember what I said. Reduce number of hits to database."

This is the **HIGHEST PRIORITY** requirement. Every repository method and handler must be designed to minimize database round trips.

---

## Database Optimization Patterns (MUST FOLLOW)

### 1. **Use Batch for Multiple Independent Queries**

When you need to execute multiple queries that can run in parallel:

```go
batch := &pgx.Batch{}

// Queue all queries
var result1 Type1
dblib.QueueReturnRow(batch, query1, scanner1, &result1)

var results2 []Type2
dblib.QueueReturn(batch, query2, scanner2, &results2)

// Execute in SINGLE round trip
err := r.db.SendBatch(ctx, batch).Close()
```

**Examples**:
- `Search()`: count query + data query → 1 batch
- `GetHistory()`: count query + data query → 1 batch
- `GetProfileWithRelatedEntities()`: profile + addresses + contacts + emails → 1 batch

### 2. **Use CTE for Dependent Operations**

When operations depend on each other (e.g., capture old values, update, insert audit):

```go
sql := `
    WITH old_data AS (
        SELECT * FROM table WHERE id = $1
    ),
    updated_data AS (
        UPDATE table SET ... WHERE id = $1 RETURNING *
    ),
    audit_insert AS (
        INSERT INTO audit_logs (...)
        SELECT ... FROM old_data
        WHERE old_value IS DISTINCT FROM new_value
    )
    SELECT row_to_json(t.*) FROM updated_data t
`
```

**Examples**:
- `UpdateSectionReturning()`: capture old + update + audit insert → 1 CTE
- `ApproveAndApplyUpdates()`: fetch request + approve → 1 CTE
- `RejectAndReturn()`: fetch request + reject → 1 CTE

### 3. **Use UNNEST for Bulk Operations**

For bulk inserts/operations on arrays:

```go
sql := `
    INSERT INTO table (col1, col2, col3)
    SELECT * FROM UNNEST(
        $1::text[],
        $2::text[],
        $3::text[]
    )
`
r.db.Exec(ctx, sql, array1, array2, array3)
```

**Examples**:
- Bulk audit log insertion in `UpdateSectionReturning()`
- Bulk inserts in `CreateWithRelatedEntities()`

### 4. **Use RETURNING to Avoid Extra SELECT**

Always use RETURNING clause instead of separate SELECT:

```go
// ❌ BAD: 2 round trips
UPDATE table SET ... WHERE id = $1
SELECT * FROM table WHERE id = $1

// ✅ GOOD: 1 round trip
UPDATE table SET ... WHERE id = $1 RETURNING *
```

---

## Phase 6 Performance Metrics (ACHIEVED)

| Endpoint | Method | Database Calls | Technique |
|----------|--------|----------------|-----------|
| AGT-022: Search Agents | Search() | **1** | Batch (count + data) |
| AGT-023: Get Profile | GetProfileWithRelatedEntities() | **1** | Batch (4 queries) |
| AGT-024: Get Update Form | GetUpdateForm() | 1 | Single SELECT |
| AGT-025: Update Section | UpdateSectionReturning() | **1** | CTE (old + update + audit) |
| AGT-026: Approve Update | ApproveProfileUpdate() | **2** | CTE + CTE |
| AGT-027: Reject Update | RejectProfileUpdate() | **1** | CTE |
| AGT-028: Audit History | GetHistory() | **1** | Batch (count + data) |

---

## Critical Issues Fixed in This Session

### Issue 1: UpdateSectionReturning - Separate Audit Insert
**Problem**: Audit INSERT was a separate database call after batch
**Solution**: Combined everything in single CTE with UNNEST
**Reduction**: 2 → 1 round trips

### Issue 2: ApproveProfileUpdate - 3 Separate Calls
**Problem**: FindByID() + UpdateSectionReturning() + Approve()
**Solution**: Created ApproveAndApplyUpdates() CTE method
**Reduction**: 3 → 2 round trips

### Issue 3: RejectProfileUpdate - 2 Separate Calls
**Problem**: FindByID() + Reject()
**Solution**: Created RejectAndReturn() CTE method
**Reduction**: 2 → 1 round trips

### Issue 4: Wrong Database Interface Type
**Problem**: `dblib.XODB` in AgentProfileUpdateRequestRepository
**Solution**: Changed to `dblib.DB`
**Status**: Fixed ✅

---

## Repository Pattern (ESTABLISHED)

### Standard Repository Structure

```go
type XxxRepository struct {
    db  dblib.DB  // ← ALWAYS use dblib.DB, NOT dblib.XODB
    cfg *config.Config
}

func NewXxxRepository(db dblib.DB, cfg *config.Config) *XxxRepository {
    return &XxxRepository{db: db, cfg: cfg}
}
```

### Timeout Pattern

```go
cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
defer cancel()
```

Timeout levels:
- `QueryTimeoutLow`: Simple queries (SELECT by ID, single updates)
- `QueryTimeoutMed`: Complex queries (CTEs, batches, multiple JOINs)
- `QueryTimeoutHigh`: Very complex operations (bulk operations)

### Error Handling Pattern

```go
err := dblib.SelectOne(cCtx, r.db, query, scanner, &result)
if err != nil {
    return nil, fmt.Errorf("failed to [operation]: %w", err)
}
```

Always wrap errors with context using `fmt.Errorf` with `%w`.

---

## Handler Pattern (ESTABLISHED)

### Handler Structure

```go
type XxxHandler struct {
    *serverHandler.Base
    repo1 *repo.Xxx1Repository
    repo2 *repo.Xxx2Repository
}

func NewXxxHandler(repo1 *repo.Xxx1Repository, repo2 *repo.Xxx2Repository) *XxxHandler {
    base := serverHandler.New("Description").SetPrefix("/v1").AddPrefix("")
    return &XxxHandler{
        Base:  base,
        repo1: repo1,
        repo2: repo2,
    }
}
```

### Route Registration

```go
func (h *XxxHandler) Routes() []serverRoute.Route {
    return []serverRoute.Route{
        serverRoute.GET("/path", h.HandlerMethod).Name("Description"),
        serverRoute.POST("/path", h.HandlerMethod).Name("Description"),
    }
}
```

### Handler Method Signature

```go
func (h *XxxHandler) HandlerMethod(
    sctx *serverRoute.Context,
    req RequestDTO,
) (*ResponseDTO, error) {
    log.Info(sctx.Ctx, "Message")

    // Business logic

    return &ResponseDTO{
        StatusCodeAndMessage: port.SuccessMessage,
        // ... fields
    }, nil
}
```

---

## SQL Patterns and Best Practices

### 1. Parameterized Queries (ALWAYS)

```go
// ❌ BAD: SQL injection risk
sql := fmt.Sprintf("SELECT * FROM table WHERE id = '%s'", userInput)

// ✅ GOOD: Parameterized
sql := "SELECT * FROM table WHERE id = $1"
r.db.QueryRow(ctx, sql, userInput)
```

### 2. NULL-Safe Comparisons

Use `IS DISTINCT FROM` for NULL-safe comparisons:

```sql
WHERE old_value IS DISTINCT FROM new_value
```

This handles NULL correctly (NULL != NULL returns false, but NULL IS DISTINCT FROM NULL returns false).

### 3. JSON Handling

Return complex objects as JSON:

```sql
SELECT row_to_json(t.*) FROM (
    SELECT * FROM table WHERE id = $1
) t
```

Parse in Go:

```go
var jsonData []byte
err := r.db.QueryRow(ctx, sql, id).Scan(&jsonData)

var result domain.Type
err = json.Unmarshal(jsonData, &result)
```

### 4. Array Operations with UNNEST

Expand arrays into rows:

```sql
SELECT unnest(ARRAY['val1', 'val2', 'val3'])
```

Use in bulk operations:

```sql
INSERT INTO table (col1, col2)
SELECT * FROM UNNEST($1::text[], $2::int[])
```

---

## Common Mistakes to Avoid

### ❌ Don't Do This

1. **Multiple separate database calls when batch is possible**
2. **Separate SELECT after UPDATE**
3. **Loop inserts instead of bulk**
4. **Not using CTEs for dependent operations**
5. **Using dblib.XODB instead of dblib.DB**

### ✅ Do This Instead

1. **Use batch for parallel queries**
2. **Use RETURNING clause**
3. **Use UNNEST for bulk operations**
4. **Use CTE for dependent operations**
5. **Use dblib.DB consistently**

---

## Critical Fields and Business Rules

### Critical Fields (Require Approval)

```go
criticalFields := map[string]bool{
    "first_name":    true,
    "middle_name":   true,
    "last_name":     true,
    "pan_number":    true,
    "aadhar_number": true,
}
```

### Business Rules Implemented

- **BR-AGT-PRF-005**: Name updates require approval + audit logging
- **BR-AGT-PRF-006**: PAN updates require approval + validation
- **BR-AGT-PRF-022**: Multi-criteria agent search with pagination
- **FR-AGT-PRF-004**: Multi-criteria search functionality
- **FR-AGT-PRF-005**: Profile dashboard view with related entities
- **FR-AGT-PRF-006**: Profile updates with approval workflow
- **FR-AGT-PRF-022**: Audit trail for all changes

---

## Key Takeaways for Future Sessions

1. **ALWAYS minimize database round trips** - This is non-negotiable
2. **Use batch for parallel queries** - count + data, profile + related entities
3. **Use CTE for dependent operations** - capture old, update, audit
4. **Use UNNEST for bulk operations** - batch inserts, array operations
5. **Use RETURNING to avoid extra SELECT** - UPDATE ... RETURNING *
6. **Use dblib.DB, NOT dblib.XODB** - Standard interface type
7. **Parameterize all queries** - Prevent SQL injection
8. **Use IS DISTINCT FROM for NULL-safe comparisons** - Handle NULLs correctly
9. **Return complex objects as JSON when needed** - row_to_json()
10. **Follow established patterns** - Don't reinvent the wheel

---

## Files to Reference

**Critical Pattern Examples**:
- `repo/postgres/agent_profile.go` - Search, GetProfileWithRelatedEntities, UpdateSectionReturning
- `repo/postgres/agent_audit_log.go` - GetHistory with batch
- `repo/postgres/agent_profile_update_request.go` - CTE examples
- `handler/profile_update.go` - Complete handler implementation

**Documentation**:
- `PHASE_6_IMPLEMENTATION_PLAN.md` - Original plan
- `BR_AGT_PRF_016_REINSTATEMENT_WORKFLOW.md` - Workflow details
- `PHASE_8_9_10_CONTEXT.md` - Future phases specifications
- `PHASE_7_ISSUE_ANALYSIS.md` - Issues with Phase 7 branch

---

**⚠️ REMEMBER**: Read this file at the start of every session to maintain context and follow established patterns!
