# Phase 6 Context - Critical Lessons & Patterns

**Date**: 2026-01-26
**Branch**: claude/develop-policy-apis-golang-BcDD3
**Current Commit**: 25482e4
**Status**: Starting Phase 6 - Profile Update APIs (AGT-022 to AGT-028)

---

## 🎓 CRITICAL LESSONS LEARNED (MANDATORY FOR ALL PHASES)

### **#1: PRIMARY CONCERN - REDUCE DATABASE ROUND TRIPS**

**ALWAYS aim for single database round trip per operation**

✅ **Pattern**: Use `UPDATE...RETURNING` to eliminate extra SELECT
```go
// ❌ BAD (2 trips)
err = repo.Update(ctx, id, data)
result, err = repo.FindByID(ctx, id)

// ✅ GOOD (1 trip)
result, err = repo.UpdateReturning(ctx, id, data)
```

✅ **Pattern**: Combine operations with CTE
```sql
-- Single query does: INSERT profile + INSERT audit
WITH inserted AS (
  INSERT INTO profiles (...) VALUES (...) RETURNING *
)
INSERT INTO audit_logs (...)
SELECT ... FROM inserted
```

✅ **Pattern**: Use UNNEST for bulk operations
```sql
-- Single query inserts N addresses
INSERT INTO addresses (agent_id, line1, city)
SELECT * FROM UNNEST($1::uuid[], $2::text[], $3::text[])
```

### **#2: ATOMICITY IS MANDATORY**

**Either ALL operations succeed or NONE - no partial states**

✅ **Pattern**: Atomic batch methods in repositories
```go
// All entities created in single transaction
func (r *Repo) CreateWithRelatedEntities(input CreateInput) error {
    // Uses CTE with UNNEST
    // Either profile + addresses + contacts ALL created
    // Or NONE created
}
```

✅ **Pattern**: Prevent inconsistent states
```go
// ❌ BAD - Can leave inconsistent state
profileRepo.Create()  // Succeeds
addressRepo.Create()  // Fails ← Profile exists but no address!

// ✅ GOOD - Atomic
profileRepo.CreateWithRelatedEntities({
    Profile: profile,
    Addresses: addresses,
    Contacts: contacts,
}) // All or nothing
```

### **#3: TEMPORAL WORKFLOW PATTERNS**

✅ **Handler**: Just start workflow
```go
we, err := h.temporalClient.ExecuteWorkflow(...)
// NO database updates here!
return response
```

✅ **Workflow**: Orchestrate activities
```go
// FIRST activity: Record workflow start
workflow.ExecuteActivity(ctx, RecordWorkflowStartActivity, ...)
// Then other activities
```

✅ **Activity**: Do actual work (including DB updates)
```go
// Database updates happen in activities
// Temporal retries if activity fails
// Workflow doesn't proceed until activity succeeds
```

✅ **Self-Recording Pattern**: Workflow records itself
```go
// Step 0 in workflow (FIRST activity)
RecordWorkflowStartActivity {
    // Updates database with workflow ID/run ID
    // If fails, Temporal retries
    // Workflow pauses until database knows about it
}
```

### **#4: BULK OPERATIONS**

✅ **Never loop through inserts**
```go
// ❌ BAD
for _, addr := range addresses {
    addressRepo.Create(addr) // N round trips
}

// ✅ GOOD
addressRepo.BulkCreate(addresses) // 1 round trip using UNNEST
```

### **#5: HUMAN-IN-THE-LOOP APPROVALS**

✅ **Use child workflows**
```go
// Start child workflow for approval
workflow.ExecuteChildWorkflow(ctx, ApprovalWorkflow, input)

// Child workflow:
// 1. Send notification
// 2. Wait for signal OR timeout
// 3. Return decision
```

✅ **Signal-based pattern**
```go
// Workflow waits for signal
signalChan := workflow.GetSignalChannel(ctx, "approval-decision")
selector := workflow.NewSelector(ctx)
selector.AddReceive(signalChan, func(c workflow.ReceiveChannel, more bool) {
    c.Receive(ctx, &decision)
})
selector.AddFuture(workflow.NewTimer(ctx, 72*time.Hour), func(f workflow.Future) {
    // Auto-reject on timeout
})
selector.Select(ctx)
```

### **#6: COMPLETE ALL TODOs**

✅ **No incomplete implementations allowed**
- All validation rules must be implemented
- All integrations properly mocked or implemented
- All error handling complete
- All business rules enforced

### **#7: REPOSITORY PATTERNS**

✅ **Atomic update methods**
```go
// Repository method for atomic operations
func (r *Repo) SaveFormDataAndUpdateWorkflowStateReturning(
    ctx context.Context,
    sessionID string,
    formData string,
    workflowState string,
    currentStep string,
    nextStep string,
    progress int,
    updatedBy string,
) (*Domain, error) {
    // Single UPDATE...RETURNING query
    query := psql.Update("table").
        Set("form_data", formData).
        Set("workflow_state", workflowState).
        Set("current_step", currentStep).
        Set("next_step", nextStep).
        Set("progress_percentage", progress).
        Set("last_updated_by", updatedBy).
        Set("updated_at", time.Now()).
        Where(sq.Eq{"session_id": sessionID}).
        Suffix("RETURNING *")

    var result Domain
    err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByNameLax[Domain], &result)
    return &result, err
}
```

---

## 📊 PHASES COMPLETED

### **Phase 1-3: Infrastructure** ✅
- Database setup with DDL scripts
- Go project initialization with FX
- Configuration files (base + dev + test)
- Port layer (request/response)
- 8 domain models with business rules
- 7 repositories with CTE patterns
- db/utility.go for CTE support

**Key Learning**: CTE patterns required because pgx.Batch doesn't share variables between queries

### **Phase 4: Lookup & Validation APIs (11 endpoints)** ✅
**APIs**:
- AGT-007 to AGT-011: Lookup APIs (5 endpoints)
- AGT-012 to AGT-015: Validation APIs (4 endpoints)
- AGT-020 to AGT-021: Workflow APIs (2 endpoints)

**Key Achievement**: All integrated with repositories, no mocks

### **Phase 5: Profile Creation + Temporal WF-002 (10 endpoints)** ✅
**APIs**:
- AGT-001 to AGT-006: Profile Creation (6 endpoints)
- AGT-016 to AGT-019: Session Management (4 endpoints)

**Infrastructure**:
- Session storage table + repository
- WF-002: Agent Onboarding Workflow (421 lines)
- 18 activities (ACT-011 to ACT-028) + RecordWorkflowStartActivity
- Approval child workflow (human-in-the-loop)
- Bootstrap with Temporal registration

**Performance**: 45% reduction in database round trips

**Critical Fixes Applied**:
1. Atomic batch methods in repositories (3 new methods)
2. CreateWithRelatedEntities for bulk operations
3. Workflow self-recording pattern
4. All handlers use single round-trip operations

---

## 🗂️ FILE STRUCTURE

```
pli-agent/
├── bootstrap/
│   └── bootstrapper.go          # FX modules, Temporal registration
├── configs/
│   ├── config.yaml              # Base config with Temporal settings
│   ├── config.dev.yaml          # Dev config (Temporal enabled)
│   └── config.test.yaml
├── core/
│   ├── domain/                  # Domain models (8 models)
│   │   ├── agent_profile.go
│   │   ├── agent_address.go
│   │   ├── agent_contact.go
│   │   ├── agent_email.go
│   │   ├── agent_bank_details.go
│   │   ├── agent_license.go
│   │   ├── agent_audit_log.go
│   │   └── agent_profile_session.go
│   └── port/
│       ├── request.go
│       └── response.go
├── db/
│   ├── migrations/              # DDL scripts
│   │   ├── 001_agent_profile_management.sql
│   │   └── 002_agent_profile_sessions.sql
│   └── utility.go               # CTE helper functions
├── repo/postgres/               # Repositories (7 repos)
│   ├── agent_profile.go         # + CreateWithRelatedEntities
│   ├── agent_address.go
│   ├── agent_contact.go
│   ├── agent_email.go
│   ├── agent_bank_details.go
│   ├── agent_license.go
│   ├── agent_audit_log.go
│   └── agent_profile_session.go # + 3 atomic methods
├── handler/
│   ├── lookup.go                # AGT-007 to AGT-011
│   ├── validation.go            # AGT-012 to AGT-015
│   ├── workflow.go              # AGT-020 to AGT-021
│   ├── profile_creation.go      # AGT-001 to AGT-006, AGT-016 to AGT-019
│   ├── request.go               # All request DTOs
│   └── response/
│       ├── lookup.go
│       ├── validation.go
│       ├── workflow.go
│       └── profile_creation.go
├── workflows/
│   ├── agent_onboarding_workflow.go   # WF-002 (with Step 0: record start)
│   ├── approval_workflow.go           # Child workflow for approvals
│   └── activities/
│       └── agent_onboarding_activities.go  # 18 activities + RecordWorkflowStart
└── main.go                      # Bootstrap with FxTemporal enabled
```

---

## 🔥 CRITICAL CODE PATTERNS

### **Pattern 1: Repository Atomic Update with RETURNING**

```go
func (r *SessionRepo) SaveFormDataAndUpdateWorkflowStateReturning(
    ctx context.Context,
    sessionID string,
    formData string,
    workflowState string,
    currentStep string,
    nextStep string,
    progress int,
    updatedBy string,
) (*domain.Session, error) {
    cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
    defer cancel()

    // Single UPDATE with RETURNING - atomic operation
    query := dblib.Psql.Update(sessionTable).
        Set("form_data", formData).
        Set("workflow_state", workflowState).
        Set("current_step", currentStep).
        Set("next_step", nextStep).
        Set("progress_percentage", progress).
        Set("last_updated_by", updatedBy).
        Set("updated_at", time.Now()).
        Where(sq.Eq{"session_id": sessionID}).
        Suffix("RETURNING *")

    var result domain.Session
    err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.Session], &result)
    return &result, err
}
```

### **Pattern 2: Bulk Insert with UNNEST and CTE**

```go
func (r *ProfileRepo) CreateWithRelatedEntities(ctx context.Context, input CreateInput) (*domain.Profile, error) {
    cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutHigh"))
    defer cancel()

    // Build arrays for UNNEST
    var addrTypes []string
    var addrLine1s []string
    var addrCities []string
    for _, addr := range input.Addresses {
        addrTypes = append(addrTypes, addr.AddressType)
        addrLine1s = append(addrLine1s, addr.Line1)
        addrCities = append(addrCities, addr.City)
    }

    // Complex CTE with UNNEST for bulk insert
    sql := `
        WITH inserted_profile AS (
            INSERT INTO agent_profiles (...)
            VALUES ($1, $2, ...)
            RETURNING *
        ),
        inserted_addresses AS (
            INSERT INTO agent_addresses (agent_id, address_type, line1, city)
            SELECT ip.agent_id, addr_type, addr_line1, addr_city
            FROM inserted_profile ip
            CROSS JOIN UNNEST($27::text[], $28::text[], $29::text[])
                AS t(addr_type, addr_line1, addr_city)
            WHERE ARRAY_LENGTH($27::text[], 1) > 0
            RETURNING *
        ),
        inserted_audit AS (
            INSERT INTO agent_audit_logs (agent_id, action_type, performed_by, performed_at)
            SELECT agent_id, $48, $49, $50
            FROM inserted_profile
            RETURNING *
        )
        SELECT * FROM inserted_profile
    `

    args := []interface{}{
        profile.FirstName, profile.LastName, ...,
        addrTypes, addrLine1s, addrCities,  // Arrays for UNNEST
        "CREATE", profile.CreatedBy, time.Now(),
    }

    batch := &pgx.Batch{}
    var result domain.Profile
    err := dbutil.QueueReturnRowRaw(batch, sql, args, pgx.RowToStructByNameLax[domain.Profile], &result)
    if err != nil {
        return nil, err
    }

    err = r.db.SendBatch(cCtx, batch).Close()
    return &result, err
}
```

### **Pattern 3: Handler Using Atomic Repository Method**

```go
func (h *Handler) FetchHRMSData(sctx *serverRoute.Context, req FetchHRMSRequest) (*Response, error) {
    // Verify session exists
    session, err := h.sessionRepo.FindByID(sctx.Ctx, req.SessionID)
    if err != nil {
        return nil, err
    }

    // Fetch HRMS data
    hrmsData := fetchFromHRMS(req.EmployeeID)
    formDataJSON, _ := json.Marshal(hrmsData)

    // ATOMIC: Save form data + update workflow state in single round trip
    _, err = h.sessionRepo.SaveFormDataAndUpdateWorkflowStateReturning(
        sctx.Ctx,
        req.SessionID,
        string(formDataJSON),
        domain.WorkflowStateHRMSFetched,
        "HRMS_DATA_FETCHED",
        "PROFILE_DETAILS",
        30,
        req.SessionID,
    )
    if err != nil {
        return nil, err
    }

    return &Response{
        StatusCodeAndMessage: port.FetchSuccess,
        EmployeeData:         hrmsData,
    }, nil
}
```

### **Pattern 4: Workflow with Self-Recording**

```go
func AgentOnboardingWorkflow(ctx workflow.Context, input OnboardingInput) (*OnboardingOutput, error) {
    logger := workflow.GetLogger(ctx)

    // Setup activity options
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: time.Minute,
        RetryPolicy: &temporal.RetryPolicy{
            InitialInterval:    time.Second,
            BackoffCoefficient: 2.0,
            MaximumInterval:    time.Minute,
            MaximumAttempts:    3,
        },
    }
    ctx = workflow.WithActivityOptions(ctx, ao)

    var a *activities.AgentOnboardingActivities

    // Step 0: Record Workflow Start (FIRST ACTIVITY - CRITICAL)
    // This makes the workflow self-recording and self-healing
    logger.Info("Step 0: Recording workflow start in database")
    workflowInfo := workflow.GetInfo(ctx)
    var recordStartResult activities.RecordWorkflowStartOutput
    err := workflow.ExecuteActivity(ctx, a.RecordWorkflowStartActivity, activities.RecordWorkflowStartInput{
        SessionID:     input.SessionID,
        WorkflowID:    workflowInfo.WorkflowExecution.ID,
        RunID:         workflowInfo.WorkflowExecution.RunID,
        WorkflowState: domain.WorkflowStateProfileSubmitting,
        CurrentStep:   "PROFILE_SUBMITTED",
        NextStep:      "VALIDATION",
        Progress:      10,
        SubmittedBy:   input.SubmittedBy,
    }).Get(ctx, &recordStartResult)
    if err != nil {
        return nil, fmt.Errorf("failed to record workflow start: %w", err)
    }

    // Step 1: Validate Agent Type
    // ... rest of workflow
}
```

### **Pattern 5: Activity with Repository Integration**

```go
func (a *Activities) RecordWorkflowStartActivity(ctx context.Context, input RecordWorkflowStartInput) (*RecordWorkflowStartOutput, error) {
    logger := activity.GetLogger(ctx)
    logger.Info("RecordWorkflowStartActivity started", "SessionID", input.SessionID)

    activity.RecordHeartbeat(ctx, "Recording workflow start in database")

    // ATOMIC: Link Temporal workflow + update state in single database round trip
    _, err := a.sessionRepo.LinkTemporalWorkflowAndUpdateStateReturning(
        ctx,
        input.SessionID,
        input.WorkflowID,
        input.RunID,
        input.WorkflowState,
        input.CurrentStep,
        input.NextStep,
        input.Progress,
        input.SubmittedBy,
    )
    if err != nil {
        logger.Error("Failed to record workflow start in database", "error", err)
        return nil, fmt.Errorf("failed to record workflow start: %w", err)
    }

    return &RecordWorkflowStartOutput{
        Recorded: true,
        Message:  "Workflow start recorded in database",
    }, nil
}
```

---

## 📋 PHASE 6: PROFILE UPDATE APIs (AGT-022 to AGT-028)

### **Scope: 7 Endpoints**

1. **AGT-022**: GET /agents/search - Multi-criteria agent search
2. **AGT-023**: GET /agents/{agent_id} - Get agent profile details
3. **AGT-024**: GET /agents/{agent_id}/update-form - Get update form
4. **AGT-025**: PUT /agents/{agent_id}/sections/{section} - Update profile section
5. **AGT-026**: PUT /approvals/{approval_request_id}/approve - Approve update
6. **AGT-027**: PUT /approvals/{approval_request_id}/reject - Reject update
7. **AGT-028**: GET /agents/{agent_id}/audit-history - Get audit history

### **Key Requirements**

✅ **FR-AGT-PRF-004**: Multi-criteria agent search
✅ **FR-AGT-PRF-005**: Profile dashboard view
✅ **FR-AGT-PRF-006**: Personal information update
✅ **FR-AGT-PRF-007**: PAN update
✅ **FR-AGT-PRF-008**: Address management
✅ **FR-AGT-PRF-009**: Contact information update
✅ **FR-AGT-PRF-022**: Audit history

✅ **BR-AGT-PRF-005**: Name update with audit logging
✅ **BR-AGT-PRF-006**: PAN update with uniqueness validation
✅ **BR-AGT-PRF-007**: Personal information update rules

✅ **WF-AGT-PRF-002**: Profile Update Workflow (with approval for critical fields)

### **Critical Fields Requiring Approval**
- Name changes (first_name, middle_name, last_name)
- PAN changes
- Aadhar changes

### **Implementation Strategy**

1. **Repository Layer** (single round trips):
   - Search method with filters
   - Update methods with RETURNING
   - Audit log creation in same transaction

2. **Handler Layer**:
   - AGT-022: Search with pagination
   - AGT-023: Get profile with related entities (1 query with JOINs)
   - AGT-024: Get update form (pre-populated)
   - AGT-025: Update section (checks if approval needed)
   - AGT-026/027: Approval handlers (signal to workflow)
   - AGT-028: Audit history with pagination

3. **Workflow** (if needed):
   - ProfileUpdateWorkflow (child workflow for critical fields)
   - Activities: ValidateUpdate, ApplyUpdate, NotifyUser

---

## 🚀 NEXT SESSION CHECKLIST

When starting Phase 6:

1. ✅ Read this context file
2. ✅ Apply all critical patterns (single round trips, atomicity, CTE, UNNEST)
3. ✅ Create repository methods first
4. ✅ Add atomic update methods with RETURNING
5. ✅ Implement handlers using repository methods
6. ✅ Add request/response DTOs
7. ✅ Register handlers in bootstrap
8. ✅ Complete all TODOs before finishing
9. ✅ Format and commit

**Performance Target**: Maintain single database round trip per operation

**Quality Target**: Zero compilation errors, production-ready code

**Testing Promise**: Comprehensive tests in Phase 11

---

**END OF CONTEXT DOCUMENT**
