# Old Phase 7 Code - Pre-Replacement Version

This folder contains the original implementation of license-related files **BEFORE Phase 7 implementation** (AGT-029 to AGT-038).

## Source Commit
- **Commit**: `47f3d4a` - "feat: Complete address and contact update implementation in CreateApprovalRequestWithChanges"
- **Date**: Just before Phase 7 started (commit ea894a1)
- **Purpose**: Preserve original business logic, audit patterns, and validation methods for analysis and potential restoration

## Files Preserved

### 1. `core/domain/agent_license.go` (99 lines)
**Original domain model** with:
- Basic AgentLicense entity structure
- License type constants (PROVISIONAL, PERMANENT)
- License status constants (ACTIVE, EXPIRED, SUSPENDED)
- Basic helper methods

**What was in original vs new:**
- Original: Simpler domain model
- New (Phase 7): Enhanced with business rule constants, expiry status calculation, renewal period rules

### 2. `core/domain/agent_license_reminder.go` (55 lines)
**Original reminder entity** - This file was NOT modified in Phase 7
- Still exists in current codebase unchanged
- Included here for complete context

### 3. `repo/postgres/agent_license.go` (487 lines) ⚠️ CRITICAL
**Original repository implementation** with comprehensive business logic:

#### Methods in Original (13 total):
1. ✅ `Create()` - INSERT license + INSERT audit (CTE pattern)
2. ✅ `FindByID()` - Get license by ID
3. ✅ `FindByAgentID()` - Get all agent licenses
4. ✅ `FindPrimaryLicense()` - Get primary license ⚠️ MISSING in new
5. ✅ `FindByLicenseNumber()` - Lookup by license number ⚠️ MISSING in new
6. ✅ `Update()` - UPDATE license + INSERT audit (CTE pattern)
7. ✅ `RenewLicense()` - Renewal logic + INSERT audit (CTE pattern)
8. ✅ `FindExpiringLicenses()` - Get licenses expiring in N days
9. ✅ `FindExpiredLicenses()` - Get already expired licenses ⚠️ MISSING in new
10. ✅ `MarkAsExpired()` - Mark single license expired + audit ⚠️ MISSING in new
11. ✅ `BatchMarkAsExpired()` - Batch expiry operation + audit ⚠️ MISSING in new
12. ✅ `Delete()` - Soft delete + INSERT audit (batch pattern)
13. ✅ `ValidateLicenseNumberUniqueness()` - VR-AGT-PRF-020 ⚠️ MISSING in new

#### Key Features in Original Implementation:

**A. Audit Logging (CRITICAL - Missing in Phase 7)**
- ✅ All CRUD operations include audit log insertion
- ✅ Used CTE pattern: `WITH inserted AS (...) INSERT INTO audit_logs SELECT ... FROM inserted`
- ✅ Implements BR-AGT-PRF-005 (Audit Logging) and FR-AGT-PRF-022 (Audit History Tracking)

**B. CTE Pattern for Single Database Round Trip**
```go
// Original pattern from commit 2492ddc
WITH inserted AS (
    INSERT INTO agent_licenses (...) VALUES (...) RETURNING *
)
INSERT INTO agent_audit_logs (agent_id, action_type, field_name, new_value, ...)
SELECT agent_id, ..., FROM inserted
RETURNING (SELECT ROW(license_id, ...) FROM inserted)
```

**C. Batch Operations**
- ✅ `BatchMarkAsExpired()` - Process multiple licenses in single batch
- ✅ Used pgx.Batch to combine multiple operations

**D. Validation Methods**
- ✅ `ValidateLicenseNumberUniqueness()` - Essential for VR-AGT-PRF-020

## What Changed in Phase 7 (New Implementation)

### New Repository (current): `repo/postgres/agent_license.go` (396 lines)

#### Methods in New (9 total):
1. ⚠️ `Create()` - INSERT license (NO AUDIT) ❌
2. ✅ `FindByID()` - Same as original
3. ⚠️ `FindByAgentID()` - Enhanced with status filter parameter
4. ❌ `FindPrimaryLicense()` - REMOVED
5. ❌ `FindByLicenseNumber()` - REMOVED
6. ⚠️ `Update()` - UPDATE license (NO AUDIT) ❌
7. ⚠️ `Renew()` - Enhanced renewal logic (NO AUDIT) ❌
8. ⚠️ `FindExpiring()` - Enhanced with pagination + office filter
9. ❌ `FindExpiredLicenses()` - REMOVED
10. ❌ `MarkAsExpired()` - REMOVED
11. ❌ `BatchMarkAsExpired()` - REMOVED
12. ⚠️ `Delete()` - Soft delete (NO AUDIT) ❌
13. ⚠️ `DeactivateExpiredAgents()` - NEW but incomplete (only SELECT, no UPDATE) ❌
14. ❌ `ValidateLicenseNumberUniqueness()` - REMOVED ❌

### Additions in Phase 7:
- ✅ Enhanced renewal calculation logic
- ✅ Integration with LicenseReminderRepository
- ✅ Pagination support in FindExpiring()
- ✅ Office filter in FindExpiring()
- ✅ Provisional→Permanent conversion logic

### Losses in Phase 7:
- ❌ ALL audit logging removed from ALL operations
- ❌ 5 utility methods removed
- ❌ CTE pattern for audit logs not used
- ❌ DeactivateExpiredAgents() only does SELECT, doesn't UPDATE

## Critical Missing Functionality Analysis

### Priority 1: CRITICAL (Must Restore)
1. **Audit Logging in ALL operations** - BR-AGT-PRF-005, FR-AGT-PRF-022
   - Create() needs: CTE for INSERT license + INSERT audit
   - Update() needs: CTE for UPDATE license + INSERT audit
   - Renew() needs: CTE for UPDATE license + INSERT audit
   - Delete() needs: Batch/CTE for soft delete + INSERT audit
   - DeactivateExpiredAgents() needs: UPDATE + INSERT audit (not just SELECT)

2. **ValidateLicenseNumberUniqueness()** - VR-AGT-PRF-020
   - Essential validation method
   - Simple SELECT COUNT query
   - Required for license number uniqueness checks

### Priority 2: Important (Should Restore)
3. **FindPrimaryLicense()** - Get primary license for agent
4. **FindByLicenseNumber()** - Lookup by license number
5. **MarkAsExpired()** - Mark single license as expired with audit
6. **Fix DeactivateExpiredAgents()** - Currently only SELECT, needs UPDATE + audit

### Priority 3: Optional (May Restore)
7. **BatchMarkAsExpired()** - Bulk expiry operation
8. **FindExpiredLicenses()** - Get already expired licenses (vs expiring soon)

## Business Logic Comparison

### Update() Function
**Original**:
- Purpose: Update static license fields (number, resident status, etc.)
- Pattern: CTE for UPDATE + INSERT audit
- Business Rules: FR-AGT-PRF-014, BR-AGT-PRF-012

**New (Phase 7)**:
- Purpose: Same - update static fields
- Pattern: Simple UPDATE (NO AUDIT) ❌
- API: AGT-032

**Conclusion**: Same business purpose, but missing audit logging

### RenewLicense() / Renew()
**Original (RenewLicense)**:
- Purpose: Extend license validity, increment renewal count
- Pattern: CTE for UPDATE + INSERT audit
- Logic: Basic renewal period calculation

**New (Renew)**:
- Purpose: Same + enhanced provisional→permanent conversion
- Pattern: Simple UPDATE (NO AUDIT) ❌
- Logic: Enhanced with exam validation, max provisional renewals check
- API: AGT-033

**Conclusion**: New has better business logic, but missing audit logging

### DeactivateExpiredAgents() vs BatchMarkAsExpired()
**Original (BatchMarkAsExpired)**:
- Purpose: Batch mark licenses as expired
- Pattern: pgx.Batch with UPDATE + INSERT audit for each
- Actual Work: UPDATE license_status + INSERT audit logs

**New (DeactivateExpiredAgents)**:
- Purpose: Batch deactivate expired agents (BR-AGT-PRF-013)
- Pattern: Simple SELECT query ❌
- Actual Work: **ONLY returns agent IDs, doesn't deactivate!**
- Issue: Function is misnamed - should be FindExpiredAgents()

**Conclusion**: New function is incomplete - only finds, doesn't deactivate

## Restoration Strategy Recommendations

### Option A: Surgical Restoration (Recommended)
1. Keep all Phase 7 functionality (AGT-029 to AGT-038 APIs)
2. Restore audit logging using CTE pattern from original
3. Restore ValidateLicenseNumberUniqueness()
4. Restore FindPrimaryLicense() and FindByLicenseNumber()
5. Fix DeactivateExpiredAgents() to actually deactivate

### Option B: Full Revert + Rebuild
1. Revert repo/postgres/agent_license.go to commit 47f3d4a
2. Add Phase 7 enhancements on top of original
3. Keep all audit patterns intact

### Option C: Hybrid Merge
1. Extract audit logging helper method
2. Apply to all current operations
3. Restore missing utility methods
4. Merge best of both implementations

## How to Use This Folder

1. **Compare implementations**:
   ```bash
   diff old/repo/postgres/agent_license.go repo/postgres/agent_license.go
   ```

2. **Extract specific methods**:
   - Copy ValidateLicenseNumberUniqueness() from old version
   - Copy CTE audit patterns from Create(), Update(), etc.
   - Copy FindPrimaryLicense(), FindByLicenseNumber()

3. **Analyze business logic**:
   - Review old RenewLicense() logic
   - Compare with new Renew() logic
   - Merge best parts

4. **Restore audit patterns**:
   - Extract CTE pattern from old Create()
   - Apply to new Create(), Update(), Renew(), Delete()

## Files NOT Included (New in Phase 7)
These are entirely new, nothing to restore:
- `handler/license_management.go` (555 lines) - NEW
- `handler/response/license_management.go` (172 lines) - NEW
- `repo/postgres/license_reminder.go` (150 lines) - NEW
- `core/domain/license_reminder.go` (75 lines) - NEW

## Summary Statistics

| Metric | Original (47f3d4a) | New (Phase 7) | Difference |
|--------|-------------------|---------------|------------|
| **Total Lines** | 641 | 1,559 | +918 lines |
| **Repository Methods** | 13 | 9 | -4 methods |
| **Methods with Audit** | 13 (100%) | 0 (0%) | -13 ❌ |
| **Validation Methods** | 1 | 0 | -1 ❌ |
| **CTE Operations** | 5 | 0 | -5 ❌ |
| **New Handler APIs** | 0 | 10 | +10 ✅ |

## Next Steps
1. ✅ Analyze old implementation patterns (THIS FOLDER)
2. ⏳ Decide restoration strategy (Option A, B, or C)
3. ⏳ Restore audit logging to all operations
4. ⏳ Restore missing utility methods
5. ⏳ Fix DeactivateExpiredAgents() to actually deactivate
6. ⏳ Test merged implementation

---

**Generated**: 2026-01-27
**Commit Preserved**: 47f3d4a (pre-Phase 7)
**Current Branch**: claude/phase-7-license-management-jt5yr
