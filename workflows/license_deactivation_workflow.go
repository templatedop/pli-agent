package workflows

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// DeactivationInput is the input for license deactivation workflow
// WF-AGT-PRF-007: License Deactivation Workflow
type DeactivationInput struct {
	BatchDate time.Time `json:"batch_date"`
	DryRun    bool      `json:"dry_run"`
}

// DeactivationOutput is the output summary
type DeactivationOutput struct {
	BatchID                 string    `json:"batch_id"`
	BatchDate               time.Time `json:"batch_date"`
	TotalExpiredLicenses    int       `json:"total_expired_licenses"`
	AgentsDeactivated       int       `json:"agents_deactivated"`
	PortalAccessDisabled    int       `json:"portal_access_disabled"`
	CommissionStopped       int       `json:"commission_stopped"`
	NotificationsSent       int       `json:"notifications_sent"`
	ProcessingTimeSeconds   int       `json:"processing_time_seconds"`
	FailedDeactivations     int       `json:"failed_deactivations"`
	FailedNotifications     int       `json:"failed_notifications"`
}

// ExpiredLicenseInfo represents a license that has expired
type ExpiredLicenseInfo struct {
	LicenseID   string `json:"license_id"`
	AgentID     string `json:"agent_id"`
	AgentCode   string `json:"agent_code"`
	AgentName   string `json:"agent_name"`
	LicenseLine string `json:"license_line"`
	RenewalDate string `json:"renewal_date"`
}

// BatchUpdateResult contains batch operation results
type BatchUpdateResult struct {
	TotalCount   int      `json:"total_count"`
	SuccessCount int      `json:"success_count"`
	FailedCount  int      `json:"failed_count"`
	FailedIDs    []string `json:"failed_ids,omitempty"`
}

// LicenseDeactivationWorkflow deactivates agents with expired licenses
// WF-AGT-PRF-007: License Deactivation Workflow
// BR-AGT-PRF-013: Auto-Deactivation on License Expiry
//
// This workflow runs daily at 2:00 AM via Temporal Schedule
// Actions:
// 1. Find all expired licenses
// 2. Batch process (100 agents at a time):
//    - Update agent status to DEACTIVATED
//    - Disable portal access
//    - Stop commission processing
//    - Send notification
//    - Create audit logs
func LicenseDeactivationWorkflow(ctx workflow.Context, input DeactivationInput) (*DeactivationOutput, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting license deactivation workflow", "batch_date", input.BatchDate, "dry_run", input.DryRun)

	startTime := workflow.Now(ctx)

	// Set workflow options with retries
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		HeartbeatTimeout:    30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	output := &DeactivationOutput{
		BatchID:   workflow.GetInfo(ctx).WorkflowExecution.ID,
		BatchDate: input.BatchDate,
	}

	// Step 1: Find expired licenses
	logger.Info("Step 1: Finding expired licenses")
	var expiredLicenses []ExpiredLicenseInfo
	err := workflow.ExecuteActivity(ctx, "FindExpiredLicensesActivity", input.BatchDate).Get(ctx, &expiredLicenses)
	if err != nil {
		logger.Error("Failed to find expired licenses", "error", err)
		return nil, err
	}
	output.TotalExpiredLicenses = len(expiredLicenses)
	logger.Info("Found expired licenses", "count", len(expiredLicenses))

	if len(expiredLicenses) == 0 {
		logger.Info("No expired licenses found - workflow complete")
		output.ProcessingTimeSeconds = int(workflow.Now(ctx).Sub(startTime).Seconds())
		return output, nil
	}

	if input.DryRun {
		logger.Info("Dry run - skipping updates")
		output.ProcessingTimeSeconds = int(workflow.Now(ctx).Sub(startTime).Seconds())
		return output, nil
	}

	// Step 2: Process in batches of 100
	batchSize := 100
	totalBatches := (len(expiredLicenses) + batchSize - 1) / batchSize

	for i := 0; i < len(expiredLicenses); i += batchSize {
		end := i + batchSize
		if end > len(expiredLicenses) {
			end = len(expiredLicenses)
		}
		batch := expiredLicenses[i:end]
		batchNum := i/batchSize + 1

		logger.Info("Processing batch", "batch_number", batchNum, "total_batches", totalBatches, "size", len(batch))

		// Activity 1: Update agent status to DEACTIVATED
		var statusResult BatchUpdateResult
		err = workflow.ExecuteActivity(ctx, "BatchUpdateAgentStatusActivity", batch).Get(ctx, &statusResult)
		if err != nil {
			logger.Error("Failed to update agent status", "batch", batchNum, "error", err)
			// Continue with next batch even if this fails
			output.FailedDeactivations += len(batch)
			continue
		}
		output.AgentsDeactivated += statusResult.SuccessCount
		output.FailedDeactivations += statusResult.FailedCount

		// Activity 2: Disable portal access
		var portalResult BatchUpdateResult
		err = workflow.ExecuteActivity(ctx, "BatchDisablePortalAccessActivity", batch).Get(ctx, &portalResult)
		if err != nil {
			logger.Error("Failed to disable portal access", "batch", batchNum, "error", err)
		} else {
			output.PortalAccessDisabled += portalResult.SuccessCount
		}

		// Activity 3: Stop commission processing
		var commissionResult BatchUpdateResult
		err = workflow.ExecuteActivity(ctx, "BatchStopCommissionActivity", batch).Get(ctx, &commissionResult)
		if err != nil {
			logger.Error("Failed to stop commission", "batch", batchNum, "error", err)
		} else {
			output.CommissionStopped += commissionResult.SuccessCount
		}

		// Activity 4: Send notifications
		var notifResult BatchUpdateResult
		err = workflow.ExecuteActivity(ctx, "BatchSendNotificationActivity", batch).Get(ctx, &notifResult)
		if err != nil {
			logger.Error("Failed to send notifications", "batch", batchNum, "error", err)
			output.FailedNotifications += len(batch)
		} else {
			output.NotificationsSent += notifResult.SuccessCount
			output.FailedNotifications += notifResult.FailedCount
		}

		logger.Info("Batch complete", "batch_number", batchNum, "deactivated", statusResult.SuccessCount)
	}

	// Step 3: Create audit log for entire batch operation
	logger.Info("Step 3: Creating batch audit log")
	err = workflow.ExecuteActivity(ctx, "CreateBatchAuditLogActivity", output).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to create audit log", "error", err)
		// Don't fail workflow if audit log fails
	}

	output.ProcessingTimeSeconds = int(workflow.Now(ctx).Sub(startTime).Seconds())
	logger.Info("License deactivation workflow completed",
		"deactivated", output.AgentsDeactivated,
		"failed", output.FailedDeactivations,
		"notifications_sent", output.NotificationsSent,
		"duration_seconds", output.ProcessingTimeSeconds)

	return output, nil
}
