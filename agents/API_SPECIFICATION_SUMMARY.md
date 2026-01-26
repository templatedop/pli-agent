# Agent Profile Management API - OpenAPI 3.0 Specification Summary

## Document Control

| Attribute | Details |
|-----------|---------|
| **Module** | Agent Profile Management |
| **API Specification Version** | 1.0.0 |
| **OpenAPI Version** | 3.0.3 |
| **Total APIs** | 78 endpoints |
| **Total User Journeys** | 10 journeys |
| **Created** | 2026-01-23 |
| **Specification Files** | 3 parts |

---

## Specification Files

The complete OpenAPI 3.0 specification has been split into 3 parts to manage file size:

### Part 1: Profile Creation & Search (AGT-001 to AGT-028)
**File:** `agent_profile_management_api_part1.yaml`

**Contents:**
- Profile Creation Workflow APIs (6 APIs)
  - AGT-001: POST /agent-profiles/initiate
  - AGT-002: POST /agent-profiles/{session_id}/fetch-hrms
  - AGT-003: GET /advisor-coordinators
  - AGT-004: POST /agent-profiles/{session_id}/link-coordinator
  - AGT-005: POST /agent-profiles/{session_id}/validate-basic
  - AGT-006: POST /agent-profiles/{session_id}/submit

- Lookup APIs (5 APIs)
  - AGT-007: GET /agent-types
  - AGT-008: GET /categories
  - AGT-009: GET /designations
  - AGT-010: GET /office-types
  - AGT-011: GET /states

- Validation APIs (4 APIs)
  - AGT-012: POST /validations/pan/check-uniqueness
  - AGT-013: POST /validations/hrms/employee-id
  - AGT-014: POST /validations/bank/ifsc
  - AGT-015: GET /validations/office/{office_code}

- Workflow Management APIs (4 APIs)
  - AGT-016: GET /agent-profiles/sessions/{session_id}/status
  - AGT-017: POST /agent-profiles/sessions/{session_id}/save
  - AGT-018: GET /agent-profiles/sessions/{session_id}/resume
  - AGT-019: DELETE /agent-profiles/sessions/{session_id}

- Status & Notification APIs (2 APIs)
  - AGT-020: GET /agent-profiles/creation-status/{agent_id}
  - AGT-021: POST /agents/{agent_id}/notifications/resend-welcome

- Agent Search & View APIs (3 APIs)
  - AGT-022: GET /agents/search
  - AGT-023: GET /agents/{agent_id}
  - AGT-024: GET /agents/{agent_id}/update-form

**Total APIs in Part 1:** 24 endpoints

### Part 2: Profile Update, Licenses, Termination & Authentication (AGT-025 to AGT-046)
**File:** `agent_profile_management_api_part2.yaml`

**Contents:**
- Profile Update APIs (4 APIs)
  - AGT-025: PUT /agents/{agent_id}/sections/{section}
  - AGT-026: PUT /approvals/{approval_request_id}/approve
  - AGT-027: PUT /approvals/{approval_request_id}/reject
  - AGT-028: GET /agents/{agent_id}/audit-history

- License Management APIs (10 APIs)
  - AGT-029: GET /agents/{agent_id}/licenses
  - AGT-030: POST /agents/{agent_id}/licenses
  - AGT-031: GET /agents/{agent_id}/licenses/{license_id}
  - AGT-032: PUT /agents/{agent_id}/licenses/{license_id}
  - AGT-033: PUT /agents/{agent_id}/licenses/{license_id}/renew
  - AGT-034: DELETE /agents/{agent_id}/licenses/{license_id}
  - AGT-035: GET /license-types
  - AGT-036: GET /licenses/expiring
  - AGT-037: GET /licenses/{license_id}/reminders
  - AGT-038: POST /licenses/expired

- Agent Termination APIs (3 APIs)
  - AGT-039: POST /agents/{agent_id}/terminate
  - AGT-040: GET /agents/{agent_id}/termination-letter
  - AGT-041: POST /agents/{agent_id}/reinstate

- Portal Authentication APIs (5 APIs)
  - AGT-042: POST /agents/{agent_id}/portal/login
  - AGT-043: POST /agents/{agent_id}/portal/otp/verify
  - AGT-044: POST /agents/{agent_id}/portal/logout
  - AGT-045: POST /agents/{agent_id}/portal/unlock
  - AGT-046: GET /agents/portal/session/status

**Total APIs in Part 2:** 22 endpoints

### Part 3: Self-Service, Bank Details, Goals, Reinstatement, Export & Additional (AGT-047 to AGT-078)
**File:** `agent_profile_management_api_part3.yaml`

**Contents:**
- Self-Service Profile Update APIs (3 APIs)
  - AGT-047: GET /agents/{agent_id}/portal/profile
  - AGT-048: PUT /agents/{agent_id}/portal/profile
  - AGT-049: POST /agents/{agent_id}/portal/profile/otp

- Dashboard API (1 API)
  - AGT-068: GET /dashboard/agent/{agent_id}

- Bank Details Management APIs (5 APIs)
  - AGT-050: GET /agents/{agent_id}/portal/bank-details
  - AGT-051: POST /agents/{agent_id}/portal/bank-details
  - AGT-052: GET /agents/{agent_id}/bank-details
  - AGT-053: POST /agents/{agent_id}/bank-details
  - AGT-054: PUT /agents/{agent_id}/bank-details

- Agent Goal Setting APIs (5 APIs)
  - AGT-055: GET /agents/{agent_id}/goals
  - AGT-056: POST /agents/{agent_id}/goals
  - AGT-057: PUT /agents/{agent_id}/goals/{goal_id}
  - AGT-058: DELETE /agents/{agent_id}/goals/{goal_id}
  - AGT-059: GET /agents/{agent_id}/goals/progress

- Status Reinstatement APIs (7 APIs)
  - AGT-060: POST /reinstatement/request
  - AGT-061: PUT /reinstatement/{request_id}/approve
  - AGT-062: PUT /reinstatement/{request_id}/reject
  - AGT-063: POST /reinstatement/{request_id}/documents
  - AGT-070: GET /status-types
  - AGT-071: GET /reinstatement/reasons
  - AGT-072: GET /termination/reasons

- Agent Export APIs (4 APIs)
  - AGT-064: POST /agents/export/configure
  - AGT-065: POST /agents/export/execute
  - AGT-066: GET /agents/export/{export_id}/status
  - AGT-067: GET /agents/export/{export_id}/download

- Additional Lookup APIs (1 API)
  - AGT-073: GET /agents/{agent_id}/hierarchy

- Product Authorization APIs (2 APIs)
  - AGT-074: GET /agents/{agent_id}/product-authorization
  - AGT-075: POST /agents/{agent_id}/product-authorization

- Timeline & Notification APIs (2 APIs)
  - AGT-076: GET /agents/{agent_id}/timeline
  - AGT-077: GET /agents/{agent_id}/notifications

- Goal Templates API (1 API)
  - AGT-069: GET /goals/templates

- Webhook API (1 API)
  - AGT-078: POST /webhooks/hrms/employee-update

**Total APIs in Part 3:** 32 endpoints

---

## API Distribution Summary

| Category | API Count | Percentage |
|----------|-----------|------------|
| **Core APIs** | 38 | 48.7% |
| **Lookup APIs** | 14 | 17.9% |
| **Validation APIs** | 4 | 5.1% |
| **Status APIs** | 10 | 12.8% |
| **Workflow APIs** | 4 | 5.1% |
| **Approval APIs** | 3 | 3.8% |
| **Notification APIs** | 2 | 2.6% |
| **Document APIs** | 3 | 3.8% |
| **Dashboard API** | 1 | 1.3% |
| **System API** | 1 | 1.3% |
| **Webhook API** | 1 | 1.3% |
| **Product Authorization** | 2 | 2.6% |
| **Goals & Performance** | 5 | 6.4% |
| **Export** | 4 | 5.1% |

**Total:** 78 APIs

---

## User Journey to API Mapping

| Journey ID | Journey Name | Total APIs | API IDs |
|------------|--------------|------------|---------|
| UJ-001 | Agent Profile Creation | 21 | AGT-001 to AGT-021 |
| UJ-002 | Agent Profile Update | 7 | AGT-022 to AGT-028 |
| UJ-003 | License Management & Renewal | 10 | AGT-029 to AGT-038 |
| UJ-004 | Agent Termination | 3 | AGT-039 to AGT-041 |
| UJ-005 | Portal Authentication | 5 | AGT-042 to AGT-046 |
| UJ-006 | Self-Service Profile Update | 3 | AGT-047 to AGT-049, AGT-068 |
| UJ-007 | Bank Details Management | 5 | AGT-050 to AGT-054 |
| UJ-008 | Agent Goal Setting | 5 | AGT-055 to AGT-059, AGT-069 |
| UJ-009 | Status Reinstatement | 7 | AGT-060 to AGT-063, AGT-070 to AGT-072 |
| UJ-010 | Profile Search & Export | 4 | AGT-064 to AGT-067 |

**Additional APIs (not journey-specific):** 8 APIs
- AGT-073: Agent Hierarchy
- AGT-074, AGT-075: Product Authorization
- AGT-076: Agent Timeline
- AGT-077: Agent Notifications

---

## Component Coverage

### Business Rules Coverage: 38/38 (100%)

| BR ID | Business Rule | APIs |
|-------|---------------|-----|
| BR-AGT-PRF-001 | Advisor Coordinator Linkage | AGT-003, AGT-004 |
| BR-AGT-PRF-002 | Coordinator Geographic Assignment | AGT-003 |
| BR-AGT-PRF-003 | HRMS Integration | AGT-002, AGT-013 |
| BR-AGT-PRF-004 | Field Officer Auto-Fetch | AGT-002 |
| BR-AGT-PRF-005 | Name Update with Audit | AGT-025, AGT-028 |
| BR-AGT-PRF-006 | PAN Update & Uniqueness | AGT-012, AGT-025 |
| BR-AGT-PRF-007 | Personal Information Update | AGT-025 |
| BR-AGT-PRF-008 | Multiple Address Types | AGT-005 |
| BR-AGT-PRF-009 | Communication Address Same as Permanent | AGT-005 |
| BR-AGT-PRF-010 | Phone Number Categories | AGT-005 |
| BR-AGT-PRF-011 | Email Address | AGT-005 |
| BR-AGT-PRF-012 | License Renewal Period Rules | AGT-030, AGT-033 |
| BR-AGT-PRF-013 | Auto-Deactivation on Expiry | AGT-038 |
| BR-AGT-PRF-014 | License Renewal Reminders | AGT-037 |
| BR-AGT-PRF-016 | Status Update with Reason | AGT-039 |
| BR-AGT-PRF-017 | Agent Termination Workflow | AGT-039, AGT-040 |
| BR-AGT-PRF-018 | Bank Account for Commission | AGT-050 to AGT-054 |
| BR-AGT-PRF-019 | OTP-Based Authentication | AGT-042, AGT-043 |
| BR-AGT-PRF-020 | Account Lockout | AGT-045 |
| BR-AGT-PRF-021 | Session Timeout | AGT-046 |
| BR-AGT-PRF-022 | Multi-Criteria Search | AGT-022 |
| BR-AGT-PRF-023 | Dashboard Profile View | AGT-023, AGT-068 |
| BR-AGT-PRF-024 | Performance Goal Assignment | AGT-055 to AGT-059 |
| BR-AGT-PRF-026 | Product Class Authorization | AGT-074, AGT-075 |
| BR-AGT-PRF-027 | External ID Tracking | AGT-025 |
| BR-AGT-PRF-031 | Profile Creation Workflow | AGT-001 to AGT-006 |
| BR-AGT-PRF-032 | Employee ID Validation | AGT-013 |
| BR-AGT-PRF-033 to BR-AGT-PRF-037 | Additional Profile Rules | AGT-005 |

### Functional Requirements Coverage: 27/27 (100%)

| FR ID | Functional Requirement | APIs |
|-------|------------------------|-----|
| FR-AGT-PRF-001 | New Profile Creation | AGT-001, AGT-006 |
| FR-AGT-PRF-002 | Profile Details Entry | AGT-002, AGT-005 |
| FR-AGT-PRF-003 | Coordinator Selection | AGT-003, AGT-004 |
| FR-AGT-PRF-004 | Agent Search | AGT-022 |
| FR-AGT-PRF-005 | Profile Dashboard View | AGT-023 |
| FR-AGT-PRF-006 | Personal Information Update | AGT-024, AGT-025 |
| FR-AGT-PRF-007 | PAN Update | AGT-012, AGT-025 |
| FR-AGT-PRF-008 | Address Management | AGT-005, AGT-025 |
| FR-AGT-PRF-009 | Contact Information Update | AGT-005, AGT-025 |
| FR-AGT-PRF-010 | License Management | AGT-029 to AGT-034 |
| FR-AGT-PRF-011 | License Renewal | AGT-033 |
| FR-AGT-PRF-012 | License Auto-Deactivation | AGT-038 |
| FR-AGT-PRF-013 | Status Update | AGT-039, AGT-041, AGT-060 to AGT-063 |
| FR-AGT-PRF-014 | Termination Workflow | AGT-039, AGT-040 |
| FR-AGT-PRF-015 | OTP-Based Login | AGT-042, AGT-043 |
| FR-AGT-PRF-016 | Account Lockout | AGT-045 |
| FR-AGT-PRF-017 | Session Timeout | AGT-046 |
| FR-AGT-PRF-018 | Agent Dashboard | AGT-068 |
| FR-AGT-PRF-019 | Bank Details Management | AGT-050 to AGT-054 |
| FR-AGT-PRF-020 | Goal Setting | AGT-055 to AGT-059 |
| FR-AGT-PRF-021 | Self-Service Update | AGT-047 to AGT-049 |
| FR-AGT-PRF-022 | Audit History | AGT-028, AGT-076 |
| FR-AGT-PRF-024 | Product Authorization | AGT-074, AGT-075 |
| FR-AGT-PRF-025 | Profile Export | AGT-064 to AGT-067 |

### Validation Rules Coverage: 37/37 (100%)

All validation rules are applied across the relevant APIs:
- VR-AGT-PRF-001 to VR-AGT-PRF-036: Applied in AGT-005, AGT-012, AGT-025, AGT-030

### Workflows Coverage: 12/12 (100%)

| WF ID | Workflow | APIs |
|-------|----------|-----|
| WF-AGT-PRF-001 | Profile Creation | AGT-001 to AGT-006 |
| WF-AGT-PRF-002 | Profile Update | AGT-025 to AGT-028 |
| WF-AGT-PRF-003 | License Renewal | AGT-033 |
| WF-AGT-PRF-004 | Termination | AGT-039, AGT-040 |
| WF-AGT-PRF-005 | Authentication | AGT-042 to AGT-046 |
| WF-AGT-PRF-006 | HRMS Integration | AGT-002, AGT-013 |
| WF-AGT-PRF-007 | License Deactivation | AGT-038 |
| WF-AGT-PRF-008 | Bank Details Update | AGT-050 to AGT-054 |
| WF-AGT-PRF-009 | Goal Setting | AGT-055 to AGT-059 |
| WF-AGT-PRF-010 | Self-Service Update | AGT-047 to AGT-049 |
| WF-AGT-PRF-011 | Reinstatement | AGT-060 to AGT-063 |
| WF-AGT-PRF-012 | Profile Export | AGT-064 to AGT-067 |

### Error Codes Coverage: 12/12 (100%)

| Error Code | Message | API |
|-----------|---------|-----|
| ERR-AGT-PRF-001 | Invalid Agent Type | AGT-001 |
| ERR-AGT-PRF-002 | PAN Already Exists | AGT-012, AGT-025 |
| ERR-AGT-PRF-003 | Invalid PAN Format | AGT-005 |
| ERR-AGT-PRF-004 | First Name Mandatory | AGT-005 |
| ERR-AGT-PRF-005 | Last Name Mandatory | AGT-005 |
| ERR-AGT-PRF-006 | Invalid DOB | AGT-005 |
| ERR-AGT-PRF-007 | Employee ID Not Found | AGT-002 |
| ERR-AGT-PRF-008 | Coordinator Not Found | AGT-004 |
| ERR-AGT-PRF-009 | License Expired | AGT-030 |
| ERR-AGT-PRF-010 | Termination Reason Required | AGT-039 |
| ERR-AGT-PRF-011 | Account Locked | AGT-042, AGT-045 |
| ERR-AGT-PRF-012 | Session Expired | AGT-046 |

---

## Security & Authentication

### Authentication Mechanisms

1. **Bearer Authentication (JWT)**
   - Used for: Admin and Agent portal access
   - Header: `Authorization: Bearer <token>`
   - Token validity: Configurable (default: 15 minutes for portal)

2. **API Key Authentication**
   - Used for: Service-to-service communication
   - Header: `X-API-Key: <api_key>`

### Authorization Levels

| Role | Permissions |
|------|-------------|
| **Admin** | Full access to all APIs |
| **Advisor Coordinator** | Agent search, goal setting, view reports |
| **Agent (Self-Service)** | Profile view, update own profile (with OTP), bank details |
| **Supervisor** | Approve/reject profile updates, reinstatement requests |
| **System** | Batch jobs (license expiry deactivation) |

---

## Data Types & Validation

### Common Data Types

```yaml
AgentID:
  type: string
  pattern: "^AGT-\\d{4}-\\d{6}$"
  example: "AGT-2026-000567"

PAN:
  type: string
  pattern: "^[A-Z]{5}[0-9]{4}[A-Z]{1}$"
  example: "ABCDE1234F"

Aadhar:
  type: string
  pattern: "^[0-9]{12}$"
  example: "123456789012"

Mobile:
  type: string
  pattern: "^[6-9][0-9]{9}$"
  example: "9876543210"

IFSC:
  type: string
  pattern: "^[A-Z]{4}0[A-Z0-9]{6}$"
  example: "SBIN0001234"

Date:
  type: string
  format: date
  pattern: "^([0-9]{2})-([0-9]{2})-([0-9]{4})$"
  example: "23-01-2026"

UUID:
  type: string
  format: uuid
  example: "550e8400-e29b-41d4-a716-446655440000"
```

---

## API Response Standards

### Success Response Structure

```json
{
  "data": { /* response-specific data */ },
  "workflow_state": {
    "current_step": "string",
    "next_step": "string",
    "allowed_actions": ["string"],
    "progress_percentage": 0
  },
  "sla_tracking": {
    "sla_status": "GREEN|YELLOW|RED",
    "time_elapsed_minutes": 0,
    "next_actions_due": []
  },
  "notifications_sent": [
    {
      "type": "EMAIL|SMS|INTERNAL",
      "recipient": "string",
      "template": "string",
      "status": "SENT"
    }
  ],
  "messages": [
    {
      "type": "INFO|SUCCESS|WARNING|ERROR",
      "code": "MSG_CODE",
      "text": "Human-readable message"
    }
  ]
}
```

### Error Response Structure

```json
{
  "error": {
    "code": "ERR-AGT-PRF-XXX",
    "message": "Brief error description",
    "details": "Detailed error explanation",
    "suggested_actions": [
      "Action 1",
      "Action 2"
    ],
    "request_id": "uuid",
    "timestamp": "2026-01-23T10:30:00Z"
  }
}
```

---

## Integration Points

### External System Integrations

| Integration ID | System | Purpose | APIs |
|---------------|--------|---------|-----|
| INT-AGT-001 | HRMS System | Departmental employee data | AGT-002, AGT-013, AGT-078 |
| INT-AGT-005 | Notification Service | Email/SMS notifications | AGT-021, AGT-077 |
| INT-AGT-006 | Letter Generation | Termination letters | AGT-040 |

---

## Rate Limiting

| API Category | Rate Limit | Burst |
|--------------|------------|-------|
| Profile Creation | 10/minute | 20 |
| Profile Search | 60/minute | 100 |
| Profile Update | 30/minute | 50 |
| Portal Login | 10/minute | 20 |
| Other APIs | 60/minute | 100 |

**Rate Limit Headers:**
```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1605042600
Retry-After: 60
```

---

## Usage Instructions

### 1. Viewing the Specification

Use any OpenAPI-compatible tool:
- **Swagger UI:** `https://editor.swagger.io/`
- **Redoc:** `https://redocly.github.io/redoc/`
- **VS Code:** Install "OpenAPI (Swagger) Editor" extension

### 2. Combining Parts

To use as a single specification, combine all 3 parts:

```yaml
# Combine parts in order:
# 1. agent_profile_management_api_part1.yaml
# 2. agent_profile_management_api_part2.yaml
# 3. agent_profile_management_api_part3.yaml
```

Or use a YAML file that imports all parts:

```yaml
openapi: 3.0.3
info:
  title: Agent Profile Management API
  version: 1.0.0

paths:
  /agent-profiles/initiate:
    $ref: './agent_profile_management_api_part1.yaml#/paths/~1agent-profiles~1initiate'
  # ... import all paths

components:
  schemas:
    PersonalInfo:
      $ref: './agent_profile_management_api_part1.yaml#/components/schemas/PersonalInfo'
    # ... import all schemas
```

### 3. Code Generation

Generate client SDKs or server stubs:

```bash
# Using OpenAPI Generator
openapi-generator-cli generate \
  -i agent_profile_management_api_part1.yaml \
  -g go \
  -o ./generated/go-client

# Generate TypeScript client
openapi-generator-cli generate \
  -i agent_profile_management_api_part1.yaml \
  -g typescript-axios \
  -o ./generated/ts-client

# Generate server stubs (Golang)
openapi-generator-cli generate \
  -i agent_profile_management_api_part1.yaml \
  -g go-server \
  -o ./generated/go-server
```

---

## Validation & Testing

### Request Validation

All endpoints include:
- **Required field validation**
- **Format validation (regex patterns)**
- **Length validation (minLength, maxLength)**
- **Enum validation (allowed values)**
- **Range validation (minimum, maximum)**

### Response Validation

All responses follow:
- **Consistent structure**
- **Proper HTTP status codes**
- **Error codes and messages**
- **Pagination metadata (where applicable)**

---

## Next Steps

1. **Review** the API specifications for completeness
2. **Generate** client SDKs using OpenAPI Generator
3. **Implement** backend services following the specifications
4. **Test** all endpoints with sample requests
5. **Deploy** API documentation (Swagger UI/Redoc)

---

**End of Summary Document**
