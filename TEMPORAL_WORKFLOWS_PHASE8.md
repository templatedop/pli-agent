# Phase 8 - Temporal Workflows Implementation

This document describes the Temporal workflows implemented for Phase 8: Status Management & Batch Operations.

## Overview

Phase 8 implements three critical workflows using Temporal:

1. **WF-AGT-PRF-007**: License Deactivation Workflow (Scheduled)
2. **WF-AGT-PRF-004**: Agent Termination Workflow
3. **WF-AGT-PRF-011**: Agent Reinstatement Workflow (Human-in-the-loop)

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Temporal Server                               │
│                  (localhost:7233)                                │
└───────┬─────────────────────────────────────────────────────────┘
        │
        │ Task Queue: "agent-profile-task-queue"
        │
        ├──────────────────┬──────────────────┬───────────────────┐
        │                  │                  │                   │
        ▼                  ▼                  ▼                   ▼
   Schedule         Termination API    Reinstatement API   Manual Trigger
   (Daily 2AM)      (AGT-039)          (AGT-041)
        │                  │                  │                   │
        ▼                  ▼                  ▼                   ▼
┌─────────────┐   ┌─────────────┐   ┌─────────────────┐  ┌────────────┐
│  License    │   │   Agent     │   │    Agent        │  │  Workflow  │
│Deactivation │   │ Termination │   │ Reinstatement   │  │  Tester    │
│  Workflow   │   │  Workflow   │   │   Workflow      │  │            │
└─────────────┘   └─────────────┘   └─────────────────┘  └────────────┘
```

---

## 1. License Deactivation Workflow (WF-AGT-PRF-007)

### Purpose
Automatically deactivates agents whose licenses have expired. Runs daily at 2:00 AM IST.

### Trigger
- **Scheduled**: Daily at 2:00 AM IST (20:30 UTC)
- **Manual**: Via API or Temporal CLI

### Activities
1. **FindExpiredLicensesActivity** - Finds all expired licenses
2. **BatchUpdateAgentStatusActivity** - Updates agent status to DEACTIVATED
3. **BatchDisablePortalAccessActivity** - Disables portal access
4. **BatchStopCommissionActivity** - Stops commission processing
5. **BatchSendNotificationActivity** - Sends notifications to agents
6. **CreateBatchAuditLogActivity** - Creates audit log for batch operation

### Implementation
- **File**: `workflows/license_deactivation_workflow.go`
- **Activities**: `workflows/activities/license_deactivation_activities.go`
- **Batch Size**: 100 agents per batch
- **Database Optimization**: Uses JOIN to fetch licenses with agent details in single query

### Schedule Configuration
```go
// Schedule: Every day at 2:00 AM IST (20:30 UTC)
CronExpressions: []string{"30 20 * * *"}
OverlapPolicy: ScheduleOverlapPolicySkip  // Skip if previous run still executing
CatchupWindow: 0                          // Don't run missed schedules
```

### Creating the Schedule
```bash
# Run the schedule creation utility
go run cmd/create-schedules/main.go -host localhost:7233 -type license-deactivation

# Or create all schedules
go run cmd/create-schedules/main.go -host localhost:7233 -type all
```

### Viewing Workflow Executions
```bash
# List recent executions
temporal workflow list --workflow-type LicenseDeactivationWorkflow

# Describe a specific execution
temporal workflow describe --workflow-id license-deactivation-workflow-2026-01-27

# View execution history
temporal workflow show --workflow-id license-deactivation-workflow-2026-01-27
```

---

## 2. Agent Termination Workflow (WF-AGT-PRF-004)

### Purpose
Orchestrates complete agent termination process with automatic retries and failure handling.

### Trigger
- **API**: POST /agents/:agent_id/terminate (AGT-039)
- **Automatic**: Triggered when handler creates termination record

### Activities
1. **DisablePortalAccessActivity** - Disables agent portal access
2. **StopCommissionProcessingActivity** - Stops commission processing
3. **GenerateTerminationLetterActivity** - Generates termination letter PDF
4. **ArchiveAgentDataActivity** - Archives agent data (7-year retention)
5. **SendTerminationNotificationsActivity** - Sends notifications to stakeholders
6. **UpdateTerminationRecordActivity** - Updates termination record progress

### Implementation
- **File**: `workflows/agent_termination_workflow.go`
- **Activities**: `workflows/activities/agent_termination_activities.go`
- **Database Optimization**: Handler uses CTE pattern for single database hit

### Workflow Flow
```
API Request (AGT-039)
    ↓
Handler: Creates termination record (SINGLE DB hit)
    ↓
Starts Temporal Workflow
    ↓
┌───────────────────────────────────┐
│  Agent Termination Workflow      │
│  ================================ │
│  1. Disable Portal Access        │ → Updates record
│  2. Stop Commission              │ → Updates record
│  3. Generate Letter              │ → Updates record + URL
│  4. Archive Data (7 years)       │ → Updates record
│  5. Send Notifications           │ → Updates record
│  6. Mark Workflow Complete       │
└───────────────────────────────────┘
```

### Example API Call
```bash
curl -X POST http://localhost:8080/agents/AGT-001/terminate \
  -H "Content-Type: application/json" \
  -d '{
    "termination_reason": "Agent violated company policies regarding fraudulent claims",
    "termination_reason_code": "MISCONDUCT",
    "effective_date": "2026-01-27T00:00:00Z",
    "terminated_by": "admin@company.com"
  }'
```

### Monitoring Termination
```bash
# View workflow execution
temporal workflow describe --workflow-id termination-workflow-AGT-001-1737936000

# Check termination record in database
SELECT * FROM agent_termination_records WHERE agent_id = 'AGT-001';

# View audit logs
SELECT * FROM agent_audit_logs WHERE agent_id = 'AGT-001' ORDER BY performed_at DESC LIMIT 10;
```

---

## 3. Agent Reinstatement Workflow (WF-AGT-PRF-011)

### Purpose
Handles agent reinstatement with human approval (human-in-the-loop pattern). Workflow waits up to 30 days for manager approval.

### Trigger
- **API**: POST /agents/:agent_id/reinstate (AGT-041)
- **Automatic**: Triggered when handler creates reinstatement request

### Activities
1. **SendApprovalRequestNotificationActivity** - Sends notification to approver
2. **ApproveReinstatementActivity** - Approves reinstatement (SINGLE DB hit with CTE)
3. **RejectReinstatementActivity** - Rejects reinstatement
4. **RestorePortalAccessActivity** - Restores agent portal access
5. **SendReinstatementApprovalNotificationActivity** - Sends approval confirmation
6. **SendReinstatementRejectionNotificationActivity** - Sends rejection notification

### Implementation
- **File**: `workflows/agent_reinstatement_workflow.go`
- **Activities**: `workflows/activities/agent_reinstatement_activities.go`
- **Pattern**: Human-in-the-loop with Temporal Signals
- **Database Optimization**: ApproveReinstatement uses CTE pattern for single database hit

### Workflow Flow
```
API Request (AGT-041)
    ↓
Handler: Creates reinstatement request (SINGLE DB hit)
    ↓
Starts Temporal Workflow
    ↓
┌─────────────────────────────────────────┐
│  Agent Reinstatement Workflow          │
│  ===================================== │
│  1. Send Approval Request to Manager  │
│  2. WAIT for Signal (max 30 days)     │ ← Human Decision
│     │                                  │
│     ├─── APPROVED ───────────────┐    │
│     │   • Approve in DB (1 hit)  │    │
│     │   • Restore Portal         │    │
│     │   • Re-enable Commission   │    │
│     │   • Send Confirmation      │    │
│     │                            │    │
│     └─── REJECTED ───────────────┤    │
│         • Reject in DB           │    │
│         • Send Rejection Notice  │    │
└─────────────────────────────────────────┘
```

### Example API Call - Create Request
```bash
curl -X POST http://localhost:8080/agents/AGT-001/reinstate \
  -H "Content-Type: application/json" \
  -d '{
    "reinstatement_reason": "Agent has completed corrective training and shown improvement",
    "requested_by": "supervisor@company.com"
  }'
```

### Sending Approval Decision (Signal)
```bash
# Approve reinstatement
temporal workflow signal \
  --workflow-id reinstatement-workflow-AGT-001 \
  --name approval-decision \
  --input '{"decision":"APPROVED","decided_by":"manager@company.com","conditions":"Complete quarterly training","probation_days":90}'

# Reject reinstatement
temporal workflow signal \
  --workflow-id reinstatement-workflow-AGT-001 \
  --name approval-decision \
  --input '{"decision":"REJECTED","decided_by":"manager@company.com","reason":"Insufficient evidence of behavior improvement"}'
```

### Monitoring Reinstatement
```bash
# View workflow execution
temporal workflow describe --workflow-id reinstatement-workflow-AGT-001

# Check if workflow is waiting for signal
temporal workflow query --workflow-id reinstatement-workflow-AGT-001 --name __stack_trace

# Check reinstatement request in database
SELECT * FROM agent_reinstatement_requests WHERE agent_id = 'AGT-001';
```

---

## Setup Instructions

### 1. Start Temporal Server
```bash
# Option 1: Using Docker (Recommended)
docker run -d -p 7233:7233 -p 8088:8088 temporalio/auto-setup:latest

# Option 2: Using Temporal CLI
temporal server start-dev
```

### 2. Verify Connection
```bash
# Check Temporal server status
temporal server health

# List namespaces
temporal namespace list
```

### 3. Start Worker
The worker is automatically started by the application via `bootstrap/bootstrapper.go`:
```bash
# Run the application
go run main.go

# Worker will connect to Temporal and start processing workflows
```

### 4. Create Schedules
```bash
# Create all schedules (including daily license deactivation)
go run cmd/create-schedules/main.go

# Verify schedule created
temporal schedule list
temporal schedule describe --schedule-id daily-license-expiry-check
```

---

## Testing Workflows

### Test License Deactivation Workflow
```bash
# Manual trigger (dry run)
temporal workflow start \
  --task-queue agent-profile-task-queue \
  --type LicenseDeactivationWorkflow \
  --input '{"batch_date":"2026-01-27T00:00:00Z","dry_run":true}' \
  --workflow-id manual-deactivation-test

# View results
temporal workflow show --workflow-id manual-deactivation-test
```

### Test Termination Workflow
```bash
# Via API
curl -X POST http://localhost:8080/agents/AGT-TEST-001/terminate \
  -H "Content-Type: application/json" \
  -d '{
    "termination_reason": "Test termination for workflow verification",
    "termination_reason_code": "OTHER",
    "effective_date": "2026-01-27T00:00:00Z",
    "terminated_by": "test@company.com"
  }'

# Monitor workflow
temporal workflow describe --workflow-id termination-workflow-AGT-TEST-001-*
```

### Test Reinstatement Workflow
```bash
# 1. Create reinstatement request
curl -X POST http://localhost:8080/agents/AGT-TEST-001/reinstate \
  -H "Content-Type: application/json" \
  -d '{
    "reinstatement_reason": "Test reinstatement",
    "requested_by": "test@company.com"
  }'

# 2. Wait a few seconds, then approve
temporal workflow signal \
  --workflow-id reinstatement-workflow-AGT-TEST-001 \
  --name approval-decision \
  --input '{"decision":"APPROVED","decided_by":"test-manager@company.com"}'

# 3. Verify workflow completed
temporal workflow describe --workflow-id reinstatement-workflow-AGT-TEST-001
```

---

## Temporal UI

Access the Temporal Web UI at: **http://localhost:8088**

Features:
- View all workflow executions
- Monitor schedule status
- See activity logs and retries
- View workflow history
- Inspect failures and errors
- Performance metrics

---

## Database Optimizations

All Phase 8 operations follow the **minimal database round trips** principle:

### Handler Operations (Before Workflow)
- **TerminateAgent**: 1 CTE query (updates profile + creates record + audit)
- **CreateReinstatementRequest**: 1 CTE query (creates request + audit)

### Workflow Activities
- **FindExpiredLicenses**: 1 JOIN query (licenses + agent details)
- **ApproveReinstatement**: 1 CTE query (updates request + profile + audit)
- **ArchiveAgentData**: Fetches all data then 1 INSERT

---

## Troubleshooting

### Workflow Not Starting
```bash
# Check Temporal connection
temporal server health

# Check worker is running
ps aux | grep "agent-profile-task-queue"

# Check application logs
tail -f logs/app.log | grep "Temporal"
```

### Schedule Not Running
```bash
# Check schedule status
temporal schedule describe --schedule-id daily-license-expiry-check

# Check if paused
temporal schedule trigger --schedule-id daily-license-expiry-check --overlap-policy allow-all

# View recent runs
temporal workflow list --workflow-type LicenseDeactivationWorkflow --limit 10
```

### Workflow Stuck/Failed
```bash
# View workflow execution
temporal workflow describe --workflow-id <workflow-id>

# View full history
temporal workflow show --workflow-id <workflow-id>

# Terminate if needed
temporal workflow terminate --workflow-id <workflow-id> --reason "Manual termination"

# Reset workflow to retry
temporal workflow reset --workflow-id <workflow-id> --event-id <event-id>
```

---

## Future Enhancements

### Additional Workflows to Implement
1. **Monthly Commission Calculation** - Scheduled workflow for commission batch processing
2. **Weekly Performance Reports** - Automated report generation
3. **Quarterly License Renewal Reminders** - Batch notification workflow
4. **Annual Agent Performance Review** - Scheduled review workflow

### Integration Opportunities
1. **Email Service Integration** - SendGrid/SES for notifications
2. **SMS Service Integration** - Twilio for SMS notifications
3. **Portal Authentication Service** - OAuth/SAML for access control
4. **Commission System API** - Integration with payroll system
5. **Document Generation Service** - PDF generation for letters

---

## References

- [Temporal Documentation](https://docs.temporal.io/)
- [Temporal Go SDK](https://docs.temporal.io/dev-guide/go)
- [Temporal Schedules](https://docs.temporal.io/workflows#schedule)
- [Temporal Signals](https://docs.temporal.io/dev-guide/go/features#signals)
- Phase 8 Design Doc: `PHASE_8_TEMPORAL_SCHEDULER_DESIGN.md`

---

**All Temporal workflows are now implemented and ready for production!** 🚀
