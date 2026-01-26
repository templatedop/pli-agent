# Phase 7 - License Management Implementation (AGT-029 to AGT-038)

**Date**: 2026-01-26
**Branch**: claude/phase-7-license-management-[session-id]
**Status**: 🚧 IN PROGRESS

---

## 📋 Overview

Implementing 10 License Management APIs covering the complete license lifecycle from issuance to renewal and expiry management.

### User Journey
**UJ-003**: License Management & Renewal

### APIs to Implement (10 endpoints)

| API ID | Endpoint | Method | Description | Priority |
|--------|----------|--------|-------------|----------|
| AGT-029 | `/agents/{agent_id}/licenses` | GET | Get all licenses for agent | CRITICAL |
| AGT-030 | `/agents/{agent_id}/licenses` | POST | Add new license | CRITICAL |
| AGT-031 | `/agents/{agent_id}/licenses/{license_id}` | GET | Get license details | HIGH |
| AGT-032 | `/agents/{agent_id}/licenses/{license_id}` | PUT | Update license | HIGH |
| AGT-033 | `/agents/{agent_id}/licenses/{license_id}/renew` | PUT | Renew license | CRITICAL |
| AGT-034 | `/agents/{agent_id}/licenses/{license_id}` | DELETE | Delete license | MEDIUM |
| AGT-035 | `/license-types` | GET | Get license types (lookup) | MEDIUM |
| AGT-036 | `/licenses/expiring` | GET | Get expiring licenses | HIGH |
| AGT-037 | `/licenses/{license_id}/reminders` | GET | Get reminder schedule | MEDIUM |
| AGT-038 | `/licenses/expired` | POST | Trigger expiry deactivation | CRITICAL |

---

## 🎯 Business Rules

### BR-AGT-PRF-012: License Renewal Period Rules
- **Provisional License**: 1 year renewal period
- **Permanent License (after exam)**: 5 years renewal period
- **Maximum Provisional Renewals**: 2 times before exam required
- **Annual Renewal**: Required for permanent licenses

### BR-AGT-PRF-013: Auto-Deactivation on Expiry
- Agents with expired licenses are automatically deactivated
- Portal access disabled
- Commission processing stopped

### BR-AGT-PRF-014: License Renewal Reminders
- **30 days before expiry**
- **15 days before expiry**
- **7 days before expiry**
- **On expiry day**

---

## 🗄️ Database Schema

### Tables

#### agent_licenses (E-06)
```sql
- license_id (UUID, PK)
- agent_id (UUID, FK → agent_profiles)
- license_line (ENUM: LIFE, GENERAL)
- license_type (ENUM: PROVISIONAL, PERMANENT)
- license_number (VARCHAR, UNIQUE)
- resident_status (ENUM: RESIDENT, NON_RESIDENT)
- license_date, renewal_date, authority_date (DATE)
- renewal_count (INTEGER)
- license_status (ENUM: ACTIVE, EXPIRED, RENEWED)
- licentiate_exam_passed (BOOLEAN)
- licentiate_exam_date, licenticate_certificate_number
- is_primary (BOOLEAN)
- audit fields (created_at, updated_at, created_by, updated_by, deleted_at, version)
```

#### agent_license_reminders (E-07)
```sql
- reminder_id (UUID, PK)
- license_id (UUID, FK → agent_licenses)
- reminder_type (ENUM: 30_DAYS, 15_DAYS, 7_DAYS, EXPIRY_DAY)
- reminder_date, sent_date (DATE/TIMESTAMP)
- sent_status (ENUM: PENDING, SENT, FAILED)
- email_sent, sms_sent (BOOLEAN)
- failure_reason, retry_count
- audit fields
```

---

## 🏗️ Implementation Plan

### ✅ Phase 1: Domain Models (COMPLETED)
- [x] `core/domain/agent_license.go` (139 lines)
- [x] `core/domain/license_reminder.go` (75 lines)

### 🚧 Phase 2: Repository Layer (IN PROGRESS)
- [ ] `repo/postgres/agent_license.go`
  - [ ] Create() - Add new license with automatic renewal date calculation
  - [ ] FindByID() - Get specific license
  - [ ] FindByAgentID() - Get all licenses for an agent
  - [ ] Update() - Update license details
  - [ ] Renew() - Renew license with period calculation and status update
  - [ ] Delete() - Soft delete license
  - [ ] FindExpiring() - Get licenses expiring within N days
  - [ ] DeactivateExpiredAgents() - Batch job to deactivate expired licenses

- [ ] `repo/postgres/license_reminder.go`
  - [ ] Create() - Create reminder
  - [ ] FindByLicenseID() - Get all reminders for a license
  - [ ] UpdateSentStatus() - Mark reminder as sent/failed
  - [ ] FindPendingReminders() - Get pending reminders to send

### 🚧 Phase 3: Handler Layer
- [ ] `handler/license_management.go`
  - [ ] GetAgentLicenses() - AGT-029
  - [ ] AddLicense() - AGT-030
  - [ ] GetLicenseDetails() - AGT-031
  - [ ] UpdateLicense() - AGT-032
  - [ ] RenewLicense() - AGT-033 (COMPLEX)
  - [ ] DeleteLicense() - AGT-034
  - [ ] GetLicenseTypes() - AGT-035
  - [ ] GetExpiringLicenses() - AGT-036
  - [ ] GetLicenseReminders() - AGT-037
  - [ ] TriggerExpiryDeactivation() - AGT-038 (Batch job)

### 🚧 Phase 4: DTOs
- [ ] `handler/request.go` (additions)
  - [ ] AddLicenseRequest
  - [ ] UpdateLicenseRequest
  - [ ] RenewLicenseRequest
  - [ ] GetExpiringLicensesQuery
  - [ ] TriggerExpiryRequest

- [ ] `handler/response/license_management.go`
  - [ ] AgentLicensesResponse
  - [ ] LicenseDetailsResponse
  - [ ] AddLicenseResponse
  - [ ] RenewLicenseResponse
  - [ ] ExpiringLicensesResponse
  - [ ] LicenseRemindersResponse
  - [ ] ExpiryDeactivationResponse

### 🚧 Phase 5: Bootstrap Registration
- [ ] `bootstrap/bootstrapper.go`
  - [ ] Register AgentLicenseRepository in FxRepo
  - [ ] Register LicenseReminderRepository in FxRepo
  - [ ] Register LicenseManagementHandler in FxHandler

---

## 🔑 Key Implementation Details

### Renewal Date Calculation (BR-AGT-PRF-012)
```go
func CalculateRenewalDate(licenseType string, licenseDate time.Time, examPassed bool) time.Time {
    if licenseType == "PROVISIONAL" {
        return licenseDate.AddDate(1, 0, 0) // +1 year
    }
    // Permanent license after exam
    if examPassed {
        return licenseDate.AddDate(5, 0, 0) // +5 years
    }
    return licenseDate.AddDate(1, 0, 0) // Annual renewal
}
```

### Reminder Scheduling (BR-AGT-PRF-014)
```go
reminders := []ReminderType{
    {Type: "30_DAYS", Date: renewalDate.AddDate(0, 0, -30)},
    {Type: "15_DAYS", Date: renewalDate.AddDate(0, 0, -15)},
    {Type: "7_DAYS", Date: renewalDate.AddDate(0, 0, -7)},
    {Type: "EXPIRY_DAY", Date: renewalDate},
}
```

### Expiry Deactivation Workflow (AGT-038)
1. Find all licenses with `renewal_date < batch_date` AND `status = ACTIVE`
2. Update agent status to `DEACTIVATED`
3. Disable portal access
4. Update license status to `EXPIRED`
5. Create audit logs
6. Send notifications (email/SMS)
7. Return batch results

---

## 🎓 Framework Patterns to Follow

### Single Database Round Trip
- Use `pgx.Batch` for all operations
- Combine count + data queries using CTE
- Use `UPDATE...RETURNING` for atomic operations

### Type Safety
- Use Squirrel query builder
- Use `pgx.RowToStructByNameLax` for row mapping
- Use `dblib.SelectOne`, `dblib.SelectRows`, `dblib.QueueReturn`

### Error Handling
- Consistent error wrapping with `fmt.Errorf`
- Proper `pgx.ErrNoRows` handling
- Clear error messages

### Audit Trail
- All create/update/delete operations must create audit logs
- Use `action_type` enum: `LICENSE_ADD`, `LICENSE_UPDATE`, `LICENSE_RENEW`, `LICENSE_DELETE`

---

## 📝 Testing Checklist

- [ ] AGT-029: Search licenses with status filter
- [ ] AGT-030: Add provisional license → verify renewal date = +1 year
- [ ] AGT-030: Add permanent license with exam → verify renewal date = +5 years
- [ ] AGT-033: Renew provisional (1st time) → verify renewal_count = 1
- [ ] AGT-033: Renew provisional (3rd time) → should fail (max 2 renewals)
- [ ] AGT-033: Renew provisional with exam passed → convert to PERMANENT
- [ ] AGT-036: Get expiring licenses within 30 days
- [ ] AGT-037: Verify 4 reminders scheduled (30, 15, 7 days, expiry)
- [ ] AGT-038: Batch deactivation → verify agent status = DEACTIVATED

---

## 📦 Expected Code Metrics

| Component | Files | Est. Lines | Description |
|-----------|-------|------------|-------------|
| Domain | 2 | 214 | AgentLicense, LicenseReminder |
| Repository | 2 | ~800 | License + Reminder repos |
| Handler | 1 | ~700 | 10 endpoint handlers |
| Request DTOs | 1 | ~100 | 6 request structures |
| Response DTOs | 1 | ~150 | 7 response structures |
| Bootstrap | 1 | ~5 | FX registrations |
| **TOTAL** | 8 | ~1,969 | Production code |

---

**Session**: https://claude.ai/code/session_01TgC4gchEZCp8AAzuEpQ6NR
