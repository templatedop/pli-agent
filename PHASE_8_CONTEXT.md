# Phase 8: Status Management APIs - Implementation Context

## Overview
Phase 8 implements **10 endpoints** for comprehensive agent status management including Termination, Reinstatement, and Status Lookups with Temporal workflow orchestration.

**API Range**: AGT-039 to AGT-072 (selective)
**Total Endpoints**: 10
**Temporal Workflows**: 2 (WF-AGT-PRF-004, WF-AGT-PRF-011)

---

## API Endpoints Breakdown

### Part 1: Termination APIs (2 endpoints)

#### AGT-039: POST `/agents/{agent_id}/terminate`
**Purpose**: Terminate agent with comprehensive workflow orchestration

**Business Rules**:
- **BR-AGT-PRF-017**: Agent Termination Workflow
  - Termination reason mandatory (min 20 chars)
  - Effective date >= today
  - Reason codes: RESIGNATION, MISCONDUCT, NON_PERFORMANCE, FRAUD, OTHER
- **VR-AGT-PRF-020**: Termination Reason Mandatory (min 20 chars)
- **VR-AGT-PRF-021**: Termination Date Future or Today

**Workflow Triggers (WF-AGT-PRF-004)**:
1. Update agent status to 'TERMINATED'
2. Disable portal access
3. Generate termination letter
4. Stop commission processing
5. Archive agent data (7-year retention)
6. Send termination notifications
7. Create audit logs

**Request**:
```json
{
  "termination_reason_code": "RESIGNATION",
  "termination_reason_text": "Agent voluntarily resigned due to personal reasons",
  "effective_date": "2026-02-15",
  "terminated_by": "ADMIN_USER_ID"
}
```

**Response**:
```json
{
  "status_code": 201,
  "message": "Agent terminated successfully",
  "termination_id": "uuid",
  "agent_id": "uuid",
  "termination_reason": "RESIGNATION",
  "effective_date": "2026-02-15",
  "termination_letter_url": "/agents/{agent_id}/termination-letter",
  "portal_disabled": true,
  "commission_stopped": true,
  "workflow_id": "temporal_workflow_id"
}
```

#### AGT-040: GET `/agents/{agent_id}/termination-letter`
**Purpose**: Retrieve generated termination letter

**Query Parameters**:
- `format`: pdf | html (default: pdf)

**Response**: PDF or HTML content with termination letter

**Integration**: Letter Generation Service (INT-AGT-006) - **STUB IMPLEMENTATION**

---

### Part 2: Reinstatement APIs (5 endpoints)

#### AGT-041: POST `/agents/{agent_id}/reinstate`
#### AGT-060: POST `/agents/reinstatement-requests`
**Purpose**: Create reinstatement request with approval workflow

**Business Rules**:
- **BR-AGT-PRF-016**: Status updates require mandatory reason (min 10 chars)
  - Reason codes: LICENSE_RENEWED, APPEAL_APPROVED, CLEARANCE_OBTAINED, MUTUAL_AGREEMENT, WRONGFUL_TERMINATION
  - Effective date >= today
- **Eligibility**: Agent status must be TERMINATED or SUSPENDED

**Request**:
```json
{
  "agent_id": "uuid",
  "reinstatement_reason_code": "LICENSE_RENEWED",
  "reinstatement_reason_text": "Agent renewed expired license",
  "effective_date": "2026-02-20",
  "requested_by": "SUPERVISOR_ID"
}
```

**Response**:
```json
{
  "status_code": 201,
  "message": "Reinstatement request created",
  "request_id": "uuid",
  "agent_id": "uuid",
  "request_status": "PENDING",
  "reinstatement_reason": "LICENSE_RENEWED",
  "effective_date": "2026-02-20",
  "workflow_id": "temporal_workflow_id",
  "pending_approval": true
}
```

#### AGT-061: POST `/agents/reinstatement-requests/{request_id}/approve`
**Purpose**: Approve reinstatement request (triggers workflow continuation)

**Workflow Step**: Human-in-the-loop approval (72-hour timeout)

**Actions on Approval**:
1. Update request status to APPROVED
2. Signal Temporal workflow to continue
3. Workflow updates agent status to ACTIVE
4. Enable portal access
5. Resume commission processing
6. Send approval notifications
7. Create audit logs

**Request**:
```json
{
  "approved_by": "SUPERVISOR_ID",
  "approval_comments": "All clearances obtained, approved for reinstatement"
}
```

**Response**:
```json
{
  "status_code": 200,
  "message": "Reinstatement request approved",
  "request_id": "uuid",
  "agent_id": "uuid",
  "request_status": "APPROVED",
  "approved_by": "SUPERVISOR_ID",
  "approved_at": "2026-02-18T10:30:00Z",
  "agent_status_updated": "ACTIVE",
  "portal_enabled": true,
  "commission_resumed": true
}
```

#### AGT-062: POST `/agents/reinstatement-requests/{request_id}/reject`
**Purpose**: Reject reinstatement request (terminates workflow)

**Request**:
```json
{
  "rejected_by": "SUPERVISOR_ID",
  "rejection_reason": "Disciplinary issues not resolved, cannot approve reinstatement"
}
```

**Validation**: Rejection reason mandatory (min 10 chars)

**Response**:
```json
{
  "status_code": 200,
  "message": "Reinstatement request rejected",
  "request_id": "uuid",
  "agent_id": "uuid",
  "request_status": "REJECTED",
  "rejected_by": "SUPERVISOR_ID",
  "rejected_at": "2026-02-18T11:00:00Z",
  "rejection_reason": "Disciplinary issues not resolved"
}
```

#### AGT-063: POST `/agents/reinstatement-requests/{request_id}/documents`
**Purpose**: Upload supporting documents for reinstatement

**Request**: Multipart form data
- `document_type`: LICENSE_RENEWAL, CLEARANCE_CERTIFICATE, APPEAL_APPROVAL, OTHER
- `file`: PDF, JPG, PNG (max 5MB)

**Validation**:
- File type: PDF, JPG, PNG only
- File size: Max 5MB per file
- Document type based on termination reason

**Response**:
```json
{
  "status_code": 201,
  "message": "Document uploaded successfully",
  "document_id": "uuid",
  "request_id": "uuid",
  "document_type": "LICENSE_RENEWAL",
  "file_name": "renewed_license.pdf",
  "file_size": 245678,
  "uploaded_at": "2026-02-18T09:00:00Z"
}
```

---

### Part 3: Lookup APIs (3 endpoints)

#### AGT-070: GET `/agent-status-types`
**Purpose**: Get all agent status types

**Response**:
```json
{
  "status_code": 200,
  "message": "Success",
  "status_types": [
    {"code": "ACTIVE", "name": "Active", "description": "Agent is active and can perform duties"},
    {"code": "SUSPENDED", "name": "Suspended", "description": "Agent temporarily suspended"},
    {"code": "TERMINATED", "name": "Terminated", "description": "Agent contract terminated"},
    {"code": "DEACTIVATED", "name": "Deactivated", "description": "Agent deactivated due to inactivity"}
  ]
}
```

#### AGT-071: GET `/reinstatement-reasons`
**Purpose**: Get all reinstatement reason codes

**Response**:
```json
{
  "status_code": 200,
  "message": "Success",
  "reinstatement_reasons": [
    {"code": "LICENSE_RENEWED", "name": "License Renewed", "description": "Agent renewed expired license"},
    {"code": "APPEAL_APPROVED", "name": "Appeal Approved", "description": "Termination appeal approved by management"},
    {"code": "CLEARANCE_OBTAINED", "name": "Clearance Obtained", "description": "Required clearances obtained"},
    {"code": "MUTUAL_AGREEMENT", "name": "Mutual Agreement", "description": "Reinstatement by mutual agreement"},
    {"code": "WRONGFUL_TERMINATION", "name": "Wrongful Termination", "description": "Termination was wrongful, being corrected"}
  ]
}
```

#### AGT-072: GET `/termination-reasons`
**Purpose**: Get all termination reason codes

**Response**:
```json
{
  "status_code": 200,
  "message": "Success",
  "termination_reasons": [
    {"code": "RESIGNATION", "name": "Resignation", "description": "Agent voluntarily resigned"},
    {"code": "MISCONDUCT", "name": "Misconduct", "description": "Terminated due to misconduct"},
    {"code": "NON_PERFORMANCE", "name": "Non-Performance", "description": "Terminated due to poor performance"},
    {"code": "FRAUD", "name": "Fraud", "description": "Terminated due to fraudulent activities"},
    {"code": "OTHER", "name": "Other", "description": "Other termination reason"}
  ]
}
```

---

## Database Schema

### Table 1: `agent_termination_records`
```sql
CREATE TABLE agent_termination_records (
    termination_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agent_profiles(agent_id),
    termination_reason_code VARCHAR(50) NOT NULL,
    termination_reason_text TEXT NOT NULL CHECK (LENGTH(termination_reason_text) >= 20),
    effective_date DATE NOT NULL CHECK (effective_date >= CURRENT_DATE),
    termination_letter_url TEXT,
    portal_disabled_at TIMESTAMP,
    commission_stopped_at TIMESTAMP,
    archive_id UUID,
    workflow_id VARCHAR(255),
    workflow_status VARCHAR(50),
    created_by VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(100),
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    version INTEGER DEFAULT 1
);

CREATE INDEX idx_agent_termination_agent_id ON agent_termination_records(agent_id);
CREATE INDEX idx_agent_termination_effective_date ON agent_termination_records(effective_date);
```

### Table 2: `agent_reinstatement_requests`
```sql
CREATE TABLE agent_reinstatement_requests (
    request_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agent_profiles(agent_id),
    request_reason_code VARCHAR(50) NOT NULL,
    request_reason_text TEXT NOT NULL CHECK (LENGTH(request_reason_text) >= 10),
    effective_date DATE NOT NULL CHECK (effective_date >= CURRENT_DATE),
    request_status VARCHAR(50) NOT NULL DEFAULT 'PENDING', -- PENDING, APPROVED, REJECTED
    approved_by VARCHAR(100),
    approved_at TIMESTAMP,
    approval_comments TEXT,
    rejected_by VARCHAR(100),
    rejected_at TIMESTAMP,
    rejection_reason TEXT,
    workflow_id VARCHAR(255),
    workflow_status VARCHAR(50),
    created_by VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(100),
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    version INTEGER DEFAULT 1
);

CREATE INDEX idx_agent_reinstatement_agent_id ON agent_reinstatement_requests(agent_id);
CREATE INDEX idx_agent_reinstatement_status ON agent_reinstatement_requests(request_status);
```

### Table 3: `agent_documents`
```sql
CREATE TABLE agent_documents (
    document_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agent_profiles(agent_id),
    reference_type VARCHAR(50) NOT NULL, -- REINSTATEMENT, TERMINATION, etc.
    reference_id UUID NOT NULL, -- FK to reinstatement_request_id or termination_id
    document_type VARCHAR(50) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL,
    file_path TEXT NOT NULL,
    mime_type VARCHAR(100),
    uploaded_by VARCHAR(100) NOT NULL,
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    version INTEGER DEFAULT 1
);

CREATE INDEX idx_agent_documents_reference ON agent_documents(reference_type, reference_id);
CREATE INDEX idx_agent_documents_agent_id ON agent_documents(agent_id);
```

### Table 4: `agent_data_archives`
```sql
CREATE TABLE agent_data_archives (
    archive_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL,
    termination_id UUID NOT NULL REFERENCES agent_termination_records(termination_id),
    archive_date DATE NOT NULL,
    retention_period_years INTEGER DEFAULT 7,
    retention_expiry_date DATE NOT NULL,
    archive_data JSONB NOT NULL,
    archive_size BIGINT,
    created_by VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_agent_archives_agent_id ON agent_data_archives(agent_id);
CREATE INDEX idx_agent_archives_termination ON agent_data_archives(termination_id);
```

---

## Temporal Workflows

### WF-AGT-PRF-004: Termination Workflow

**Trigger**: AGT-039 Terminate Agent API

**Activities**:
1. **ValidateTerminationRequest**
   - Check termination reason (min 20 chars)
   - Validate effective date >= today
   - Verify agent is in ACTIVE status

2. **UpdateAgentStatusToTerminated**
   - Update agent_profiles.agent_status = 'TERMINATED'
   - Update agent_profiles.status_effective_date
   - Create audit log

3. **DisablePortalAccess**
   - Update portal_disabled_at timestamp
   - Invalidate all active sessions
   - Block future login attempts

4. **GenerateTerminationLetter**
   - Call letter generation service (STUB)
   - Store letter URL in termination record

5. **StopCommissionProcessing**
   - Update commission_stopped_at timestamp
   - Notify commission service to stop processing

6. **ArchiveAgentData**
   - Collect all agent data
   - Store in agent_data_archives table
   - Set 7-year retention period

7. **SendTerminationNotifications**
   - Email to agent
   - SMS to registered mobile
   - Notification to supervisors

8. **CreateAuditLog**
   - Log all workflow steps
   - Create compliance records

**Duration**: ~5-10 minutes
**Retry Policy**: 3 attempts with exponential backoff (2s, 4s, 8s)
**Error Handling**: Compensating transactions for rollback

### WF-AGT-PRF-011: Reinstatement Workflow

**Trigger**: AGT-060 Create Reinstatement Request API

**Activities**:
1. **ValidateReinstatementEligibility**
   - Check agent status (TERMINATED or SUSPENDED)
   - Validate reinstatement reason (min 10 chars)
   - Check termination reason compatibility
   - Review time elapsed since termination
   - Check disciplinary history

2. **CollectAndValidateDocuments**
   - Verify required documents uploaded
   - Validate document types based on termination reason
   - Check file size and format

3. **SendApprovalRequest** (Child Workflow - Human-in-the-loop)
   - Create approval task
   - Send notification to supervisors
   - **Wait for signal** (approve/reject) with 72-hour timeout
   - Auto-reject if timeout exceeded

4. **UpdateAgentStatusToActive** (on approval)
   - Update agent_profiles.agent_status = 'ACTIVE'
   - Update agent_profiles.status_effective_date
   - Update reinstatement_request status = 'APPROVED'
   - Create audit log

5. **EnablePortalAndCommission** (on approval)
   - Enable portal access
   - Resume commission processing
   - Restore system permissions

6. **SendNotificationAndAudit**
   - Email to agent (approval/rejection)
   - SMS notification
   - Notify supervisors
   - Create compliance records

**Duration**: Up to 72 hours (waiting for approval)
**Signals**:
- `approve_reinstatement` - From AGT-061
- `reject_reinstatement` - From AGT-062
**Retry Policy**: 3 attempts with 2s exponential backoff
**Timeout**: 72 hours for approval, auto-reject on timeout

---

## Domain Models

### 1. `core/domain/agent_termination.go`
```go
type AgentTermination struct {
    TerminationID         string    `db:"termination_id"`
    AgentID               string    `db:"agent_id"`
    TerminationReasonCode string    `db:"termination_reason_code"`
    TerminationReasonText string    `db:"termination_reason_text"`
    EffectiveDate         time.Time `db:"effective_date"`
    TerminationLetterURL  string    `db:"termination_letter_url"`
    PortalDisabledAt      time.Time `db:"portal_disabled_at"`
    CommissionStoppedAt   time.Time `db:"commission_stopped_at"`
    ArchiveID             string    `db:"archive_id"`
    WorkflowID            string    `db:"workflow_id"`
    WorkflowStatus        string    `db:"workflow_status"`
    CreatedBy             string    `db:"created_by"`
    CreatedAt             time.Time `db:"created_at"`
    UpdatedBy             string    `db:"updated_by"`
    UpdatedAt             time.Time `db:"updated_at"`
    DeletedAt             *time.Time `db:"deleted_at"`
    Version               int       `db:"version"`
}

const (
    TerminationReasonResignation    = "RESIGNATION"
    TerminationReasonMisconduct     = "MISCONDUCT"
    TerminationReasonNonPerformance = "NON_PERFORMANCE"
    TerminationReasonFraud          = "FRAUD"
    TerminationReasonOther          = "OTHER"
)
```

### 2. `core/domain/reinstatement_request.go`
```go
type ReinstatementRequest struct {
    RequestID          string    `db:"request_id"`
    AgentID            string    `db:"agent_id"`
    RequestReasonCode  string    `db:"request_reason_code"`
    RequestReasonText  string    `db:"request_reason_text"`
    EffectiveDate      time.Time `db:"effective_date"`
    RequestStatus      string    `db:"request_status"`
    ApprovedBy         string    `db:"approved_by"`
    ApprovedAt         *time.Time `db:"approved_at"`
    ApprovalComments   string    `db:"approval_comments"`
    RejectedBy         string    `db:"rejected_by"`
    RejectedAt         *time.Time `db:"rejected_at"`
    RejectionReason    string    `db:"rejection_reason"`
    WorkflowID         string    `db:"workflow_id"`
    WorkflowStatus     string    `db:"workflow_status"`
    CreatedBy          string    `db:"created_by"`
    CreatedAt          time.Time `db:"created_at"`
    UpdatedBy          string    `db:"updated_by"`
    UpdatedAt          time.Time `db:"updated_at"`
    DeletedAt          *time.Time `db:"deleted_at"`
    Version            int       `db:"version"`
}

const (
    RequestStatusPending  = "PENDING"
    RequestStatusApproved = "APPROVED"
    RequestStatusRejected = "REJECTED"

    ReinstatementReasonLicenseRenewed       = "LICENSE_RENEWED"
    ReinstatementReasonAppealApproved       = "APPEAL_APPROVED"
    ReinstatementReasonClearanceObtained    = "CLEARANCE_OBTAINED"
    ReinstatementReasonMutualAgreement      = "MUTUAL_AGREEMENT"
    ReinstatementReasonWrongfulTermination  = "WRONGFUL_TERMINATION"
)
```

### 3. `core/domain/agent_document.go`
```go
type AgentDocument struct {
    DocumentID    string    `db:"document_id"`
    AgentID       string    `db:"agent_id"`
    ReferenceType string    `db:"reference_type"`
    ReferenceID   string    `db:"reference_id"`
    DocumentType  string    `db:"document_type"`
    FileName      string    `db:"file_name"`
    FileSize      int64     `db:"file_size"`
    FilePath      string    `db:"file_path"`
    MimeType      string    `db:"mime_type"`
    UploadedBy    string    `db:"uploaded_by"`
    UploadedAt    time.Time `db:"uploaded_at"`
    DeletedAt     *time.Time `db:"deleted_at"`
    Version       int       `db:"version"`
}

const (
    ReferenceTypeReinstatement = "REINSTATEMENT"
    ReferenceTypeTermination   = "TERMINATION"

    DocumentTypeLicenseRenewal      = "LICENSE_RENEWAL"
    DocumentTypeClearanceCertificate = "CLEARANCE_CERTIFICATE"
    DocumentTypeAppealApproval      = "APPEAL_APPROVAL"
    DocumentTypeOther               = "OTHER"
)
```

---

## Repository Layer Patterns

### Audit Logging with CTE Pattern
All write operations MUST include audit logging:

```go
// Example: Create termination with audit log
WITH inserted AS (
    INSERT INTO agent_termination_records (...)
    VALUES (...)
    RETURNING *
)
INSERT INTO agent_audit_logs (agent_id, action_type, field_name, new_value, performed_by, performed_at)
SELECT agent_id, 'AGENT_TERMINATED', 'agent_status', 'TERMINATED', created_by, NOW()
FROM inserted
RETURNING (SELECT ROW(...) FROM inserted)
```

### Single Database Round Trip
- Use CTE for INSERT + audit
- Use UPDATE...RETURNING for atomic operations
- Use dblib.SelectOne, dblib.SelectRows, dblib.Insert, dblib.Update

---

## Implementation Summary

### Total Lines of Code (Estimated):
- Domain Models: ~300 lines (3 files)
- Repositories: ~900 lines (3 files with audit logging)
- Handlers: ~1,200 lines (3 files)
- Response DTOs: ~400 lines (2 files)
- Temporal Workflows: ~800 lines (2 files)
- **Total: ~3,600 lines**

### Critical Business Rules:
- BR-AGT-PRF-016: Status updates require mandatory reason
- BR-AGT-PRF-017: Agent Termination Workflow
- VR-AGT-PRF-020: Termination Reason Mandatory (min 20 chars)
- VR-AGT-PRF-021: Termination Date Future or Today

### Workflow Orchestration:
- WF-AGT-PRF-004: Termination (7 activities)
- WF-AGT-PRF-011: Reinstatement (6 activities with human approval)

---

## Integration Points

### External Services (STUB Implementation):
1. **Letter Generation Service** (INT-AGT-006)
   - Generate termination letters
   - Return letter URL
   - **Implementation**: Return stub URL

2. **File Storage Service**
   - Store uploaded documents
   - **Implementation**: Store file path locally

3. **Commission Service**
   - Stop/resume commission processing
   - **Implementation**: Update timestamp only

4. **Notification Service**
   - Send email/SMS
   - **Implementation**: Log notification intent

---

## Phase 8 Implementation Checklist

- [ ] Domain Models (3 files)
- [ ] Repository Layer with Audit Logging (3 files)
- [ ] Handler Layer (3 files)
- [ ] Response DTOs (2 files)
- [ ] Request DTOs (in handler/request.go)
- [ ] Temporal Workflows (2 files)
- [ ] Bootstrap Registration
- [ ] Context Document (this file)
- [ ] Commit and Push

---

**Generated**: 2026-01-27
**Phase**: 8 - Status Management APIs
**Branch**: claude/phase-8-status-management-jt5yr
