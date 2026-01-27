# Phase 8-10 Context - Status Management, Search & Dashboard, Batch & Webhook APIs

**Date**: 2026-01-27
**Branch**: claude/develop-policy-apis-golang-BcDD3
**Status**: Planning document for Phases 8, 9, 10
**Purpose**: Track implementation across different sessions

---

## 📋 TABLE OF CONTENTS

1. [Phase 8: Status Management APIs](#phase-8-status-management-apis)
2. [Phase 9: Search & Dashboard APIs](#phase-9-search--dashboard-apis)
3. [Phase 10: Batch & Webhook APIs](#phase-10-batch--webhook-apis)
4. [Database Requirements](#database-requirements)
5. [Repository Methods Needed](#repository-methods-needed)
6. [Temporal Workflows Required](#temporal-workflows-required)
7. [Integration Points](#integration-points)
8. [Implementation Checklist](#implementation-checklist)

---

## 🔴 PHASE 8: STATUS MANAGEMENT APIs

**Scope**: 10 endpoints covering termination, reinstatement, and status lookups

### **Business Rules**

| Rule ID | Description | Enforcement |
|---------|-------------|-------------|
| **BR-AGT-PRF-016** | Status Update with Reason | Termination/reinstatement requires reason (min 20 chars for termination, min 10 for reinstatement) |
| **BR-AGT-PRF-017** | Agent Termination Workflow | Termination triggers: status update, portal disable, commission stop, letter generation, data archival (7 years) |

### **Functional Requirements**

| FR ID | Description | APIs |
|-------|-------------|------|
| **FR-AGT-PRF-013** | Status Update | AGT-039, AGT-041, AGT-060 to AGT-063 |
| **FR-AGT-PRF-014** | Termination Workflow | AGT-039, AGT-040 |

### **Validation Rules**

| VR ID | Description | Validation |
|-------|-------------|------------|
| **VR-AGT-PRF-020** | Termination Reason Required | `minLength: 20` |

### **Workflows**

| WF ID | Workflow Name | Purpose | APIs |
|-------|---------------|---------|------|
| **WF-AGT-PRF-004** | Termination Workflow | Orchestrates termination process with multiple activities | AGT-039, AGT-040 |
| **WF-AGT-PRF-011** | Reinstatement Workflow | Human-in-the-loop approval for reinstatement | AGT-041, AGT-060 to AGT-063 |

---

### **API Details**

#### **AGT-039: POST /agents/{agent_id}/terminate**

**Purpose**: Terminate an agent with reason and auto-generate termination letter

**Request**:
```json
{
  "termination_reason": "string (min 20 chars)",
  "termination_reason_code": "RESIGNATION|MISCONDUCT|NON_PERFORMANCE|FRAUD|OTHER",
  "effective_date": "DD-MM-YYYY",
  "terminated_by": "string",
  "generate_termination_letter": true,
  "notified": true
}
```

**Response**:
```json
{
  "agent_id": "AGT-2026-000567",
  "termination_status": "TERMINATED",
  "termination_details": {
    "reason": "string",
    "effective_date": "DD-MM-YYYY",
    "terminated_by": "string"
  },
  "actions_performed": [
    "STATUS_UPDATED_TO_TERMINATED",
    "PORTAL_ACCESS_DISABLED",
    "COMMISSION_PROCESSING_STOPPED",
    "TERMINATION_LETTER_GENERATED",
    "DATA_ARCHIVED_7_YEARS",
    "NOTIFICATIONS_SENT"
  ],
  "termination_letter_url": "https://...",
  "notifications_sent": [
    {
      "type": "EMAIL",
      "recipient": "agent@example.com",
      "status": "SENT"
    }
  ]
}
```

**Business Logic**:
1. Validate termination reason (min 20 chars)
2. Check agent exists and is not already TERMINATED
3. Start WF-AGT-PRF-004 (Termination Workflow)
4. Workflow activities:
   - ACT-TRM-001: Update agent status to TERMINATED
   - ACT-TRM-002: Disable portal access (update agent_authentication table)
   - ACT-TRM-003: Stop commission processing (flag in agent_profiles)
   - ACT-TRM-004: Generate termination letter (call INT-AGT-006)
   - ACT-TRM-005: Archive agent data (create archive record with 7-year retention)
   - ACT-TRM-006: Send notifications (call INT-AGT-005)
   - ACT-TRM-007: Create audit log
5. Return termination details with letter URL

**Database Updates**:
- `agent_profiles.status = 'TERMINATED'`
- `agent_profiles.termination_date = effective_date`
- `agent_profiles.termination_reason = termination_reason`
- `agent_profiles.termination_reason_code = termination_reason_code`
- `agent_profiles.commission_enabled = false`
- `agent_authentication.status = 'DISABLED'`
- Insert into `agent_termination_records` table
- Insert into `agent_data_archives` table with retention_until = NOW() + 7 years
- Insert into `agent_audit_logs`

**Error Handling**:
- ERR-AGT-PRF-010: Termination reason required (min 20 chars)
- E-AGT-PRF-008: Agent not found

---

#### **AGT-040: GET /agents/{agent_id}/termination-letter**

**Purpose**: Download termination letter (PDF or HTML)

**Request**:
- Path: `agent_id`
- Query: `format=PDF|HTML` (default: PDF)

**Response**: Binary file (application/pdf or text/html)

**Business Logic**:
1. Verify agent exists and is TERMINATED
2. Fetch termination record from `agent_termination_records`
3. Call INT-AGT-006 (Letter Generation Service) with template
4. Return generated letter as binary file

**Integration**: INT-AGT-006 (Letter Generation Service)

---

#### **AGT-041: POST /agents/{agent_id}/reinstate**

**Purpose**: Initiate agent reinstatement (requires approval)

**Request**:
```json
{
  "reinstatement_reason": "string (min 10 chars)",
  "effective_date": "DD-MM-YYYY",
  "reinstated_by": "string",
  "supporting_documents": ["doc1_url", "doc2_url"]
}
```

**Response**:
```json
{
  "agent_id": "AGT-2026-000567",
  "reinstatement_status": "PENDING_APPROVAL",
  "request_id": "uuid"
}
```

**Business Logic**:
1. Validate reinstatement reason (min 10 chars)
2. Check agent exists and is TERMINATED or SUSPENDED
3. Start WF-AGT-PRF-011 (Reinstatement Workflow)
4. Workflow:
   - Create reinstatement request record
   - Trigger approval child workflow (wait for AGT-061 or AGT-062)
   - If approved: Update status to ACTIVE
   - If rejected: Keep current status
5. Return request_id for tracking

**Database Updates**:
- Insert into `agent_reinstatement_requests` table with status='PENDING_APPROVAL'

---

#### **AGT-060: POST /reinstatement/request**

**Purpose**: Create reinstatement request (alternative endpoint for AGT-041)

**Request**:
```json
{
  "agent_id": "AGT-2026-000567",
  "reinstatement_reason": "string (min 10 chars)",
  "reinstatement_reason_code": "string",
  "effective_date": "DD-MM-YYYY",
  "supporting_documents": ["doc1_url", "doc2_url"],
  "requested_by": "string"
}
```

**Response**: Same as AGT-041

---

#### **AGT-061: PUT /reinstatement/{request_id}/approve**

**Purpose**: Approve reinstatement request (signals workflow)

**Request**:
```json
{
  "action": "APPROVE",
  "approver_comments": "string",
  "approved_by": "string"
}
```

**Response**:
```json
{
  "request_id": "uuid",
  "status": "APPROVED",
  "agent_id": "AGT-2026-000567",
  "approved_at": "2026-01-27T10:30:00Z"
}
```

**Business Logic**:
1. Fetch reinstatement request by request_id
2. Validate request status is PENDING_APPROVAL
3. Send signal to WF-AGT-PRF-011 workflow:
   - Signal name: `reinstatement-decision`
   - Signal data: `{decision: "APPROVED", approved_by: "...", comments: "..."}`
4. Workflow resumes and executes approval activities:
   - ACT-RNS-001: Update agent status to ACTIVE
   - ACT-RNS-002: Enable portal access
   - ACT-RNS-003: Resume commission processing
   - ACT-RNS-004: Send notifications
   - ACT-RNS-005: Create audit log
5. Update reinstatement request status to APPROVED

**Database Updates**:
- `agent_reinstatement_requests.status = 'APPROVED'`
- `agent_reinstatement_requests.approved_by = approved_by`
- `agent_reinstatement_requests.approved_at = NOW()`
- `agent_profiles.status = 'ACTIVE'`
- `agent_profiles.reinstatement_date = effective_date`
- `agent_authentication.status = 'ACTIVE'`

---

#### **AGT-062: PUT /reinstatement/{request_id}/reject**

**Purpose**: Reject reinstatement request (signals workflow)

**Request**:
```json
{
  "action": "REJECT",
  "rejector_comments": "string (min 10 chars)",
  "rejected_by": "string"
}
```

**Business Logic**: Similar to AGT-061 but sends REJECT signal to workflow

**Database Updates**:
- `agent_reinstatement_requests.status = 'REJECTED'`
- `agent_reinstatement_requests.rejected_by = rejected_by`
- `agent_reinstatement_requests.rejection_reason = rejector_comments`

---

#### **AGT-063: POST /reinstatement/{request_id}/documents**

**Purpose**: Upload supporting documents for reinstatement request

**Request**: multipart/form-data
- `documents`: array of files
- `document_type`: IDENTITY_PROOF | ADDRESS_PROOF | LICENCE_COPY | NO_DUE_CERTIFICATE

**Response**:
```json
{
  "request_id": "uuid",
  "documents_uploaded": [
    {
      "document_id": "uuid",
      "document_type": "IDENTITY_PROOF",
      "file_name": "pan_card.pdf",
      "upload_status": "SUCCESS"
    }
  ]
}
```

**Business Logic**:
1. Validate file types (PDF, JPG, PNG)
2. Validate file size (max 5MB per file)
3. Upload to document storage service
4. Store document metadata in database
5. Link documents to reinstatement request

**Database Updates**:
- Insert into `agent_documents` table
- Link via `agent_reinstatement_documents` join table

---

#### **AGT-070: GET /status-types** (Lookup)

**Purpose**: Get list of agent status types

**Response**:
```json
{
  "status_types": [
    {"code": "ACTIVE", "name": "Active"},
    {"code": "SUSPENDED", "name": "Suspended"},
    {"code": "TERMINATED", "name": "Terminated"},
    {"code": "DEACTIVATED", "name": "Deactivated"}
  ]
}
```

**Implementation**: Static lookup from `status_types` table or constants

---

#### **AGT-071: GET /reinstatement/reasons** (Lookup)

**Purpose**: Get list of reinstatement reasons

**Response**:
```json
{
  "reinstatement_reasons": [
    {"code": "WRONGFUL_TERMINATION", "name": "Wrongful Termination"},
    {"code": "APPEAL_APPROVED", "name": "Appeal Approved"},
    {"code": "MUTUAL_AGREEMENT", "name": "Mutual Agreement"}
  ]
}
```

---

#### **AGT-072: GET /termination/reasons** (Lookup)

**Purpose**: Get list of termination reasons

**Response**:
```json
{
  "termination_reasons": [
    {"code": "RESIGNATION", "name": "Resignation", "description": "Voluntary resignation"},
    {"code": "MISCONDUCT", "name": "Misconduct", "description": "Code of conduct violation"},
    {"code": "NON_PERFORMANCE", "name": "Non Performance", "description": "Failure to meet targets"},
    {"code": "FRAUD", "name": "Fraud", "description": "Fraudulent activities"},
    {"code": "OTHER", "name": "Other", "description": "Other reasons"}
  ]
}
```

---

### **Phase 8 Repository Methods Needed**

```go
// repo/postgres/agent_profile.go
func (r *AgentProfileRepository) TerminateAgentReturning(
    ctx context.Context,
    agentID string,
    reason string,
    reasonCode string,
    effectiveDate time.Time,
    terminatedBy string,
) (*domain.AgentProfile, error)

func (r *AgentProfileRepository) ReinstateAgentReturning(
    ctx context.Context,
    agentID string,
    effectiveDate time.Time,
) (*domain.AgentProfile, error)

// repo/postgres/agent_reinstatement_request.go (NEW)
func (r *ReinstatementRequestRepository) Create(
    ctx context.Context,
    request *domain.ReinstatementRequest,
) (*domain.ReinstatementRequest, error)

func (r *ReinstatementRequestRepository) ApproveReturning(
    ctx context.Context,
    requestID string,
    approvedBy string,
    comments string,
) (*domain.ReinstatementRequest, error)

func (r *ReinstatementRequestRepository) RejectReturning(
    ctx context.Context,
    requestID string,
    rejectedBy string,
    comments string,
) (*domain.ReinstatementRequest, error)
```

---

### **Phase 8 Domain Models Needed**

```go
// core/domain/agent_termination_record.go (NEW)
type AgentTerminationRecord struct {
    TerminationID      string    `db:"termination_id"`
    AgentID            string    `db:"agent_id"`
    TerminationReason  string    `db:"termination_reason"`
    ReasonCode         string    `db:"reason_code"`
    EffectiveDate      time.Time `db:"effective_date"`
    TerminatedBy       string    `db:"terminated_by"`
    LetterGeneratedURL string    `db:"letter_generated_url"`
    CreatedAt          time.Time `db:"created_at"`
}

// core/domain/agent_reinstatement_request.go (NEW)
type AgentReinstatementRequest struct {
    RequestID          string    `db:"request_id"`
    AgentID            string    `db:"agent_id"`
    ReinstatementReason string   `db:"reinstatement_reason"`
    ReasonCode         string    `db:"reason_code"`
    EffectiveDate      time.Time `db:"effective_date"`
    RequestedBy        string    `db:"requested_by"`
    Status             string    `db:"status"` // PENDING_APPROVAL, APPROVED, REJECTED
    ApprovedBy         *string   `db:"approved_by"`
    ApprovedAt         *time.Time `db:"approved_at"`
    RejectedBy         *string   `db:"rejected_by"`
    RejectionReason    *string   `db:"rejection_reason"`
    CreatedAt          time.Time `db:"created_at"`
}

// core/domain/agent_data_archive.go (NEW)
type AgentDataArchive struct {
    ArchiveID     string    `db:"archive_id"`
    AgentID       string    `db:"agent_id"`
    ArchivedData  string    `db:"archived_data"` // JSON blob of complete agent data
    ArchivedAt    time.Time `db:"archived_at"`
    RetentionUntil time.Time `db:"retention_until"` // archived_at + 7 years
    ArchivedBy    string    `db:"archived_by"`
}
```

---

### **Phase 8 Database Tables Needed**

```sql
-- Migration: 003_agent_status_management.sql

CREATE TABLE agent_termination_records (
    termination_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agent_profiles(agent_id),
    termination_reason TEXT NOT NULL CHECK (LENGTH(termination_reason) >= 20),
    reason_code VARCHAR(50) NOT NULL,
    effective_date DATE NOT NULL,
    terminated_by VARCHAR(255) NOT NULL,
    letter_generated_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT fk_termination_agent FOREIGN KEY (agent_id) REFERENCES agent_profiles(agent_id)
);

CREATE TABLE agent_reinstatement_requests (
    request_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agent_profiles(agent_id),
    reinstatement_reason TEXT NOT NULL CHECK (LENGTH(reinstatement_reason) >= 10),
    reason_code VARCHAR(50),
    effective_date DATE NOT NULL,
    requested_by VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING_APPROVAL',
    approved_by VARCHAR(255),
    approved_at TIMESTAMP WITH TIME ZONE,
    rejected_by VARCHAR(255),
    rejection_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT fk_reinstatement_agent FOREIGN KEY (agent_id) REFERENCES agent_profiles(agent_id),
    CONSTRAINT chk_status CHECK (status IN ('PENDING_APPROVAL', 'APPROVED', 'REJECTED'))
);

CREATE TABLE agent_data_archives (
    archive_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL,
    archived_data JSONB NOT NULL,
    archived_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    retention_until TIMESTAMP WITH TIME ZONE NOT NULL,
    archived_by VARCHAR(255) NOT NULL,
    CONSTRAINT chk_retention CHECK (retention_until > archived_at)
);

CREATE TABLE agent_documents (
    document_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agent_profiles(agent_id),
    document_type VARCHAR(50) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_url TEXT NOT NULL,
    file_size_bytes INTEGER NOT NULL,
    uploaded_by VARCHAR(255) NOT NULL,
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT fk_document_agent FOREIGN KEY (agent_id) REFERENCES agent_profiles(agent_id)
);

CREATE TABLE agent_reinstatement_documents (
    request_id UUID NOT NULL REFERENCES agent_reinstatement_requests(request_id),
    document_id UUID NOT NULL REFERENCES agent_documents(document_id),
    PRIMARY KEY (request_id, document_id)
);

-- Indexes
CREATE INDEX idx_termination_agent_id ON agent_termination_records(agent_id);
CREATE INDEX idx_termination_effective_date ON agent_termination_records(effective_date);
CREATE INDEX idx_reinstatement_agent_id ON agent_reinstatement_requests(agent_id);
CREATE INDEX idx_reinstatement_status ON agent_reinstatement_requests(status);
CREATE INDEX idx_archive_agent_id ON agent_data_archives(agent_id);
CREATE INDEX idx_archive_retention ON agent_data_archives(retention_until);
CREATE INDEX idx_documents_agent_id ON agent_documents(agent_id);
```

---

## 🔵 PHASE 9: SEARCH & DASHBOARD APIs

**Scope**: 7 endpoints covering search, profile view, audit, dashboard, hierarchy, timeline, notifications

### **Business Rules**

| Rule ID | Description | Enforcement |
|---------|-------------|-------------|
| **BR-AGT-PRF-022** | Multi-Criteria Search | Support search by agent_id, name, PAN, mobile, status, office |
| **BR-AGT-PRF-023** | Dashboard Profile View | Single query fetches profile with all related entities |

### **Functional Requirements**

| FR ID | Description | APIs |
|-------|-------------|------|
| **FR-AGT-PRF-004** | Agent Search | AGT-022 |
| **FR-AGT-PRF-005** | Profile Dashboard View | AGT-023 |
| **FR-AGT-PRF-018** | Agent Dashboard | AGT-068 |
| **FR-AGT-PRF-021** | Self-Service Update | AGT-068 |
| **FR-AGT-PRF-022** | Audit History | AGT-028, AGT-076 |

---

### **API Details**

#### **AGT-022: GET /agents/search**

**Purpose**: Multi-criteria agent search with pagination

**Request** (Query Parameters):
```
agent_id=AGT-2026-000567
name=John
pan_number=ABCDE1234F
mobile_number=9876543210
status=ACTIVE|SUSPENDED|TERMINATED|DEACTIVATED
office_code=OFF-001
page=1
limit=20
```

**Response**:
```json
{
  "results": [
    {
      "agent_id": "AGT-2026-000567",
      "name": "John Doe",
      "agent_type": "ADVISOR",
      "pan": "ABCDE1234F",
      "mobile": "9876543210",
      "email": "john@example.com",
      "status": "ACTIVE",
      "advisor_coordinator": "Jane Smith",
      "office": "Mumbai Branch"
    }
  ],
  "pagination": {
    "current_page": 1,
    "total_pages": 10,
    "total_results": 200,
    "results_per_page": 20
  }
}
```

**Business Logic**:
1. Build dynamic WHERE clause based on provided filters
2. Apply ILIKE for partial name search
3. Use pagination (LIMIT/OFFSET)
4. Join with related tables (coordinator, office) for single query
5. Return results with pagination metadata

**Performance Optimization**:
- Use single query with LEFT JOINs instead of multiple queries
- Create composite indexes on frequently searched columns
- Apply filters efficiently using Squirrel query builder

**SQL Pattern**:
```sql
SELECT
    p.agent_id,
    CONCAT(p.first_name, ' ', p.last_name) AS name,
    p.agent_type,
    p.pan_number,
    c.mobile_number,
    e.email_address,
    p.status,
    CONCAT(coord.first_name, ' ', coord.last_name) AS advisor_coordinator,
    o.office_name
FROM agent_profiles p
LEFT JOIN agent_contacts c ON p.agent_id = c.agent_id AND c.contact_type = 'PRIMARY'
LEFT JOIN agent_emails e ON p.agent_id = e.agent_id AND e.email_type = 'PRIMARY'
LEFT JOIN agent_profiles coord ON p.advisor_coordinator_id = coord.agent_id
LEFT JOIN offices o ON p.office_code = o.office_code
WHERE
    ($1::text IS NULL OR p.agent_id = $1)
    AND ($2::text IS NULL OR CONCAT(p.first_name, ' ', p.last_name) ILIKE '%' || $2 || '%')
    AND ($3::text IS NULL OR p.pan_number = $3)
    AND ($4::text IS NULL OR c.mobile_number = $4)
    AND ($5::text IS NULL OR p.status = $5)
    AND ($6::text IS NULL OR p.office_code = $6)
ORDER BY p.created_at DESC
LIMIT $7 OFFSET $8;
```

---

#### **AGT-023: GET /agents/{agent_id}**

**Purpose**: Get complete agent profile with all related entities

**Response**:
```json
{
  "agent_profile": {
    "agent_id": "AGT-2026-000567",
    "profile_type": "ADVISOR",
    "full_name": "John Doe",
    "pan_number": "ABCDE1234F",
    "status": "ACTIVE",
    "personal_info": {
      "first_name": "John",
      "middle_name": "M",
      "last_name": "Doe",
      "date_of_birth": "01-01-1990",
      "gender": "MALE",
      "aadhar_number": "123456789012",
      "marital_status": "MARRIED"
    },
    "addresses": [
      {
        "address_id": "uuid",
        "address_type": "PERMANENT",
        "line1": "123 Main St",
        "line2": "Apt 4B",
        "city": "Mumbai",
        "state": "Maharashtra",
        "pincode": "400001",
        "country": "India"
      }
    ],
    "contacts": [
      {
        "contact_id": "uuid",
        "contact_type": "PRIMARY",
        "mobile_number": "9876543210",
        "alternate_number": "9876543211"
      }
    ],
    "emails": [
      {
        "email_id": "uuid",
        "email_type": "PRIMARY",
        "email_address": "john@example.com"
      }
    ],
    "office": {
      "office_code": "OFF-001",
      "office_name": "Mumbai Branch",
      "office_type": "BRANCH"
    }
  },
  "workflow_state": {
    "current_step": "ACTIVE",
    "progress_percentage": 100
  }
}
```

**Business Logic**:
1. Fetch profile with all related entities in single query using JOINs
2. Use JSON aggregation for arrays (addresses, contacts, emails)
3. Return complete profile object

**Performance Optimization** (CRITICAL):
```sql
-- Single query with JSON aggregation (1 round trip)
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

#### **AGT-028: GET /agents/{agent_id}/audit-history**

**Purpose**: Get audit history with pagination and date filters

**Request**:
```
/agents/AGT-2026-000567/audit-history?from_date=01-01-2026&to_date=31-01-2026&page=1&limit=50
```

**Response**:
```json
{
  "agent_id": "AGT-2026-000567",
  "audit_logs": [
    {
      "audit_id": "uuid",
      "action": "PROFILE_UPDATE",
      "field_name": "mobile_number",
      "old_value": "9876543210",
      "new_value": "9876543211",
      "performed_by": "admin@example.com",
      "performed_at": "2026-01-27T10:30:00Z",
      "ip_address": "192.168.1.1"
    }
  ],
  "pagination": {
    "current_page": 1,
    "total_pages": 5,
    "total_results": 250
  }
}
```

**Business Logic**:
1. Query `agent_audit_logs` table with filters
2. Apply date range filter if provided
3. Paginate results
4. Order by `performed_at DESC`

---

#### **AGT-068: GET /dashboard/agent/{agent_id}**

**Purpose**: Agent dashboard with performance metrics and tasks

**Response**:
```json
{
  "agent_id": "AGT-2026-000567",
  "profile_summary": {
    "name": "John Doe",
    "agent_type": "ADVISOR",
    "status": "ACTIVE",
    "pan": "ABCDE1234F"
  },
  "performance_metrics": {
    "policies_sold": 150,
    "premium_collected": 5000000.00,
    "targets_achieved": {
      "monthly_target": 100,
      "achieved": 150,
      "percentage": 150
    }
  },
  "pending_tasks": [
    {
      "task": "License renewal pending",
      "priority": "HIGH",
      "due_date": "15-02-2026",
      "overdue": false
    }
  ],
  "notifications": [
    {
      "message": "Your license will expire in 30 days",
      "type": "WARNING",
      "timestamp": "2026-01-27T10:30:00Z"
    }
  ]
}
```

**Business Logic**:
1. Fetch profile summary
2. Fetch performance metrics from separate tables (policies, premiums, targets)
3. Calculate pending tasks (license renewals, goal achievements)
4. Fetch recent notifications
5. Combine all data in single response

**Performance**: Use multiple concurrent queries or single query with UNION ALL

---

#### **AGT-073: GET /agents/{agent_id}/hierarchy**

**Purpose**: Get agent's hierarchy chain (advisor → coordinator → manager)

**Response**:
```json
{
  "agent_id": "AGT-2026-000567",
  "agent_type": "ADVISOR",
  "hierarchy_chain": [
    {
      "agent_id": "AGT-2026-000567",
      "name": "John Doe",
      "agent_type": "ADVISOR",
      "level": 1
    },
    {
      "agent_id": "AGT-2026-000100",
      "name": "Jane Smith",
      "agent_type": "ADVISOR_COORDINATOR",
      "level": 2
    }
  ]
}
```

**Business Logic**:
1. Start with current agent
2. Follow `advisor_coordinator_id` links recursively
3. Build hierarchy chain
4. Return ordered by level

**SQL Pattern** (Recursive CTE):
```sql
WITH RECURSIVE hierarchy AS (
    SELECT agent_id, first_name, last_name, agent_type, advisor_coordinator_id, 1 AS level
    FROM agent_profiles
    WHERE agent_id = $1
    UNION ALL
    SELECT p.agent_id, p.first_name, p.last_name, p.agent_type, p.advisor_coordinator_id, h.level + 1
    FROM agent_profiles p
    INNER JOIN hierarchy h ON p.agent_id = h.advisor_coordinator_id
)
SELECT * FROM hierarchy ORDER BY level;
```

---

#### **AGT-076: GET /agents/{agent_id}/timeline**

**Purpose**: Agent activity timeline with filters

**Request**:
```
/agents/AGT-2026-000567/timeline?from_date=01-01-2026&to_date=31-01-2026&activity_type=PROFILE_CHANGE&page=1&limit=50
```

**Response**:
```json
{
  "agent_id": "AGT-2026-000567",
  "timeline": [
    {
      "timestamp": "2026-01-27T10:30:00Z",
      "activity_type": "PROFILE_CHANGE",
      "description": "Mobile number updated from 9876543210 to 9876543211",
      "performed_by": "admin@example.com"
    },
    {
      "timestamp": "2026-01-26T15:00:00Z",
      "activity_type": "LICENSE_UPDATE",
      "description": "License renewed for type LIFE_INSURANCE",
      "performed_by": "john@example.com"
    }
  ],
  "pagination": {
    "current_page": 1,
    "total_pages": 10
  }
}
```

**Business Logic**:
1. Query audit logs + license changes + status changes
2. Combine into timeline format
3. Filter by activity_type if provided
4. Order by timestamp DESC
5. Paginate results

---

#### **AGT-077: GET /agents/{agent_id}/notifications**

**Purpose**: Agent notification history

**Request**:
```
/agents/AGT-2026-000567/notifications?from_date=01-01-2026&notification_type=EMAIL&page=1&limit=50
```

**Response**:
```json
{
  "notifications": [
    {
      "notification_id": "uuid",
      "type": "EMAIL",
      "template": "LICENSE_EXPIRY_REMINDER",
      "sent_at": "2026-01-27T10:30:00Z",
      "status": "DELIVERED"
    }
  ],
  "pagination": {
    "current_page": 1,
    "total_pages": 5
  }
}
```

**Business Logic**:
1. Query `agent_notifications` table
2. Apply filters (date range, notification type)
3. Paginate results
4. Order by sent_at DESC

---

### **Phase 9 Repository Methods Needed**

```go
// repo/postgres/agent_profile.go
func (r *AgentProfileRepository) Search(
    ctx context.Context,
    filters SearchFilters,
    page int,
    limit int,
) ([]domain.AgentSearchResult, *PaginationMetadata, error)

func (r *AgentProfileRepository) GetProfileWithRelatedEntities(
    ctx context.Context,
    agentID string,
) (*domain.AgentProfileComplete, error)

func (r *AgentProfileRepository) GetHierarchy(
    ctx context.Context,
    agentID string,
) ([]domain.HierarchyNode, error)

// repo/postgres/agent_audit_log.go
func (r *AuditLogRepository) GetHistory(
    ctx context.Context,
    agentID string,
    fromDate *time.Time,
    toDate *time.Time,
    page int,
    limit int,
) ([]domain.AgentAuditLog, *PaginationMetadata, error)

func (r *AuditLogRepository) GetTimeline(
    ctx context.Context,
    agentID string,
    activityType *string,
    fromDate *time.Time,
    toDate *time.Time,
    page int,
    limit int,
) ([]domain.TimelineEvent, *PaginationMetadata, error)

// repo/postgres/agent_notification.go (NEW)
func (r *NotificationRepository) GetByAgentID(
    ctx context.Context,
    agentID string,
    notificationType *string,
    fromDate *time.Time,
    toDate *time.Time,
    page int,
    limit int,
) ([]domain.AgentNotification, *PaginationMetadata, error)
```

---

### **Phase 9 Database Tables Needed**

```sql
-- Migration: 004_agent_search_dashboard.sql

CREATE TABLE agent_notifications (
    notification_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agent_profiles(agent_id),
    notification_type VARCHAR(50) NOT NULL, -- EMAIL, SMS, INTERNAL
    template VARCHAR(100) NOT NULL,
    recipient VARCHAR(255) NOT NULL,
    sent_at TIMESTAMP WITH TIME ZONE NOT NULL,
    delivered_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) NOT NULL DEFAULT 'SENT', -- SENT, DELIVERED, FAILED
    metadata JSONB,
    CONSTRAINT fk_notification_agent FOREIGN KEY (agent_id) REFERENCES agent_profiles(agent_id),
    CONSTRAINT chk_notification_type CHECK (notification_type IN ('EMAIL', 'SMS', 'INTERNAL')),
    CONSTRAINT chk_notification_status CHECK (status IN ('SENT', 'DELIVERED', 'FAILED'))
);

-- Indexes for search optimization
CREATE INDEX idx_notifications_agent_id ON agent_notifications(agent_id);
CREATE INDEX idx_notifications_sent_at ON agent_notifications(sent_at);
CREATE INDEX idx_notifications_type ON agent_notifications(notification_type);

-- Composite index for common search patterns
CREATE INDEX idx_agent_search ON agent_profiles(status, agent_type, office_code);
CREATE INDEX idx_agent_pan ON agent_profiles(pan_number);
CREATE INDEX idx_contact_mobile ON agent_contacts(mobile_number);
```

---

## 🟢 PHASE 10: BATCH & WEBHOOK APIs

**Scope**: 5 endpoints covering batch deactivation, export, and webhook

### **Business Rules**

| Rule ID | Description | Enforcement |
|---------|-------------|-------------|
| **BR-AGT-PRF-013** | Auto-Deactivation on License Expiry | Batch job runs daily, deactivates agents with expired licenses |

### **Functional Requirements**

| FR ID | Description | APIs |
|-------|-------------|------|
| **FR-AGT-PRF-012** | License Auto-Deactivation | AGT-038 |
| **FR-AGT-PRF-025** | Profile Export | AGT-064 to AGT-067 |

### **Workflows**

| WF ID | Workflow Name | Purpose | APIs |
|-------|---------------|---------|------|
| **WF-AGT-PRF-007** | License Deactivation | Batch workflow for expired license deactivation | AGT-038 |
| **WF-AGT-PRF-012** | Profile Export | Long-running workflow for large exports | AGT-064 to AGT-067 |

---

### **API Details**

#### **AGT-038: POST /licenses/expired** (Batch)

**Purpose**: Batch deactivate agents with expired licenses (scheduled job)

**Request**:
```json
{
  "batch_date": "27-01-2026",
  "dry_run": false
}
```

**Response**:
```json
{
  "batch_id": "uuid",
  "batch_date": "27-01-2026",
  "agents_deactivated": 45,
  "agent_ids": [
    "AGT-2026-000567",
    "AGT-2026-000568"
  ],
  "notifications_sent": 45,
  "dry_run": false
}
```

**Business Logic**:
1. Find all agents with licenses where `expiry_date < batch_date`
2. If `dry_run=true`: Return list without updating
3. If `dry_run=false`:
   - Start WF-AGT-PRF-007 (License Deactivation Workflow)
   - For each agent:
     - ACT-DEC-001: Update status to DEACTIVATED
     - ACT-DEC-002: Disable portal access
     - ACT-DEC-003: Stop commission processing
     - ACT-DEC-004: Send notification
     - ACT-DEC-005: Create audit log
4. Return batch summary

**Database Updates**:
```sql
-- Find expired licenses
WITH expired_agents AS (
    SELECT DISTINCT l.agent_id
    FROM agent_licenses l
    WHERE l.expiry_date < $1
    AND l.status = 'ACTIVE'
)
UPDATE agent_profiles
SET
    status = 'DEACTIVATED',
    deactivation_reason = 'LICENSE_EXPIRED',
    deactivated_at = NOW(),
    commission_enabled = false
WHERE agent_id IN (SELECT agent_id FROM expired_agents)
RETURNING *;
```

**Performance**: Use batch processing, process 100 agents at a time

---

#### **AGT-064: POST /agents/export/configure**

**Purpose**: Configure export parameters (filters, fields, format)

**Request**:
```json
{
  "export_name": "Active Advisors Report",
  "filters": {
    "status": "ACTIVE",
    "office_code": "OFF-001",
    "from_date": "01-01-2026",
    "to_date": "31-01-2026"
  },
  "fields": [
    "agent_id",
    "name",
    "pan_number",
    "mobile_number",
    "email",
    "status"
  ],
  "output_format": "EXCEL"
}
```

**Response**:
```json
{
  "export_config_id": "uuid",
  "export_name": "Active Advisors Report",
  "estimated_records": 500,
  "estimated_time_seconds": 30
}
```

**Business Logic**:
1. Validate filters and fields
2. Estimate record count based on filters
3. Store export configuration
4. Return config ID for execution

**Database Updates**:
- Insert into `agent_export_configs` table

---

#### **AGT-065: POST /agents/export/execute**

**Purpose**: Execute export asynchronously (starts workflow)

**Request**:
```json
{
  "export_config_id": "uuid",
  "requested_by": "admin@example.com"
}
```

**Response** (202 Accepted):
```json
{
  "export_id": "uuid",
  "status": "IN_PROGRESS",
  "message": "Export started, check status via AGT-066"
}
```

**Business Logic**:
1. Fetch export configuration
2. Start WF-AGT-PRF-012 (Profile Export Workflow)
3. Workflow activities:
   - ACT-EXP-001: Fetch data based on filters
   - ACT-EXP-002: Generate file (Excel/PDF)
   - ACT-EXP-003: Upload to storage
   - ACT-EXP-004: Update export status to COMPLETED
   - ACT-EXP-005: Send notification with download link
4. Return export_id for status tracking

**Database Updates**:
- Insert into `agent_export_jobs` table with status='IN_PROGRESS'

---

#### **AGT-066: GET /agents/export/{export_id}/status**

**Purpose**: Check export status (polling endpoint)

**Response**:
```json
{
  "export_id": "uuid",
  "status": "COMPLETED",
  "progress_percentage": 100,
  "file_url": "https://storage.example.com/exports/report_uuid.xlsx",
  "completed_at": "2026-01-27T10:35:00Z"
}
```

**Business Logic**:
1. Fetch export job from database
2. Return current status and progress
3. If completed, return file URL

---

#### **AGT-067: GET /agents/export/{export_id}/download**

**Purpose**: Download exported file

**Response**: Binary file (Excel or PDF)

**Business Logic**:
1. Verify export status is COMPLETED
2. Fetch file URL from export job
3. Stream file to client
4. Set appropriate Content-Type header

---

#### **AGT-078: POST /webhooks/hrms/employee-update** (Webhook)

**Purpose**: Receive HRMS employee updates (external system integration)

**Request**:
```json
{
  "event_id": "uuid",
  "event_type": "EMPLOYEE_UPDATED",
  "timestamp": "2026-01-27T10:30:00Z",
  "employee_data": {
    "employee_id": "EMP12345",
    "name": "John Doe",
    "designation": "Assistant Manager",
    "office_code": "OFF-001",
    "status": "ACTIVE"
  }
}
```

**Response**:
```json
{
  "status": "PROCESSED",
  "event_id": "uuid"
}
```

**Business Logic**:
1. Validate webhook signature (security)
2. Find agent by employee_id
3. Process based on event_type:
   - **EMPLOYEE_CREATED**: Create new agent profile draft
   - **EMPLOYEE_UPDATED**: Update agent profile with HRMS data
   - **EMPLOYEE_TRANSFERRED**: Update office_code
   - **EMPLOYEE_TERMINATED**: Mark agent as TERMINATED
4. Create audit log
5. Send acknowledgment

**Error Handling**:
- If employee_id not found: Log and acknowledge (don't fail)
- If processing fails: Return 500 for retry

**Integration**: INT-AGT-001 (HRMS System)

---

### **Phase 10 Repository Methods Needed**

```go
// repo/postgres/agent_license.go
func (r *LicenseRepository) FindExpiredLicenses(
    ctx context.Context,
    asOfDate time.Time,
) ([]domain.AgentLicense, error)

func (r *LicenseRepository) BatchDeactivateAgents(
    ctx context.Context,
    agentIDs []string,
    reason string,
) (int, error)

// repo/postgres/agent_export.go (NEW)
func (r *ExportRepository) CreateConfig(
    ctx context.Context,
    config *domain.ExportConfig,
) (*domain.ExportConfig, error)

func (r *ExportRepository) CreateJob(
    ctx context.Context,
    job *domain.ExportJob,
) (*domain.ExportJob, error)

func (r *ExportRepository) UpdateJobStatusReturning(
    ctx context.Context,
    exportID string,
    status string,
    progress int,
    fileURL *string,
) (*domain.ExportJob, error)

func (r *ExportRepository) GetJobByID(
    ctx context.Context,
    exportID string,
) (*domain.ExportJob, error)
```

---

### **Phase 10 Domain Models Needed**

```go
// core/domain/agent_export_config.go (NEW)
type AgentExportConfig struct {
    ExportConfigID   string                 `db:"export_config_id"`
    ExportName       string                 `db:"export_name"`
    Filters          map[string]interface{} `db:"filters"` // JSON
    Fields           []string               `db:"fields"`  // JSON array
    OutputFormat     string                 `db:"output_format"`
    EstimatedRecords int                    `db:"estimated_records"`
    CreatedAt        time.Time              `db:"created_at"`
}

// core/domain/agent_export_job.go (NEW)
type AgentExportJob struct {
    ExportID          string     `db:"export_id"`
    ExportConfigID    string     `db:"export_config_id"`
    RequestedBy       string     `db:"requested_by"`
    Status            string     `db:"status"` // IN_PROGRESS, COMPLETED, FAILED
    ProgressPercentage int       `db:"progress_percentage"`
    FileURL           *string    `db:"file_url"`
    StartedAt         time.Time  `db:"started_at"`
    CompletedAt       *time.Time `db:"completed_at"`
    ErrorMessage      *string    `db:"error_message"`
}
```

---

### **Phase 10 Database Tables Needed**

```sql
-- Migration: 005_agent_batch_webhook.sql

CREATE TABLE agent_export_configs (
    export_config_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    export_name VARCHAR(255) NOT NULL,
    filters JSONB NOT NULL,
    fields JSONB NOT NULL,
    output_format VARCHAR(50) NOT NULL,
    estimated_records INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT chk_output_format CHECK (output_format IN ('EXCEL', 'PDF'))
);

CREATE TABLE agent_export_jobs (
    export_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    export_config_id UUID NOT NULL REFERENCES agent_export_configs(export_config_id),
    requested_by VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'IN_PROGRESS',
    progress_percentage INTEGER NOT NULL DEFAULT 0,
    file_url TEXT,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    CONSTRAINT fk_export_config FOREIGN KEY (export_config_id) REFERENCES agent_export_configs(export_config_id),
    CONSTRAINT chk_export_status CHECK (status IN ('IN_PROGRESS', 'COMPLETED', 'FAILED')),
    CONSTRAINT chk_progress CHECK (progress_percentage BETWEEN 0 AND 100)
);

CREATE TABLE hrms_webhook_events (
    event_id UUID PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL,
    employee_id VARCHAR(50) NOT NULL,
    employee_data JSONB NOT NULL,
    received_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) NOT NULL DEFAULT 'RECEIVED',
    error_message TEXT,
    CONSTRAINT chk_event_type CHECK (event_type IN ('EMPLOYEE_CREATED', 'EMPLOYEE_UPDATED', 'EMPLOYEE_TRANSFERRED', 'EMPLOYEE_TERMINATED')),
    CONSTRAINT chk_webhook_status CHECK (status IN ('RECEIVED', 'PROCESSED', 'FAILED'))
);

-- Indexes
CREATE INDEX idx_export_jobs_status ON agent_export_jobs(status);
CREATE INDEX idx_export_jobs_requested_by ON agent_export_jobs(requested_by);
CREATE INDEX idx_webhook_events_employee_id ON hrms_webhook_events(employee_id);
CREATE INDEX idx_webhook_events_received_at ON hrms_webhook_events(received_at);
```

---

## 📊 DATABASE REQUIREMENTS SUMMARY

### **New Tables Required**

| Table | Phase | Purpose |
|-------|-------|---------|
| `agent_termination_records` | 8 | Store termination history |
| `agent_reinstatement_requests` | 8 | Store reinstatement requests with approval workflow |
| `agent_data_archives` | 8 | 7-year data retention for terminated agents |
| `agent_documents` | 8 | Store uploaded documents (reinstatement, KYC) |
| `agent_reinstatement_documents` | 8 | Join table for request-document mapping |
| `agent_notifications` | 9 | Store notification history |
| `agent_export_configs` | 10 | Store export configurations |
| `agent_export_jobs` | 10 | Track export job status |
| `hrms_webhook_events` | 10 | Log HRMS webhook events |

### **Updates to Existing Tables**

```sql
-- agent_profiles (add columns)
ALTER TABLE agent_profiles ADD COLUMN termination_date DATE;
ALTER TABLE agent_profiles ADD COLUMN termination_reason TEXT;
ALTER TABLE agent_profiles ADD COLUMN termination_reason_code VARCHAR(50);
ALTER TABLE agent_profiles ADD COLUMN reinstatement_date DATE;
ALTER TABLE agent_profiles ADD COLUMN deactivated_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE agent_profiles ADD COLUMN deactivation_reason VARCHAR(100);
ALTER TABLE agent_profiles ADD COLUMN commission_enabled BOOLEAN DEFAULT true;

-- agent_authentication (if not exists, create)
CREATE TABLE IF NOT EXISTS agent_authentication (
    agent_id UUID PRIMARY KEY REFERENCES agent_profiles(agent_id),
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    last_login TIMESTAMP WITH TIME ZONE,
    failed_login_attempts INTEGER DEFAULT 0,
    locked_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT chk_auth_status CHECK (status IN ('ACTIVE', 'DISABLED', 'LOCKED'))
);
```

---

## 🔧 REPOSITORY METHODS SUMMARY

### **Phase 8 Repository Methods**

```go
// AgentProfileRepository
TerminateAgentReturning(ctx, agentID, reason, reasonCode, date, terminatedBy) (*Profile, error)
ReinstateAgentReturning(ctx, agentID, date) (*Profile, error)

// AgentTerminationRecordRepository (NEW)
Create(ctx, record) (*TerminationRecord, error)
GetByAgentID(ctx, agentID) (*TerminationRecord, error)

// AgentReinstatementRequestRepository (NEW)
Create(ctx, request) (*ReinstatementRequest, error)
FindByID(ctx, requestID) (*ReinstatementRequest, error)
ApproveReturning(ctx, requestID, approvedBy, comments) (*ReinstatementRequest, error)
RejectReturning(ctx, requestID, rejectedBy, comments) (*ReinstatementRequest, error)

// AgentDataArchiveRepository (NEW)
Create(ctx, archive) (*DataArchive, error)
GetByAgentID(ctx, agentID) (*DataArchive, error)

// AgentDocumentRepository (NEW)
Create(ctx, document) (*Document, error)
LinkToReinstatementRequest(ctx, requestID, documentID) error
```

### **Phase 9 Repository Methods**

```go
// AgentProfileRepository
Search(ctx, filters, page, limit) ([]SearchResult, *Pagination, error)
GetProfileWithRelatedEntities(ctx, agentID) (*ProfileComplete, error)
GetHierarchy(ctx, agentID) ([]HierarchyNode, error)

// AgentAuditLogRepository
GetHistory(ctx, agentID, fromDate, toDate, page, limit) ([]AuditLog, *Pagination, error)
GetTimeline(ctx, agentID, activityType, fromDate, toDate, page, limit) ([]TimelineEvent, *Pagination, error)

// AgentNotificationRepository (NEW)
GetByAgentID(ctx, agentID, notificationType, fromDate, toDate, page, limit) ([]Notification, *Pagination, error)
```

### **Phase 10 Repository Methods**

```go
// AgentLicenseRepository
FindExpiredLicenses(ctx, asOfDate) ([]License, error)
BatchDeactivateAgents(ctx, agentIDs, reason) (int, error)

// AgentExportRepository (NEW)
CreateConfig(ctx, config) (*ExportConfig, error)
CreateJob(ctx, job) (*ExportJob, error)
UpdateJobStatusReturning(ctx, exportID, status, progress, fileURL) (*ExportJob, error)
GetJobByID(ctx, exportID) (*ExportJob, error)

// HRMSWebhookRepository (NEW)
Create(ctx, event) (*WebhookEvent, error)
UpdateStatusReturning(ctx, eventID, status, errorMsg) (*WebhookEvent, error)
```

---

## ⚙️ TEMPORAL WORKFLOWS REQUIRED

### **Phase 8 Workflows**

#### **WF-AGT-PRF-004: Termination Workflow**

```go
// workflows/agent_termination_workflow.go

func AgentTerminationWorkflow(ctx workflow.Context, input TerminationInput) (*TerminationOutput, error) {
    // Step 1: Record workflow start
    // Step 2: Validate termination request
    // Step 3: Update agent status to TERMINATED
    // Step 4: Disable portal access
    // Step 5: Stop commission processing
    // Step 6: Generate termination letter
    // Step 7: Archive agent data (7-year retention)
    // Step 8: Send notifications
    // Step 9: Create audit log
}

// Activities:
// ACT-TRM-001: ValidateTerminationActivity
// ACT-TRM-002: UpdateAgentStatusActivity
// ACT-TRM-003: DisablePortalAccessActivity
// ACT-TRM-004: StopCommissionProcessingActivity
// ACT-TRM-005: GenerateTerminationLetterActivity
// ACT-TRM-006: ArchiveAgentDataActivity
// ACT-TRM-007: SendNotificationActivity
// ACT-TRM-008: CreateAuditLogActivity
```

#### **WF-AGT-PRF-011: Reinstatement Workflow**

```go
// workflows/agent_reinstatement_workflow.go

func AgentReinstatementWorkflow(ctx workflow.Context, input ReinstatementInput) (*ReinstatementOutput, error) {
    // Step 1: Record workflow start
    // Step 2: Create reinstatement request
    // Step 3: Start approval child workflow (wait for signal)
    // Step 4: If approved:
    //         - Update agent status to ACTIVE
    //         - Enable portal access
    //         - Resume commission processing
    //         - Send notifications
    //         - Create audit log
    // Step 5: If rejected:
    //         - Update request status to REJECTED
    //         - Send rejection notification
}

// Activities:
// ACT-RNS-001: CreateReinstatementRequestActivity
// ACT-RNS-002: UpdateAgentStatusActivity
// ACT-RNS-003: EnablePortalAccessActivity
// ACT-RNS-004: ResumeCommissionProcessingActivity
// ACT-RNS-005: SendNotificationActivity
// ACT-RNS-006: CreateAuditLogActivity
```

### **Phase 10 Workflows**

#### **WF-AGT-PRF-007: License Deactivation Workflow**

```go
// workflows/license_deactivation_workflow.go

func LicenseDeactivationWorkflow(ctx workflow.Context, input DeactivationInput) (*DeactivationOutput, error) {
    // Step 1: Record workflow start
    // Step 2: Find agents with expired licenses
    // Step 3: Batch process (100 agents at a time):
    //         - Update status to DEACTIVATED
    //         - Disable portal access
    //         - Stop commission processing
    //         - Send notification
    //         - Create audit log
    // Step 4: Return batch summary
}

// Activities:
// ACT-DEC-001: FindExpiredLicensesActivity
// ACT-DEC-002: BatchUpdateAgentStatusActivity
// ACT-DEC-003: BatchDisablePortalAccessActivity
// ACT-DEC-004: BatchStopCommissionProcessingActivity
// ACT-DEC-005: BatchSendNotificationsActivity
// ACT-DEC-006: CreateBatchAuditLogsActivity
```

#### **WF-AGT-PRF-012: Profile Export Workflow**

```go
// workflows/profile_export_workflow.go

func ProfileExportWorkflow(ctx workflow.Context, input ExportInput) (*ExportOutput, error) {
    // Step 1: Record workflow start
    // Step 2: Fetch export configuration
    // Step 3: Fetch data based on filters (paginated)
    // Step 4: Generate file (Excel or PDF)
    // Step 5: Upload to storage
    // Step 6: Update export job status to COMPLETED
    // Step 7: Send notification with download link
}

// Activities:
// ACT-EXP-001: FetchExportConfigActivity
// ACT-EXP-002: FetchDataActivity
// ACT-EXP-003: GenerateFileActivity
// ACT-EXP-004: UploadToStorageActivity
// ACT-EXP-005: UpdateExportJobStatusActivity
// ACT-EXP-006: SendNotificationActivity
```

---

## 🔗 INTEGRATION POINTS

### **INT-AGT-001: HRMS System Integration**

**Purpose**: Sync employee data from HRMS

**APIs**: AGT-078 (Webhook receiver)

**Implementation**:
```go
// integration/hrms/client.go
type HRMSClient interface {
    ValidateEmployeeID(ctx context.Context, employeeID string) (*EmployeeData, error)
    GetEmployeeByID(ctx context.Context, employeeID string) (*EmployeeData, error)
}

// Webhook signature validation
func ValidateHRMSWebhookSignature(payload []byte, signature string, secret string) bool {
    // HMAC-SHA256 signature validation
    expectedSignature := hmac.SHA256(payload, secret)
    return hmac.Equal(expectedSignature, signature)
}
```

**Error Handling**:
- Invalid signature: Return 401 Unauthorized
- Employee not found: Log and acknowledge (200 OK)
- Processing error: Return 500 for retry

---

### **INT-AGT-005: Notification Service Integration**

**Purpose**: Send email/SMS notifications

**APIs**: All phases (termination, reinstatement, export, etc.)

**Implementation**:
```go
// integration/notification/client.go
type NotificationClient interface {
    SendEmail(ctx context.Context, req EmailRequest) error
    SendSMS(ctx context.Context, req SMSRequest) error
}

type EmailRequest struct {
    To       []string
    Template string
    Data     map[string]interface{}
}

// Activity wrapper
func (a *Activities) SendNotificationActivity(ctx context.Context, input NotificationInput) error {
    logger := activity.GetLogger(ctx)

    // Send notification
    err := a.notificationClient.SendEmail(ctx, EmailRequest{
        To:       input.Recipients,
        Template: input.Template,
        Data:     input.TemplateData,
    })
    if err != nil {
        logger.Error("Failed to send notification", "error", err)
        return err
    }

    // Record in database
    _, err = a.notificationRepo.Create(ctx, &domain.AgentNotification{
        AgentID:          input.AgentID,
        NotificationType: "EMAIL",
        Template:         input.Template,
        Recipient:        input.Recipients[0],
        SentAt:           time.Now(),
        Status:           "SENT",
    })

    return err
}
```

---

### **INT-AGT-006: Letter Generation Service Integration**

**Purpose**: Generate termination letters (PDF)

**APIs**: AGT-040 (Get termination letter)

**Implementation**:
```go
// integration/letter/client.go
type LetterClient interface {
    GenerateTerminationLetter(ctx context.Context, req LetterRequest) ([]byte, error)
}

type LetterRequest struct {
    AgentID           string
    AgentName         string
    TerminationReason string
    EffectiveDate     time.Time
    TerminatedBy      string
}

// Activity wrapper
func (a *Activities) GenerateTerminationLetterActivity(ctx context.Context, input LetterInput) (*LetterOutput, error) {
    // Generate letter
    letterPDF, err := a.letterClient.GenerateTerminationLetter(ctx, LetterRequest{
        AgentID:           input.AgentID,
        AgentName:         input.AgentName,
        TerminationReason: input.Reason,
        EffectiveDate:     input.EffectiveDate,
        TerminatedBy:      input.TerminatedBy,
    })
    if err != nil {
        return nil, err
    }

    // Upload to storage
    fileURL, err := a.storageClient.Upload(ctx, letterPDF, "termination-letters/")
    if err != nil {
        return nil, err
    }

    return &LetterOutput{
        LetterURL: fileURL,
    }, nil
}
```

---

## ✅ IMPLEMENTATION CHECKLIST

### **Phase 8: Status Management APIs**

#### **Database & Domain**
- [ ] Create migration `003_agent_status_management.sql`
- [ ] Create domain model `agent_termination_record.go`
- [ ] Create domain model `agent_reinstatement_request.go`
- [ ] Create domain model `agent_data_archive.go`
- [ ] Create domain model `agent_document.go`

#### **Repositories**
- [ ] Create `agent_termination_record.go` repository
- [ ] Create `agent_reinstatement_request.go` repository
- [ ] Create `agent_data_archive.go` repository
- [ ] Create `agent_document.go` repository
- [ ] Add methods to `agent_profile.go`: TerminateAgentReturning, ReinstateAgentReturning

#### **Workflows**
- [ ] Create `agent_termination_workflow.go` (WF-AGT-PRF-004)
- [ ] Create `agent_reinstatement_workflow.go` (WF-AGT-PRF-011)
- [ ] Create termination activities file with 8 activities
- [ ] Create reinstatement activities file with 6 activities
- [ ] Register workflows and activities in bootstrap

#### **Handlers**
- [ ] Create `handler/status_management.go`
- [ ] Implement AGT-039: TerminateAgent
- [ ] Implement AGT-040: GetTerminationLetter
- [ ] Implement AGT-041: ReinstateAgent
- [ ] Implement AGT-060: CreateReinstatementRequest
- [ ] Implement AGT-061: ApproveReinstatement
- [ ] Implement AGT-062: RejectReinstatement
- [ ] Implement AGT-063: UploadReinstatementDocuments
- [ ] Implement AGT-070: GetStatusTypes (lookup)
- [ ] Implement AGT-071: GetReinstatementReasons (lookup)
- [ ] Implement AGT-072: GetTerminationReasons (lookup)

#### **Request/Response DTOs**
- [ ] Add request structs to `handler/request.go`
- [ ] Add response structs to `handler/response/status_management.go`

#### **Integrations**
- [ ] Create `integration/letter/client.go` (INT-AGT-006)
- [ ] Mock letter generation service for testing

#### **Testing**
- [ ] Test termination workflow end-to-end
- [ ] Test reinstatement approval flow
- [ ] Test reinstatement rejection flow
- [ ] Test document upload
- [ ] Test lookup APIs

---

### **Phase 9: Search & Dashboard APIs**

#### **Database & Domain**
- [ ] Create migration `004_agent_search_dashboard.sql`
- [ ] Create domain model `agent_notification.go`
- [ ] Add search result DTOs
- [ ] Add timeline event DTOs

#### **Repositories**
- [ ] Add methods to `agent_profile.go`: Search, GetProfileWithRelatedEntities, GetHierarchy
- [ ] Add methods to `agent_audit_log.go`: GetHistory, GetTimeline
- [ ] Create `agent_notification.go` repository

#### **Handlers**
- [ ] Create `handler/search_dashboard.go`
- [ ] Implement AGT-022: SearchAgents
- [ ] Implement AGT-023: GetAgentProfile
- [ ] Implement AGT-028: GetAuditHistory
- [ ] Implement AGT-068: GetAgentDashboard
- [ ] Implement AGT-073: GetAgentHierarchy
- [ ] Implement AGT-076: GetAgentTimeline
- [ ] Implement AGT-077: GetAgentNotifications

#### **Request/Response DTOs**
- [ ] Add request structs to `handler/request.go`
- [ ] Add response structs to `handler/response/search_dashboard.go`

#### **Performance Optimization**
- [ ] Create composite indexes for search
- [ ] Optimize GetProfileWithRelatedEntities with JSON aggregation
- [ ] Optimize hierarchy query with recursive CTE
- [ ] Add pagination metadata helper

#### **Testing**
- [ ] Test search with multiple filters
- [ ] Test profile fetch with all related entities
- [ ] Test pagination
- [ ] Test hierarchy chain
- [ ] Test audit history
- [ ] Test timeline with filters

---

### **Phase 10: Batch & Webhook APIs**

#### **Database & Domain**
- [ ] Create migration `005_agent_batch_webhook.sql`
- [ ] Create domain model `agent_export_config.go`
- [ ] Create domain model `agent_export_job.go`
- [ ] Create domain model `hrms_webhook_event.go`

#### **Repositories**
- [ ] Add methods to `agent_license.go`: FindExpiredLicenses, BatchDeactivateAgents
- [ ] Create `agent_export.go` repository
- [ ] Create `hrms_webhook_event.go` repository

#### **Workflows**
- [ ] Create `license_deactivation_workflow.go` (WF-AGT-PRF-007)
- [ ] Create `profile_export_workflow.go` (WF-AGT-PRF-012)
- [ ] Create deactivation activities file with 6 activities
- [ ] Create export activities file with 6 activities
- [ ] Register workflows and activities in bootstrap

#### **Handlers**
- [ ] Create `handler/batch_webhook.go`
- [ ] Implement AGT-038: BatchDeactivateExpiredLicenses
- [ ] Implement AGT-064: ConfigureExport
- [ ] Implement AGT-065: ExecuteExport
- [ ] Implement AGT-066: GetExportStatus
- [ ] Implement AGT-067: DownloadExport
- [ ] Implement AGT-078: HRMSWebhook

#### **Request/Response DTOs**
- [ ] Add request structs to `handler/request.go`
- [ ] Add response structs to `handler/response/batch_webhook.go`

#### **Integrations**
- [ ] Create HMAC webhook signature validation
- [ ] Create file storage client for exports
- [ ] Create Excel generation utility
- [ ] Create PDF generation utility

#### **Testing**
- [ ] Test batch deactivation with dry_run
- [ ] Test export configuration
- [ ] Test export execution (async)
- [ ] Test export status polling
- [ ] Test file download
- [ ] Test HRMS webhook with all event types
- [ ] Test webhook signature validation

---

## 🎯 CRITICAL PATTERNS TO APPLY

### **1. Single Database Round Trip**
```go
// ✅ GOOD - Single UPDATE...RETURNING
result, err := repo.TerminateAgentReturning(ctx, agentID, reason, ...)

// ❌ BAD - Two round trips
err := repo.TerminateAgent(ctx, agentID, reason, ...)
result, err := repo.GetByID(ctx, agentID)
```

### **2. Bulk Operations with UNNEST**
```go
// ✅ GOOD - Batch deactivate N agents in single query
INSERT INTO agent_audit_logs (agent_id, action_type)
SELECT agent_id, 'DEACTIVATED'
FROM UNNEST($1::uuid[]) AS agent_id
```

### **3. Workflow Self-Recording**
```go
// Step 0: ALWAYS record workflow start as FIRST activity
err := workflow.ExecuteActivity(ctx, RecordWorkflowStartActivity, ...)
```

### **4. JSON Aggregation for Related Entities**
```sql
-- Single query fetches profile + addresses + contacts
SELECT
    p.*,
    json_agg(DISTINCT a.*) FILTER (WHERE a.address_id IS NOT NULL) AS addresses,
    json_agg(DISTINCT c.*) FILTER (WHERE c.contact_id IS NOT NULL) AS contacts
FROM agent_profiles p
LEFT JOIN agent_addresses a ON p.agent_id = a.agent_id
LEFT JOIN agent_contacts c ON p.agent_id = c.agent_id
WHERE p.agent_id = $1
GROUP BY p.agent_id;
```

### **5. Human-in-the-Loop with Child Workflows**
```go
// AGT-061/062 send signals to workflow
err := temporalClient.SignalWorkflow(ctx, workflowID, runID, "reinstatement-decision", decision)
```

### **6. Webhook Security**
```go
// ALWAYS validate webhook signatures
if !ValidateHRMSWebhookSignature(payload, signature, secret) {
    return 401, "Invalid signature"
}
```

---

## 📊 SUMMARY

**Total APIs**: 27 endpoints across 3 phases
- Phase 8: 10 APIs (Status Management)
- Phase 9: 7 APIs (Search & Dashboard)
- Phase 10: 5 APIs (Batch & Webhook) + 5 lookups

**Total Workflows**: 4 new workflows
- WF-AGT-PRF-004: Termination Workflow (8 activities)
- WF-AGT-PRF-007: License Deactivation Workflow (6 activities)
- WF-AGT-PRF-011: Reinstatement Workflow (6 activities)
- WF-AGT-PRF-012: Profile Export Workflow (6 activities)

**Total Database Tables**: 9 new tables
- 5 tables for Phase 8 (termination, reinstatement, archives, documents)
- 1 table for Phase 9 (notifications)
- 3 tables for Phase 10 (export configs, jobs, webhook events)

**Total Repository Methods**: ~30 new methods across all repositories

**Integration Points**: 3 external integrations
- INT-AGT-001: HRMS System (webhook)
- INT-AGT-005: Notification Service (email/SMS)
- INT-AGT-006: Letter Generation Service (PDF)

---

**Performance Targets**:
- Single database round trip per operation
- Batch operations using UNNEST (100 agents at a time)
- JSON aggregation for related entities
- Composite indexes for search optimization

**Quality Targets**:
- Zero compilation errors
- All business rules enforced
- All validation rules implemented
- Complete error handling
- Production-ready code

**Testing Promise**: Comprehensive tests in Phase 11

---

**END OF PHASE 8-10 CONTEXT DOCUMENT**
