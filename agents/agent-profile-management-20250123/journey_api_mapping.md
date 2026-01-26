# Agent Profile Management - Journey to API Mapping Catalog

## Document Control

| Attribute | Details |
|-----------|---------|
| **Module** | Agent Profile Management |
| **Version** | 1.0 |
| **Created** | 2026-01-23 |
| **Total Journeys** | 10 |
| **Total APIs** | 78 |

---

## API Mapping Table

| API ID | Endpoint | Method | Journey(s) | Components | Type | Priority |
|--------|----------|--------|------------|------------|------|----------|
| AGT-001 | /api/v1/agent-profiles/initiate | POST | UJ-001 | FR-001, BR-031, VR-025 | Core | CRITICAL |
| AGT-002 | /api/v1/agent-profiles/{session_id}/fetch-hrms | POST | UJ-001 | FR-002, BR-003, BR-032, VR-023, INT-001 | Core | CRITICAL |
| AGT-003 | /api/v1/advisor-coordinators | GET | UJ-001, UJ-006 | FR-003, BR-001, BR-002 | Lookup | HIGH |
| AGT-004 | /api/v1/agent-profiles/{session_id}/link-coordinator | POST | UJ-001 | FR-003, BR-001, BR-002 | Core | CRITICAL |
| AGT-005 | /api/v1/agent-profiles/{session_id}/validate-basic | POST | UJ-001 | FR-002, BR-005-011, VR-002-012 | Core | CRITICAL |
| AGT-006 | /api/v1/agent-profiles/{session_id}/submit | POST | UJ-001 | FR-001, BR-031, E-001-008 | Core | CRITICAL |
| AGT-007 | /api/v1/agent-types | GET | UJ-001, UJ-004 | | Lookup | MEDIUM |
| AGT-008 | /api/v1/categories | GET | UJ-001, UJ-002 | | Lookup | LOW |
| AGT-009 | /api/v1/designations | GET | UJ-001, UJ-002 | | Lookup | LOW |
| AGT-010 | /api/v1/office-types | GET | UJ-001 | BR-034 | Lookup | MEDIUM |
| AGT-011 | /api/v1/states | GET | UJ-001, UJ-002 | | Lookup | LOW |
| AGT-012 | /api/v1/validations/pan/check-uniqueness | POST | UJ-001, UJ-002 | BR-006, VR-002 | Validation | HIGH |
| AGT-013 | /api/v1/validations/hrms/employee-id | POST | UJ-001 | BR-003, BR-032, VR-023 | Validation | HIGH |
| AGT-014 | /api/v1/validations/bank/ifsc | POST | UJ-001, UJ-007 | BR-018, VR-017 | Validation | HIGH |
| AGT-015 | /api/v1/validations/office/{office_code} | GET | UJ-001 | BR-034, VR-027 | Validation | HIGH |
| AGT-016 | /api/v1/agent-profiles/sessions/{session_id}/status | GET | UJ-001 | | Workflow | MEDIUM |
| AGT-017 | /api/v1/agent-profiles/sessions/{session_id}/save | POST | UJ-001 | | Workflow | MEDIUM |
| AGT-018 | /api/v1/agent-profiles/sessions/{session_id}/resume | GET | UJ-001 | | Workflow | MEDIUM |
| AGT-019 | /api/v1/agent-profiles/sessions/{session_id} | DELETE | UJ-001 | | Workflow | MEDIUM |
| AGT-020 | /api/v1/agent-profiles/creation-status/{agent_id} | GET | UJ-001 | | Status | MEDIUM |
| AGT-021 | /api/v1/agents/{agent_id}/notifications/resend-welcome | POST | UJ-001 | INT-005 | Notification | LOW |
| AGT-022 | /api/v1/agents/search | GET | UJ-002, UJ-010 | FR-004, BR-022, VR-018-019 | Core | HIGH |
| AGT-023 | /api/v1/agents/{agent_id} | GET | UJ-002, UJ-003, UJ-006 | FR-005, BR-023 | Core | HIGH |
| AGT-024 | /api/v1/agents/{agent_id}/update-form | GET | UJ-002 | FR-006, FR-007, FR-008, FR-009 | Core | HIGH |
| AGT-025 | /api/v1/agents/{agent_id}/sections/{section} | PUT | UJ-002 | FR-006, FR-007, BR-005, BR-006, E-008 | Core | HIGH |
| AGT-026 | /api/v1/approvals/{approval_request_id}/approve | PUT | UJ-002 | BR-005, BR-006 | Approval | HIGH |
| AGT-027 | /api/v1/approvals/{approval_request_id}/reject | PUT | UJ-002 | | Approval | HIGH |
| AGT-028 | /api/v1/agents/{agent_id}/audit-history | GET | UJ-002 | FR-022, E-008 | Status | MEDIUM |
| AGT-029 | /api/v1/agents/{agent_id}/licenses | GET | UJ-003 | FR-010, BR-012, BR-014, E-006, E-007 | Core | CRITICAL |
| AGT-030 | /api/v1/agents/{agent_id}/licenses | POST | UJ-003 | FR-010, BR-012, VR-031-036 | Core | CRITICAL |
| AGT-031 | /api/v1/agents/{agent_id}/licenses/{license_id} | GET | UJ-003 | | Core | HIGH |
| AGT-032 | /api/v1/agents/{agent_id}/licenses/{license_id} | PUT | UJ-003 | FR-010, BR-012 | Core | HIGH |
| AGT-033 | /api/v1/agents/{agent_id}/licenses/{license_id}/renew | PUT | UJ-003 | FR-011, BR-012 | Core | CRITICAL |
| AGT-034 | /api/v1/agents/{agent_id}/licenses/{license_id} | DELETE | UJ-003 | | Core | MEDIUM |
| AGT-035 | /api/v1/license-types | GET | UJ-003 | | Lookup | MEDIUM |
| AGT-036 | /api/v1/licenses/expiring | GET | UJ-003 | BR-014 | Status | HIGH |
| AGT-037 | /api/v1/licenses/{license_id}/reminders | GET | UJ-003 | BR-014 | Status | MEDIUM |
| AGT-038 | /api/v1/licenses/expired | POST | UJ-003 | BR-013, FR-012 | System | CRITICAL |
| AGT-039 | /api/v1/agents/{agent_id}/terminate | POST | UJ-004 | FR-014, BR-017, VR-020-021 | Core | CRITICAL |
| AGT-040 | /api/v1/agents/{agent_id}/termination-letter | GET | UJ-004 | BR-017, INT-006 | Document | HIGH |
| AGT-041 | /api/v1/agents/{agent_id}/reinstate | POST | UJ-009 | FR-013, BR-016 | Core | HIGH |
| AGT-042 | /api/v1/agents/{agent_id}/portal/login | POST | UJ-005 | FR-015, BR-019, BR-020 | Core | CRITICAL |
| AGT-043 | /api/v1/agents/{agent_id}/portal/otp/verify | POST | UJ-005 | FR-015, BR-019, BR-020 | Core | CRITICAL |
| AGT-044 | /api/v1/agents/{agent_id}/portal/logout | POST | UJ-005 | FR-017, BR-021 | Core | HIGH |
| AGT-045 | /api/v1/agents/{agent_id}/portal/unlock | POST | UJ-005 | BR-020, ERR-011 | Core | HIGH |
| AGT-046 | /api/v1/agents/portal/session/status | GET | UJ-005 | BR-021, ERR-012 | Status | MEDIUM |
| AGT-047 | /api/v1/agents/{agent_id}/portal/profile | GET | UJ-006 | FR-018, FR-021 | Core | HIGH |
| AGT-048 | /api/v1/agents/{agent_id}/portal/profile | PUT | UJ-006 | FR-021, BR-005-011 | Core | HIGH |
| AGT-049 | /api/v1/agents/{agent_id}/portal/profile/otp | POST | UJ-006 | BR-019, BR-020 | Validation | HIGH |
| AGT-050 | /api/v1/agents/{agent_id}/portal/bank-details | GET | UJ-007 | FR-019, BR-018 | Core | CRITICAL |
| AGT-051 | /api/v1/agents/{agent_id}/portal/bank-details | POST | UJ-007 | FR-019, BR-018, VR-016-017 | Core | CRITICAL |
| AGT-052 | /api/v1/agents/{agent_id}/bank-details | GET | UJ-007 | FR-019, BR-018 | Core | HIGH |
| AGT-053 | /api/v1/agents/{agent_id}/bank-details | POST | UJ-007 | FR-019, BR-018, VR-016-017 | Core | HIGH |
| AGT-054 | /api/v1/agents/{agent_id}/bank-details | PUT | UJ-007 | FR-019, BR-018 | Core | HIGH |
| AGT-055 | /api/v1/agents/{agent_id}/goals | GET | UJ-008 | FR-020, BR-024 | Core | MEDIUM |
| AGT-056 | /api/v1/agents/{agent_id}/goals | POST | UJ-008 | FR-020, BR-024 | Core | MEDIUM |
| AGT-057 | /api/v1/agents/{agent_id}/goals/{goal_id} | PUT | UJ-008 | FR-020 | Core | MEDIUM |
| AGT-058 | /api/v1/agents/{agent_id}/goals/{goal_id} | DELETE | UJ-008 | FR-020 | Core | LOW |
| AGT-059 | /api/v1/agents/{agent_id}/goals/progress | GET | UJ-008 | FR-018, FR-020 | Status | MEDIUM |
| AGT-060 | /api/v1/reinstatement/request | POST | UJ-009 | FR-013, BR-016 | Core | HIGH |
| AGT-061 | /api/v1/reinstatement/{request_id}/approve | PUT | UJ-009 | | Approval | HIGH |
| AGT-062 | /api/v1/reinstatement/{request_id}/reject | PUT | UJ-009 | | Approval | HIGH |
| AGT-063 | /api/v1/reinstatement/{request_id}/documents | POST | UJ-009 | | Document | HIGH |
| AGT-064 | /api/v1/agents/export/configure | POST | UJ-010 | FR-025 | Core | MEDIUM |
| AGT-065 | /api/v1/agents/export/execute | POST | UJ-010 | FR-025, E-008 | Core | MEDIUM |
| AGT-066 | /api/v1/agents/export/{export_id}/status | GET | UJ-010 | | Status | MEDIUM |
| AGT-067 | /api/v1/agents/export/{export_id}/download | GET | UJ-010 | | Document | MEDIUM |
| AGT-068 | /api/v1/dashboard/agent/{agent_id} | GET | UJ-006 | FR-018, FR-021 | Dashboard | MEDIUM |
| AGT-069 | /api/v1/goals/templates | GET | UJ-008 | FR-020 | Lookup | LOW |
| AGT-070 | /api/v1/status-types | GET | UJ-004, UJ-009 | | Lookup | LOW |
| AGT-071 | /api/v1/reinstatement/reasons | GET | UJ-009 | | Lookup | LOW |
| AGT-072 | /api/v1/termination/reasons | GET | UJ-004 | BR-017, VR-020 | Lookup | MEDIUM |
| AGT-073 | /api/v1/agents/{agent_id}/hierarchy | GET | UJ-002 | BR-001, BR-002 | Lookup | MEDIUM |
| AGT-074 | /api/v1/agents/{agent_id}/product-authorization | GET | UJ-002 | FR-024, BR-026 | Status | MEDIUM |
| AGT-075 | /api/v1/agents/{agent_id}/product-authorization | POST | UJ-002 | FR-024, BR-026 | Core | MEDIUM |
| AGT-076 | /api/v1/agents/{agent_id}/timeline | GET | UJ-002 | E-008 | Status | MEDIUM |
| AGT-077 | /api/v1/agents/{agent_id}/notifications | GET | UJ-002, UJ-006 | INT-005 | Status | LOW |
| AGT-078 | /api/v1/webhooks/hrms/employee-update | POST | UJ-001 | INT-001 | Webhook | MEDIUM |

---

## API Category Summary

| Category | Count | Examples |
|----------|-------|----------|
| **Core** | 38 | Profile creation, updates, licenses, termination |
| **Lookup** | 14 | Agent types, states, designations, office types |
| **Validation** | 4 | PAN uniqueness, HRMS employee ID, IFSC, office code |
| **Status** | 10 | Agent status, license expiry, audit history, progress |
| **Workflow** | 4 | Session management, save, resume, cancel |
| **Approval** | 3 | Approve/reject profile updates, reinstatement |
| **Notification** | 2 | Welcome notification, notification history |
| **Document** | 3 | Termination letter, export, reinstatement documents |
| **Dashboard** | 1 | Agent dashboard |
| **System** | 1 | License deactivation batch |
| **Webhook** | 1 | HRMS employee update |

**Total**: 78 unique APIs

---

## Journey to API Count Summary

| Journey ID | Journey Name | Total APIs | Core | Lookup | Validation | Other |
|------------|--------------|------------|------|--------|------------|-------|
| UJ-001 | Agent Profile Creation | 21 | 6 | 5 | 4 | 6 |
| UJ-002 | Agent Profile Update | 7 | 3 | 0 | 0 | 4 |
| UJ-003 | License Management | 10 | 6 | 1 | 0 | 3 |
| UJ-004 | Agent Termination | 3 | 1 | 1 | 0 | 1 |
| UJ-005 | Portal Authentication | 5 | 4 | 0 | 0 | 1 |
| UJ-006 | Self-Service Update | 3 | 2 | 0 | 1 | 0 |
| UJ-007 | Bank Details Management | 5 | 4 | 0 | 0 | 1 |
| UJ-008 | Agent Goal Setting | 5 | 4 | 1 | 0 | 0 |
| UJ-009 | Status Reinstatement | 4 | 1 | 2 | 0 | 1 |
| UJ-010 | Profile Search & Export | 4 | 2 | 0 | 0 | 2 |

---

## Component Traceability Summary

### Business Rules Applied
| BR ID | Applied In Journey(s) | API Count |
|-------|----------------------|-----------|
| BR-AGT-PRF-001 | UJ-001 | 2 APIs |
| BR-AGT-PRF-002 | UJ-001 | 2 APIs |
| BR-AGT-PRF-003 | UJ-001 | 2 APIs |
| BR-AGT-PRF-005 | UJ-002 | 2 APIs |
| BR-AGT-PRF-006 | UJ-002 | 2 APIs |
| BR-AGT-PRF-012 | UJ-003 | 5 APIs |
| BR-AGT-PRF-013 | UJ-003 | 1 API |
| BR-AGT-PRF-014 | UJ-003 | 2 APIs |
| BR-AGT-PRF-017 | UJ-004 | 1 API |
| BR-AGT-PRF-019 | UJ-005 | 2 APIs |
| BR-AGT-PRF-020 | UJ-005 | 1 API |
| BR-AGT-PRF-022 | UJ-002, UJ-010 | 1 API |
| BR-AGT-PRF-024 | UJ-008 | 4 APIs |

### Functional Requirements Covered
| FR ID | Journey(s) | API Count |
|-------|-----------|-----------|
| FR-AGT-PRF-001 | UJ-001 | 2 APIs |
| FR-AGT-PRF-002 | UJ-001 | 4 APIs |
| FR-AGT-PRF-003 | UJ-001 | 2 APIs |
| FR-AGT-PRF-004 | UJ-002, UJ-010 | 1 API |
| FR-AGT-PRF-005 | UJ-002 | 1 API |
| FR-AGT-PRF-006 | UJ-002 | 2 APIs |
| FR-AGT-PRF-007 | UJ-002 | 2 APIs |
| FR-AGT-PRF-010 | UJ-003 | 5 APIs |
| FR-AGT-PRF-011 | UJ-003 | 1 API |
| FR-AGT-PRF-014 | UJ-004 | 1 API |
| FR-AGT-PRF-015 | UJ-005 | 2 APIs |
| FR-AGT-PRF-018 | UJ-006, UJ-008 | 2 APIs |
| FR-AGT-PRF-019 | UJ-007 | 5 APIs |
| FR-AGT-PRF-020 | UJ-008 | 4 APIs |
| FR-AGT-PRF-021 | UJ-006 | 2 APIs |
| FR-AGT-PRF-022 | UJ-002 | 1 API |
| FR-AGT-PRF-024 | UJ-002 | 2 APIs |
| FR-AGT-PRF-025 | UJ-010 | 3 APIs |

### Integration Points Utilized
| INT ID | Integration | Journey(s) | API Count |
|--------|-------------|-----------|-----------|
| INT-AGT-001 | HRMS System | UJ-001 | 2 APIs |
| INT-AGT-005 | Notification Service | UJ-001, UJ-002 | 2 APIs |
| INT-AGT-006 | Letter Generation | UJ-004 | 1 API |

---

**End of Document**
