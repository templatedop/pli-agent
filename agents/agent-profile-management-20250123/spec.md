# Agent Profile Management - Technical Specification

## Document Control

| Attribute | Details |
|-----------|---------|
| **Module** | Agent Profile Management |
| **Analysis Date** | 2026-01-23 |
| **Complexity** | Medium |
| **Technology Stack** | Golang, Temporal.io, PostgreSQL, Kafka, React |

---

## Complexity Assessment

| Aspect | Assessment |
|--------|------------|
| **Overall Complexity** | Medium |
| **Estimated Journeys** | 8-12 |
| **Integration Complexity** | High (8 external systems) |
| **Temporal Workflows** | 5-7 workflows |
| **API Count Estimate** | 60-80 unique APIs |

---

## Component Catalog

| Component Type | Count | ID Range |
|----------------|-------|----------|
| **Business Rules** | 38 | BR-AGT-PRF-001 to BR-AGT-PRF-037 |
| **Functional Requirements** | 27 | FR-AGT-PRF-001 to FR-AGT-PRF-027 |
| **Validation Rules** | 37 | VR-AGT-PRF-001 to VR-AGT-PRF-036 |
| **Workflows** | 12 | WF-AGT-PRF-001 to WF-AGT-PRF-012 |
| **Error Codes** | 12 | ERR-AGT-PRF-001 to ERR-AGT-PRF-012 |
| **Integration Points** | 8 | INT-AGT-001 to INT-AGT-008 |
| **Data Entities** | 8 | E-AGT-PRF-001 to E-AGT-PRF-008 |

---

## Key Business Rules Summary

### CRITICAL Priority Rules
- **BR-AGT-PRF-001**: Advisor Coordinator Linkage Requirement
- **BR-AGT-PRF-002**: Advisor Coordinator Geographic Assignment
- **BR-AGT-PRF-012**: License Renewal Period Rules (Complex 3-year provisional to permanent transition)
- **BR-AGT-PRF-013**: Auto-Deactivation on License Expiry
- **BR-AGT-PRF-017**: Agent Termination Workflow
- **BR-AGT-PRF-031**: Agent Profile Creation Workflow Orchestration

### HIGH Priority Rules
- **BR-AGT-PRF-003**: Departmental Employee HRMS Integration
- **BR-AGT-PRF-005**: Name Update with Audit Logging
- **BR-AGT-PRF-006**: PAN Update with Format and Uniqueness Validation
- **BR-AGT-PRF-014**: License Renewal Reminder Schedule
- **BR-AGT-PRF-016**: Status Update with Mandatory Reason
- **BR-AGT-PRF-018**: Bank/POSB Account Details for Commission
- **BR-AGT-PRF-019**: OTP-Based Two-Factor Authentication
- **BR-AGT-PRF-020**: Account Lockout After Failed Attempts
- **BR-AGT-PRF-022**: Multi-Criteria Agent Search
- **BR-AGT-PRF-032**: Employee ID Validation for Departmental Employees
- **BR-AGT-PRF-033**: Person Type Selection Requirement
- **BR-AGT-PRF-034**: Office Type and Office Code Association

---

## Agent Types Supported

1. **Advisor** - Front-line insurance sales agents (requires coordinator linkage)
2. **Advisor Coordinator** - Managers who oversee Advisors
3. **Departmental Employee** - Internal postal employees acting as agents (HRMS integrated)
4. **Field Officer** - External field staff with agent capabilities (optional HRMS)
5. **Direct Agent** - Agents without coordinator
6. **GDS** - Gramin Dak Sevak

---

## Status Types

| Status | Definition | Can Sell | Portal Access | Commission |
|--------|------------|----------|---------------|------------|
| **Active** | Fully active with valid license | Yes | Yes | Yes |
| **Suspended** | Temporarily suspended | No | Read-only | No |
| **Terminated** | Relationship ended | No | No | No (to date) |
| **Deactivated** | License expired | No | No | No |
| **Expired** | In grace period | No | Read-only | No |

---

## Integration Points

| System | Purpose | Data Exchange | Frequency |
|--------|---------|---------------|-----------|
| **HRMS System** | Employee data auto-population | Profile data | Real-time |
| **KYC/BCP Services** | Document verification | KYC status | Real-time |
| **Commission Processing** | Commission disbursement | Status, bank details | Monthly |
| **Policy Services** | Agent validation | Status, license | Real-time |
| **Portal Services** | Agent self-service | Profile data | Real-time |
| **Notification Service** | Reminders & alerts | Contact data | Scheduled |
| **Letter Generation** | Letters generation | Profile data | On-demand |
| **Payment Gateway** | Commission EFT | Bank details | Monthly |
| **Authentication Service** | OTP login | Credentials | Real-time |

---

## Workflow Execution Notes

This specification will guide the creation of:
1. **8-12 User Journeys** covering complete agent lifecycle
2. **60-80 APIs** including hidden/supporting APIs
3. **5-7 Temporal Workflows** for long-running processes
4. **100% Component Traceability** matrix
5. **Enhanced Response Schemas** with workflow_state and SLA tracking

---

**Next Steps**: Proceed to Phase 2 - User Journey Creation
