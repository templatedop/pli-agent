# Phase 8 - Temporal Scheduler Implementation Design

**Date**: 2026-01-27
**Phase**: 8 - Status Management & Batch Operations
**Feature**: Scheduled License Deactivation using Temporal Schedules

---

## Overview

Phase 8 includes scheduled batch operations that run automatically (e.g., daily license expiry checks). We'll implement these using **Temporal Schedules** instead of traditional cron jobs.

### Why Temporal Schedules?

✅ **Built-in Reliability**: Automatic retries, failure handling, visibility
✅ **No External Cron**: No need to maintain cron jobs or K8s CronJobs
✅ **Workflow Benefits**: Full workflow orchestration, state management, audit trail
✅ **Dynamic Control**: Can pause, resume, update schedules programmatically
✅ **Already Integrated**: Your codebase already has Temporal configured!

---

## Phase 8 Scheduled Operations

### 1. **Daily License Expiry Check** (WF-AGT-PRF-007)

**Schedule**: Every day at 2:00 AM
**Workflow**: `LicenseDeactivationWorkflow`
**Purpose**: Find and deactivate agents with expired licenses

**Actions**:
1. Find all licenses where `renewal_date < today`
2. Batch process (100 agents at a time):
   - Update agent status to `DEACTIVATED`
   - Disable portal access
   - Stop commission processing
   - Send notification
   - Create audit logs

---

## Temporal Schedule Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Temporal Schedule                         │
│  Schedule ID: "daily-license-expiry-check"                  │
│  Cron: "0 2 * * *" (Every day at 2:00 AM)                   │
└────────────────┬────────────────────────────────────────────┘
                 │
                 │ Triggers
                 ▼
┌─────────────────────────────────────────────────────────────┐
│           LicenseDeactivationWorkflow                        │
│  - Input: { batch_date: "today" }                           │
│  - Activities:                                               │
│    1. FindExpiredLicensesActivity                           │
│    2. BatchUpdateAgentStatusActivity (100 at a time)        │
│    3. BatchDisablePortalAccessActivity                      │
│    4. BatchStopCommissionActivity                           │
│    5. BatchSendNotificationActivity                         │
│    6. CreateBatchAuditLogActivity                           │
│  - Output: DeactivationSummary                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Implementation Steps

### Step 1: Create the Workflow

**File**: `/home/user/pli-agent/workflows/license_deactivation_workflow.go`

```go
package workflows

import (
    "time"
    "go.temporal.io/sdk/workflow"
)

// DeactivationInput is the input for license deactivation workflow
type DeactivationInput struct {
    BatchDate time.Time
    DryRun    bool
}

// DeactivationOutput is the output summary
type DeactivationOutput struct {
    BatchID              string
    BatchDate            time.Time
    TotalExpiredLicenses int
    AgentsDeactivated    int
    PortalAccessDisabled int
    NotificationsSent    int
    ProcessingTimeSeconds int
}

// LicenseDeactivationWorkflow deactivates agents with expired licenses
// WF-AGT-PRF-007: License Deactivation Workflow
func LicenseDeactivationWorkflow(ctx workflow.Context, input DeactivationInput) (*DeactivationOutput, error) {
    logger := workflow.GetLogger(ctx)
    logger.Info("Starting license deactivation workflow", "batch_date", input.BatchDate)

    // Set workflow options
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 10 * time.Minute,
        RetryPolicy: &temporal.RetryPolicy{
            MaximumAttempts: 3,
        },
    }
    ctx = workflow.WithActivityOptions(ctx, ao)

    output := &DeactivationOutput{
        BatchID:   workflow.GetInfo(ctx).WorkflowExecution.ID,
        BatchDate: input.BatchDate,
    }

    // Step 1: Find expired licenses
    var expiredLicenses []ExpiredLicense
    err := workflow.ExecuteActivity(ctx, "FindExpiredLicensesActivity", input.BatchDate).Get(ctx, &expiredLicenses)
    if err != nil {
        return nil, err
    }
    output.TotalExpiredLicenses = len(expiredLicenses)
    logger.Info("Found expired licenses", "count", len(expiredLicenses))

    if input.DryRun {
        logger.Info("Dry run - skipping updates")
        return output, nil
    }

    // Step 2: Process in batches of 100
    batchSize := 100
    for i := 0; i < len(expiredLicenses); i += batchSize {
        end := i + batchSize
        if end > len(expiredLicenses) {
            end = len(expiredLicenses)
        }
        batch := expiredLicenses[i:end]

        // Activity 1: Update agent status to DEACTIVATED
        var statusResult BatchUpdateResult
        err = workflow.ExecuteActivity(ctx, "BatchUpdateAgentStatusActivity", batch).Get(ctx, &statusResult)
        if err != nil {
            logger.Error("Failed to update agent status", "error", err)
            continue
        }
        output.AgentsDeactivated += statusResult.SuccessCount

        // Activity 2: Disable portal access
        var portalResult BatchUpdateResult
        err = workflow.ExecuteActivity(ctx, "BatchDisablePortalAccessActivity", batch).Get(ctx, &portalResult)
        if err != nil {
            logger.Error("Failed to disable portal access", "error", err)
        }
        output.PortalAccessDisabled += portalResult.SuccessCount

        // Activity 3: Stop commission processing
        err = workflow.ExecuteActivity(ctx, "BatchStopCommissionActivity", batch).Get(ctx, nil)
        if err != nil {
            logger.Error("Failed to stop commission", "error", err)
        }

        // Activity 4: Send notifications
        var notifResult BatchUpdateResult
        err = workflow.ExecuteActivity(ctx, "BatchSendNotificationActivity", batch).Get(ctx, &notifResult)
        if err != nil {
            logger.Error("Failed to send notifications", "error", err)
        }
        output.NotificationsSent += notifResult.SuccessCount

        logger.Info("Processed batch", "batch_number", i/batchSize+1, "size", len(batch))
    }

    // Step 3: Create audit log
    err = workflow.ExecuteActivity(ctx, "CreateBatchAuditLogActivity", output).Get(ctx, nil)
    if err != nil {
        logger.Error("Failed to create audit log", "error", err)
    }

    logger.Info("License deactivation workflow completed", "deactivated", output.AgentsDeactivated)
    return output, nil
}
```

---

### Step 2: Create Activities

**File**: `/home/user/pli-agent/workflows/activities/license_deactivation_activities.go`

```go
package activities

import (
    "context"
    "time"
    "pli-agent-api/repo/postgres"
    "go.temporal.io/sdk/activity"
)

type LicenseDeactivationActivities struct {
    licenseRepo *repo.AgentLicenseRepository
    profileRepo *repo.AgentProfileRepository
}

func NewLicenseDeactivationActivities(
    licenseRepo *repo.AgentLicenseRepository,
    profileRepo *repo.AgentProfileRepository,
) *LicenseDeactivationActivities {
    return &LicenseDeactivationActivities{
        licenseRepo: licenseRepo,
        profileRepo: profileRepo,
    }
}

// ExpiredLicense represents a license that has expired
type ExpiredLicense struct {
    LicenseID string
    AgentID   string
}

// BatchUpdateResult contains batch operation results
type BatchUpdateResult struct {
    TotalCount   int
    SuccessCount int
    FailedCount  int
}

// FindExpiredLicensesActivity finds all licenses that have expired
// ACT-DEC-001: Find Expired Licenses
func (a *LicenseDeactivationActivities) FindExpiredLicensesActivity(ctx context.Context, batchDate time.Time) ([]ExpiredLicense, error) {
    logger := activity.GetLogger(ctx)
    logger.Info("Finding expired licenses", "batch_date", batchDate)

    // Use repository method to find expired licenses
    licenses, err := a.licenseRepo.FindExpiredLicenses(ctx)
    if err != nil {
        return nil, err
    }

    result := make([]ExpiredLicense, len(licenses))
    for i, license := range licenses {
        result[i] = ExpiredLicense{
            LicenseID: license.LicenseID,
            AgentID:   license.AgentID,
        }
    }

    logger.Info("Found expired licenses", "count", len(result))
    return result, nil
}

// BatchUpdateAgentStatusActivity updates agent status to DEACTIVATED in batch
// ACT-DEC-002: Batch Update Agent Status
func (a *LicenseDeactivationActivities) BatchUpdateAgentStatusActivity(ctx context.Context, batch []ExpiredLicense) (BatchUpdateResult, error) {
    logger := activity.GetLogger(ctx)
    logger.Info("Updating agent status", "count", len(batch))

    // Extract agent IDs
    agentIDs := make([]string, len(batch))
    for i, item := range batch {
        agentIDs[i] = item.AgentID
    }

    // Batch update using repository method
    // TODO: Create BatchUpdateAgentStatus repository method
    successCount := 0
    failedCount := 0

    for _, agentID := range agentIDs {
        // Update agent status to DEACTIVATED
        // For now, use individual updates - can be optimized later with UNNEST
        err := a.profileRepo.UpdateStatus(ctx, agentID, "DEACTIVATED", "SYSTEM", "License expired")
        if err != nil {
            failedCount++
            logger.Error("Failed to update agent status", "agent_id", agentID, "error", err)
        } else {
            successCount++
        }
    }

    logger.Info("Agent status update complete", "success", successCount, "failed", failedCount)
    return BatchUpdateResult{
        TotalCount:   len(batch),
        SuccessCount: successCount,
        FailedCount:  failedCount,
    }, nil
}

// BatchDisablePortalAccessActivity disables portal access for agents
// ACT-DEC-003: Batch Disable Portal Access
func (a *LicenseDeactivationActivities) BatchDisablePortalAccessActivity(ctx context.Context, batch []ExpiredLicense) (BatchUpdateResult, error) {
    logger := activity.GetLogger(ctx)
    logger.Info("Disabling portal access", "count", len(batch))

    // TODO: Implement portal access disable logic
    // This would typically call an external API or update a portal_access table

    return BatchUpdateResult{
        TotalCount:   len(batch),
        SuccessCount: len(batch), // Assuming all succeed for now
        FailedCount:  0,
    }, nil
}

// BatchStopCommissionActivity stops commission processing for agents
// ACT-DEC-004: Batch Stop Commission
func (a *LicenseDeactivationActivities) BatchStopCommissionActivity(ctx context.Context, batch []ExpiredLicense) error {
    logger := activity.GetLogger(ctx)
    logger.Info("Stopping commission processing", "count", len(batch))

    // TODO: Implement commission stop logic
    // This would typically update a commission_processing_status table or call external API

    return nil
}

// BatchSendNotificationActivity sends notifications to agents
// ACT-DEC-005: Batch Send Notification
func (a *LicenseDeactivationActivities) BatchSendNotificationActivity(ctx context.Context, batch []ExpiredLicense) (BatchUpdateResult, error) {
    logger := activity.GetLogger(ctx)
    logger.Info("Sending notifications", "count", len(batch))

    // TODO: Implement notification sending
    // This would call a notification service (email/SMS)

    return BatchUpdateResult{
        TotalCount:   len(batch),
        SuccessCount: len(batch), // Assuming all succeed for now
        FailedCount:  0,
    }, nil
}

// CreateBatchAuditLogActivity creates audit log for batch operation
// ACT-DEC-006: Create Batch Audit Log
func (a *LicenseDeactivationActivities) CreateBatchAuditLogActivity(ctx context.Context, output interface{}) error {
    logger := activity.GetLogger(ctx)
    logger.Info("Creating batch audit log")

    // TODO: Create audit log entry for batch operation
    // Store in agent_batch_operations table

    return nil
}
```

---

### Step 3: Register Workflow and Activities

**File**: `/home/user/pli-agent/bootstrap/bootstrapper.go`

```go
// Add to FxTemporal module
fx.Provide(
    // Existing activities
    activities.NewAgentOnboardingActivities,

    // NEW: License deactivation activities
    activities.NewLicenseDeactivationActivities,
),

fx.Invoke(
    func(c client.Client, cfg *config.Config,
         onboardingActivities *activities.AgentOnboardingActivities,
         deactivationActivities *activities.LicenseDeactivationActivities) error {

        w := worker.New(c, "agent-profile-task-queue", worker.Options{})

        // Register existing workflows
        w.RegisterWorkflow(workflows.AgentOnboardingWorkflow)

        // Register NEW workflow
        w.RegisterWorkflow(workflows.LicenseDeactivationWorkflow)

        // Register existing activities
        w.RegisterActivity(onboardingActivities)

        // Register NEW activities
        w.RegisterActivity(deactivationActivities)

        return w.Start()
    },
),
```

---

### Step 4: Create Temporal Schedule

**Two Options**: Code-based or CLI-based

#### **Option A: Create Schedule via Code** (Recommended)

**File**: `/home/user/pli-agent/cmd/create-schedules/main.go`

```go
package main

import (
    "context"
    "log"
    "time"

    "go.temporal.io/sdk/client"
)

func main() {
    // Connect to Temporal
    c, err := client.NewClient(client.Options{
        HostPort:  "localhost:7233",
        Namespace: "default",
    })
    if err != nil {
        log.Fatalln("Unable to create Temporal client", err)
    }
    defer c.Close()

    // Create schedule for daily license expiry check
    scheduleID := "daily-license-expiry-check"

    schedule := client.ScheduleOptions{
        ID: scheduleID,
        Spec: client.ScheduleSpec{
            // Every day at 2:00 AM IST (20:30 UTC)
            CronExpressions: []string{"30 20 * * *"},
            // Or use CalendarSpecs for more complex schedules
        },
        Action: &client.ScheduleWorkflowAction{
            ID:        "license-deactivation-workflow",
            Workflow:  workflows.LicenseDeactivationWorkflow,
            TaskQueue: "agent-profile-task-queue",
            Args: []interface{}{
                workflows.DeactivationInput{
                    BatchDate: time.Now(),
                    DryRun:    false,
                },
            },
        },
        // Overlap policy: Skip if previous run is still executing
        OverlapPolicy: client.ScheduleOverlapPolicySkip,
    }

    handle, err := c.ScheduleClient().Create(context.Background(), schedule)
    if err != nil {
        log.Fatalln("Unable to create schedule", err)
    }

    log.Printf("Schedule created: %s", handle.GetID())
}
```

**Run**: `go run cmd/create-schedules/main.go`

---

#### **Option B: Create Schedule via Temporal CLI**

```bash
# Install Temporal CLI
go install go.temporal.io/temporalcli@latest

# Create schedule
temporal schedule create \
  --schedule-id daily-license-expiry-check \
  --workflow-id license-deactivation-workflow \
  --workflow-type LicenseDeactivationWorkflow \
  --task-queue agent-profile-task-queue \
  --cron "30 20 * * *" \
  --overlap-policy Skip
```

---

### Step 5: Manual Trigger (For Testing)

You can manually trigger the workflow via API handler:

**File**: `/home/user/pli-agent/handler/license_management.go`

```go
// ManualTriggerDeactivation manually triggers license deactivation workflow
// For testing or manual execution
func (h *AgentLicenseHandler) ManualTriggerDeactivation(sctx *serverRoute.Context, req ManualTriggerRequest) (*resp.WorkflowTriggerResponse, error) {
    // Get Temporal client from context
    temporalClient := sctx.Get("temporal_client").(client.Client)

    workflowOptions := client.StartWorkflowOptions{
        ID:        fmt.Sprintf("manual-deactivation-%s", time.Now().Format("20060102-150405")),
        TaskQueue: "agent-profile-task-queue",
    }

    input := workflows.DeactivationInput{
        BatchDate: time.Now(),
        DryRun:    req.DryRun,
    }

    we, err := temporalClient.ExecuteWorkflow(sctx.Ctx, workflowOptions, workflows.LicenseDeactivationWorkflow, input)
    if err != nil {
        return nil, err
    }

    return &resp.WorkflowTriggerResponse{
        WorkflowID: we.GetID(),
        RunID:      we.GetRunID(),
        Message:    "Workflow started successfully",
    }, nil
}
```

---

## Managing Schedules

### Pause Schedule
```bash
temporal schedule pause --schedule-id daily-license-expiry-check
```

### Resume Schedule
```bash
temporal schedule unpause --schedule-id daily-license-expiry-check
```

### Update Schedule
```go
handle := c.ScheduleClient().GetHandle(ctx, "daily-license-expiry-check")
err := handle.Update(ctx, client.ScheduleUpdateOptions{
    DoUpdate: func(input client.ScheduleUpdateInput) (*client.ScheduleUpdate, error) {
        // Change to 3:00 AM
        input.Description.Schedule.Spec.CronExpressions = []string{"0 3 * * *"}
        return &client.ScheduleUpdate{
            Schedule: &input.Description.Schedule,
        }, nil
    },
})
```

### Delete Schedule
```bash
temporal schedule delete --schedule-id daily-license-expiry-check
```

---

## Benefits Over Traditional Cron

| Feature | Traditional Cron | Temporal Schedule |
|---------|-----------------|-------------------|
| **Reliability** | ❌ No retries | ✅ Automatic retries |
| **Visibility** | ❌ No history | ✅ Full execution history |
| **Monitoring** | ❌ Manual setup | ✅ Built-in metrics |
| **State Management** | ❌ None | ✅ Workflow state persisted |
| **Failure Handling** | ❌ Manual | ✅ Automatic |
| **Distributed** | ❌ Single server | ✅ Distributed by default |
| **Audit Trail** | ❌ Limited | ✅ Complete audit trail |

---

## Next Steps for Phase 8

1. ✅ **Already Complete**: Phase 7 License Management APIs
2. 🔄 **Create Workflow**: Implement `LicenseDeactivationWorkflow`
3. 🔄 **Create Activities**: Implement 6 activities for deactivation
4. 🔄 **Register**: Add to bootstrap
5. 🔄 **Create Schedule**: Set up daily schedule
6. 🔄 **Test**: Manual trigger and verify
7. 🔄 **Monitor**: Use Temporal UI to monitor executions

---

## Temporal UI

Access at: `http://localhost:8088` (if running locally)

You'll see:
- All workflow executions
- Schedule status
- Activity logs
- Retry history
- Performance metrics

---

**All workflows in Phase 8-10 can use this same pattern!** 🚀
