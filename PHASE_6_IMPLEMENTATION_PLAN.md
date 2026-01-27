# Phase 6 Implementation Plan - Profile Update APIs

**Date**: 2026-01-27
**Branch**: claude/develop-policy-apis-golang-BcDD3
**Status**: Ready to implement
**Scope**: 7 endpoints (AGT-022 to AGT-028)

---

## 📋 SCOPE SUMMARY

### **7 Endpoints to Implement**

| API | Method | Path | Purpose |
|-----|--------|------|---------|
| **AGT-022** | GET | /agents/search | Multi-criteria agent search with pagination |
| **AGT-023** | GET | /agents/{agent_id} | Get complete agent profile with all related entities |
| **AGT-024** | GET | /agents/{agent_id}/update-form | Get pre-populated update form |
| **AGT-025** | PUT | /agents/{agent_id}/sections/{section} | Update profile section (with approval for critical fields) |
| **AGT-026** | PUT | /approvals/{approval_request_id}/approve | Approve profile update |
| **AGT-027** | PUT | /approvals/{approval_request_id}/reject | Reject profile update |
| **AGT-028** | GET | /agents/{agent_id}/audit-history | Get audit history with pagination |

---

## 🎯 IMPLEMENTATION STRATEGY

### **Phase 6A: Repository Layer (New Methods)**

#### **1. Search Method**
```go
// repo/postgres/agent_profile.go
func (r *AgentProfileRepository) Search(
    ctx context.Context,
    filters SearchFilters,
    page, limit int,
) ([]domain.AgentProfile, *PaginationMetadata, error)
```

**Features**:
- Multi-criteria search (agent_id, name, PAN, mobile, email, status, office_code)
- Pagination support
- Single query with JOINs for related data
- Total count in same query (CTE pattern)

**SQL Pattern**:
```sql
WITH filtered AS (
    SELECT p.* FROM agent_profiles p
    LEFT JOIN agent_contacts c ON p.agent_id = c.agent_id
    WHERE
        ($1::text IS NULL OR p.agent_id = $1)
        AND ($2::text IS NULL OR CONCAT(p.first_name, ' ', p.last_name) ILIKE '%' || $2 || '%')
        AND ($3::text IS NULL OR p.pan_number = $3)
        AND ($4::text IS NULL OR c.mobile_number = $4)
        AND ($5::text IS NULL OR p.status = $5)
        AND ($6::text IS NULL OR p.office_code = $6)
)
SELECT
    (SELECT COUNT(*) FROM filtered) AS total_count,
    f.*
FROM filtered f
ORDER BY f.created_at DESC
LIMIT $7 OFFSET $8;
```

---

#### **2. GetProfileWithRelatedEntities Method**
```go
// repo/postgres/agent_profile.go
func (r *AgentProfileRepository) GetProfileWithRelatedEntities(
    ctx context.Context,
    agentID string,
) (*domain.AgentProfileComplete, error)
```

**Features**:
- Single query fetches profile + addresses + contacts + emails + office
- Uses JSON aggregation (json_agg) for related entities
- No N+1 query problem

**SQL Pattern**:
```sql
SELECT
    p.*,
    json_agg(DISTINCT a.*) FILTER (WHERE a.address_id IS NOT NULL) AS addresses,
    json_agg(DISTINCT c.*) FILTER (WHERE c.contact_id IS NOT NULL) AS contacts,
    json_agg(DISTINCT e.*) FILTER (WHERE e.email_id IS NOT NULL) AS emails,
    row_to_json(o.*) AS office
FROM agent_profiles p
LEFT JOIN agent_addresses a ON p.agent_id = a.agent_id
LEFT JOIN agent_contacts c ON p.agent_id = c.agent_id
LEFT JOIN agent_emails e ON p.agent_id = e.agent_id
LEFT JOIN offices o ON p.office_code = o.office_code
WHERE p.agent_id = $1
GROUP BY p.agent_id, o.office_code;
```

---

#### **3. UpdateSection Method**
```go
// repo/postgres/agent_profile.go
func (r *AgentProfileRepository) UpdateSectionReturning(
    ctx context.Context,
    agentID string,
    section string,
    updates map[string]interface{},
    updatedBy string,
) (*domain.AgentProfile, []domain.AgentAuditLog, error)
```

**Features**:
- Dynamic field updates based on section
- Automatic audit log creation (CTE pattern)
- Single database round trip
- Returns updated profile + audit logs created

**SQL Pattern**:
```sql
WITH updated AS (
    UPDATE agent_profiles
    SET
        first_name = COALESCE($2, first_name),
        last_name = COALESCE($3, last_name),
        updated_at = NOW(),
        updated_by = $4,
        version = version + 1
    WHERE agent_id = $1
    RETURNING *
),
audit_inserted AS (
    INSERT INTO agent_audit_logs (
        agent_id, action_type, field_name, old_value, new_value,
        performed_by, performed_at
    )
    SELECT
        agent_id,
        'PROFILE_UPDATE',
        unnest($5::text[]),  -- field names
        unnest($6::text[]),  -- old values
        unnest($7::text[]),  -- new values
        $4,
        NOW()
    FROM updated
    RETURNING *
)
SELECT
    (SELECT row_to_json(u.*) FROM updated u) AS profile,
    (SELECT json_agg(a.*) FROM audit_inserted a) AS audit_logs;
```

---

#### **4. GetAuditHistory Method**
```go
// repo/postgres/agent_audit_log.go
func (r *AgentAuditLogRepository) GetHistory(
    ctx context.Context,
    agentID string,
    fromDate, toDate *time.Time,
    page, limit int,
) ([]domain.AgentAuditLog, *PaginationMetadata, error)
```

**Features**:
- Date range filter
- Pagination
- Total count in same query

---

### **Phase 6B: Handler Layer**

#### **1. AGT-022: SearchAgents Handler**
```go
// handler/profile_update.go
func (h *ProfileUpdateHandler) SearchAgents(
    sctx *serverRoute.Context,
    req SearchAgentsRequest,
) (*Response, error) {
    // Validate search criteria
    // Call repository Search method
    // Return paginated results
}
```

**Request DTO**:
```go
type SearchAgentsRequest struct {
    AgentID     *string `query:"agent_id"`
    Name        *string `query:"name"`
    PANNumber   *string `query:"pan_number"`
    Mobile      *string `query:"mobile_number"`
    Email       *string `query:"email"`
    Status      *string `query:"status"`
    OfficeCode  *string `query:"office_code"`
    Page        int     `query:"page" default:"1"`
    Limit       int     `query:"limit" default:"20"`
}
```

**Response DTO**:
```go
type SearchAgentsResponse struct {
    Results    []AgentSearchResult   `json:"results"`
    Pagination PaginationMetadata    `json:"pagination"`
}

type AgentSearchResult struct {
    AgentID     string `json:"agent_id"`
    Name        string `json:"name"`
    AgentType   string `json:"agent_type"`
    PAN         string `json:"pan"`
    Mobile      string `json:"mobile"`
    Email       string `json:"email"`
    Status      string `json:"status"`
    Coordinator string `json:"advisor_coordinator"`
    Office      string `json:"office"`
}
```

---

#### **2. AGT-023: GetAgentProfile Handler**
```go
func (h *ProfileUpdateHandler) GetAgentProfile(
    sctx *serverRoute.Context,
    agentID string,
) (*Response, error) {
    // Call GetProfileWithRelatedEntities (single query)
    // Transform to response DTO
    // Return complete profile
}
```

**Response DTO**:
```go
type AgentProfileResponse struct {
    AgentProfile  AgentProfileDTO      `json:"agent_profile"`
    WorkflowState WorkflowStateDTO     `json:"workflow_state,omitempty"`
}

type AgentProfileDTO struct {
    AgentID      string               `json:"agent_id"`
    ProfileType  string               `json:"profile_type"`
    FullName     string               `json:"full_name"`
    PANNumber    string               `json:"pan_number"`
    Status       string               `json:"status"`
    PersonalInfo PersonalInfoDTO      `json:"personal_info"`
    Addresses    []AddressDTO         `json:"addresses"`
    Contacts     []ContactDTO         `json:"contacts"`
    Emails       []EmailDTO           `json:"emails"`
    Office       OfficeDTO            `json:"office"`
}
```

---

#### **3. AGT-024: GetUpdateForm Handler**
```go
func (h *ProfileUpdateHandler) GetUpdateForm(
    sctx *serverRoute.Context,
    agentID string,
) (*Response, error) {
    // Get current profile data
    // Return pre-populated form data
    // Include editable fields metadata
}
```

**Response DTO**:
```go
type UpdateFormResponse struct {
    AgentID       string                    `json:"agent_id"`
    Sections      []SectionDTO              `json:"sections"`
    CurrentData   map[string]interface{}    `json:"current_data"`
    EditableFields map[string]FieldMetadata `json:"editable_fields"`
}

type SectionDTO struct {
    Name        string   `json:"name"`        // "personal_info", "address", "contact"
    DisplayName string   `json:"display_name"`
    Fields      []string `json:"fields"`
}

type FieldMetadata struct {
    Name             string   `json:"name"`
    DisplayName      string   `json:"display_name"`
    Type             string   `json:"type"`
    Required         bool     `json:"required"`
    Editable         bool     `json:"editable"`
    RequiresApproval bool     `json:"requires_approval"`
    ValidationRules  []string `json:"validation_rules"`
}
```

---

#### **4. AGT-025: UpdateProfileSection Handler**
```go
func (h *ProfileUpdateHandler) UpdateProfileSection(
    sctx *serverRoute.Context,
    agentID string,
    section string,
    req UpdateSectionRequest,
) (*Response, error) {
    // Validate section name
    // Check if fields require approval

    // If critical fields (name, PAN, Aadhar):
    //   - Create approval request
    //   - Start ProfileUpdateWorkflow
    //   - Return "pending approval" status

    // If non-critical fields:
    //   - Update directly using UpdateSectionReturning
    //   - Return updated profile
}
```

**Request DTO**:
```go
type UpdateSectionRequest struct {
    Section   string                 `json:"section"`   // "personal_info", "address", "contact"
    Updates   map[string]interface{} `json:"updates"`
    UpdatedBy string                 `json:"updated_by"`
    Reason    string                 `json:"reason,omitempty"` // Required for critical fields
}
```

**Response DTO**:
```go
type UpdateSectionResponse struct {
    AgentID          string                 `json:"agent_id"`
    Section          string                 `json:"section"`
    Status           string                 `json:"status"` // "UPDATED", "PENDING_APPROVAL"
    UpdatedFields    []string               `json:"updated_fields"`
    ApprovalRequired bool                   `json:"approval_required"`
    ApprovalRequestID *string               `json:"approval_request_id,omitempty"`
    UpdatedProfile   *AgentProfileDTO       `json:"updated_profile,omitempty"`
    ChangedFields    map[string]ChangeInfo  `json:"changed_fields"`
}

type ChangeInfo struct {
    OldValue string `json:"old_value"`
    NewValue string `json:"new_value"`
}
```

**Critical Fields Logic**:
```go
criticalFields := map[string]bool{
    "first_name":    true,
    "middle_name":   true,
    "last_name":     true,
    "pan_number":    true,
    "aadhar_number": true,
}

func requiresApproval(updates map[string]interface{}) bool {
    for field := range updates {
        if criticalFields[field] {
            return true
        }
    }
    return false
}
```

---

#### **5. AGT-026: ApproveProfileUpdate Handler**
```go
func (h *ProfileUpdateHandler) ApproveProfileUpdate(
    sctx *serverRoute.Context,
    approvalRequestID string,
    req ApprovalRequest,
) (*Response, error) {
    // Fetch approval request
    // Validate status is PENDING
    // Send signal to ProfileUpdateWorkflow
    // Return approval confirmation
}
```

**Request DTO**:
```go
type ApprovalRequest struct {
    Action    string `json:"action"` // "APPROVE" or "REJECT"
    Comments  string `json:"comments"`
    ApprovedBy string `json:"approved_by"`
}
```

**Response DTO**:
```go
type ApprovalResponse struct {
    ApprovalRequestID string    `json:"approval_request_id"`
    Status            string    `json:"status"`
    AgentID           string    `json:"agent_id"`
    ApprovedBy        string    `json:"approved_by"`
    ApprovedAt        time.Time `json:"approved_at"`
    Message           string    `json:"message"`
}
```

---

#### **6. AGT-027: RejectProfileUpdate Handler**
```go
func (h *ProfileUpdateHandler) RejectProfileUpdate(
    sctx *serverRoute.Context,
    approvalRequestID string,
    req RejectionRequest,
) (*Response, error) {
    // Fetch approval request
    // Validate status is PENDING
    // Send REJECT signal to ProfileUpdateWorkflow
    // Return rejection confirmation
}
```

**Request DTO**:
```go
type RejectionRequest struct {
    Action   string `json:"action"` // "REJECT"
    Reason   string `json:"reason" validate:"required,min=10"`
    RejectedBy string `json:"rejected_by"`
}
```

---

#### **7. AGT-028: GetAuditHistory Handler**
```go
func (h *ProfileUpdateHandler) GetAuditHistory(
    sctx *serverRoute.Context,
    agentID string,
    req AuditHistoryRequest,
) (*Response, error) {
    // Call repository GetHistory method
    // Return paginated audit logs
}
```

**Request DTO**:
```go
type AuditHistoryRequest struct {
    FromDate *time.Time `query:"from_date"`
    ToDate   *time.Time `query:"to_date"`
    Page     int        `query:"page" default:"1"`
    Limit    int        `query:"limit" default:"50"`
}
```

**Response DTO**:
```go
type AuditHistoryResponse struct {
    AgentID    string                `json:"agent_id"`
    AuditLogs  []AuditLogDTO         `json:"audit_logs"`
    Pagination PaginationMetadata    `json:"pagination"`
}

type AuditLogDTO struct {
    AuditID     string    `json:"audit_id"`
    Action      string    `json:"action"`
    FieldName   string    `json:"field_name"`
    OldValue    string    `json:"old_value"`
    NewValue    string    `json:"new_value"`
    PerformedBy string    `json:"performed_by"`
    PerformedAt time.Time `json:"performed_at"`
    IPAddress   string    `json:"ip_address,omitempty"`
}
```

---

### **Phase 6C: Workflow (Optional - For Critical Fields)**

#### **WF-AGT-PRF-003: ProfileUpdateWorkflow**

**Purpose**: Handle approval for critical field updates (name, PAN, Aadhar)

**Pattern**: Human-in-the-loop with child workflow (reuse ApprovalWorkflow from Phase 5)

**Activities**:
1. **RecordUpdateRequestActivity** - Create approval request record
2. **StartApprovalChildWorkflow** - Wait for approval signal
3. **ApplyUpdateActivity** - Apply approved changes (if approved)
4. **NotifyUserActivity** - Send notification

**Workflow Code**:
```go
// workflows/profile_update_workflow.go
func ProfileUpdateWorkflow(ctx workflow.Context, input UpdateInput) (*UpdateOutput, error) {
    // Step 0: Record workflow start
    // Step 1: Create approval request
    // Step 2: Start approval child workflow (wait for signal)
    // Step 3: If approved, apply update using UpdateSectionReturning
    // Step 4: Send notification
    // Return result
}
```

**Note**: This workflow is OPTIONAL for Phase 6. Can be implemented later if needed. For now, we can:
1. Implement non-critical field updates directly
2. Return "approval required" message for critical fields
3. Implement full workflow in Phase 6.1 if requested

---

## 📦 IMPLEMENTATION CHECKLIST

### **Step 1: Repository Methods** ✅
- [ ] Add `Search()` method to AgentProfileRepository
- [ ] Add `GetProfileWithRelatedEntities()` method
- [ ] Add `UpdateSectionReturning()` method
- [ ] Add `GetHistory()` method to AgentAuditLogRepository
- [ ] Test all methods compile

### **Step 2: Domain Models** ✅
- [ ] Create `AgentProfileComplete` domain model (if needed)
- [ ] Update `AgentAuditLog` domain model if needed
- [ ] Add `PaginationMetadata` struct

### **Step 3: Request DTOs** ✅
- [ ] Create `SearchAgentsRequest`
- [ ] Create `UpdateSectionRequest`
- [ ] Create `ApprovalRequest`
- [ ] Create `RejectionRequest`
- [ ] Create `AuditHistoryRequest`
- [ ] Add validation tags

### **Step 4: Response DTOs** ✅
- [ ] Create `SearchAgentsResponse` with `AgentSearchResult`
- [ ] Create `AgentProfileResponse` with all nested DTOs
- [ ] Create `UpdateFormResponse` with metadata
- [ ] Create `UpdateSectionResponse` with `ChangeInfo`
- [ ] Create `ApprovalResponse`
- [ ] Create `AuditHistoryResponse` with `AuditLogDTO`

### **Step 5: Handlers** ✅
- [ ] Create `handler/profile_update.go`
- [ ] Implement AGT-022: SearchAgents
- [ ] Implement AGT-023: GetAgentProfile
- [ ] Implement AGT-024: GetUpdateForm
- [ ] Implement AGT-025: UpdateProfileSection
- [ ] Implement AGT-026: ApproveProfileUpdate
- [ ] Implement AGT-027: RejectProfileUpdate
- [ ] Implement AGT-028: GetAuditHistory

### **Step 6: Bootstrap** ✅
- [ ] Create ProfileUpdateHandler constructor
- [ ] Register handler in FxHandler
- [ ] Register routes

### **Step 7: Testing & Verification** ✅
- [ ] Format code (`gofmt -w .`)
- [ ] Verify compilation (`go build`)
- [ ] Check for TODOs
- [ ] Commit with descriptive message

---

## 🎯 SUCCESS CRITERIA

- ✅ All 7 endpoints implemented
- ✅ Single database round trip per operation (where possible)
- ✅ Atomic operations with CTE patterns
- ✅ Automatic audit logging for updates
- ✅ Pagination support for search and audit history
- ✅ Zero compilation errors
- ✅ All critical patterns applied

---

## 📊 ESTIMATED EFFORT

| Task | Lines of Code | Complexity |
|------|--------------|------------|
| Repository Methods | ~300 lines | Medium (CTE + JSON aggregation) |
| Domain Models | ~50 lines | Low |
| Request DTOs | ~150 lines | Low |
| Response DTOs | ~300 lines | Low |
| Handlers | ~600 lines | Medium (business logic) |
| Bootstrap | ~20 lines | Low |
| **TOTAL** | **~1,420 lines** | **Medium** |

---

## 🚀 READY TO START?

Phase 6 is well-scoped and builds on all the patterns from Phase 3-5:
- ✅ Repository layer is mostly complete (just need 4 new methods)
- ✅ Temporal workflow infrastructure already exists (reuse ApprovalWorkflow)
- ✅ All critical patterns documented and proven

**Let's start with Step 1: Repository Methods!**

---

**END OF PLAN**
