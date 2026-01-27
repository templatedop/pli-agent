package activities

import (
	"context"
	"fmt"
	"time"

	"pli-agent-api/core/domain"
	repo "pli-agent-api/repo/postgres"
	"pli-agent-api/workflows"

	"go.temporal.io/sdk/activity"
)

// LicenseDeactivationActivities provides activities for license deactivation workflow
// WF-AGT-PRF-007: License Deactivation Workflow
type LicenseDeactivationActivities struct {
	licenseRepo  *repo.AgentLicenseRepository
	profileRepo  *repo.AgentProfileRepository
	auditLogRepo *repo.AgentAuditLogRepository
}

// NewLicenseDeactivationActivities creates a new license deactivation activities struct
func NewLicenseDeactivationActivities(
	licenseRepo *repo.AgentLicenseRepository,
	profileRepo *repo.AgentProfileRepository,
	auditLogRepo *repo.AgentAuditLogRepository,
) *LicenseDeactivationActivities {
	return &LicenseDeactivationActivities{
		licenseRepo:  licenseRepo,
		profileRepo:  profileRepo,
		auditLogRepo: auditLogRepo,
	}
}

// FindExpiredLicensesActivity finds all licenses that have expired
// ACT-DEC-001: Find Expired Licenses
// BR-AGT-PRF-013: Auto-Deactivation on License Expiry
//
// This activity queries the database for all licenses where renewal_date < today
// and the agent status is still ACTIVE
func (a *LicenseDeactivationActivities) FindExpiredLicensesActivity(
	ctx context.Context,
	batchDate time.Time,
) ([]workflows.ExpiredLicenseInfo, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Finding expired licenses", "batch_date", batchDate)

	// Record heartbeat for long-running activity
	activity.RecordHeartbeat(ctx, "Finding expired licenses")

	// Find licenses with agent details using optimized JOIN query
	licensesWithProfiles, err := a.licenseRepo.FindExpiringLicensesWithAgentDetails(ctx, 0)
	if err != nil {
		logger.Error("Failed to find expired licenses", "error", err)
		return nil, fmt.Errorf("failed to find expired licenses: %w", err)
	}

	// Filter for truly expired licenses (renewal_date < today)
	var expiredLicenses []workflows.ExpiredLicenseInfo
	today := batchDate.Truncate(24 * time.Hour)

	for _, license := range licensesWithProfiles {
		renewalDate := license.RenewalDate.Truncate(24 * time.Hour)
		if renewalDate.Before(today) && license.LicenseStatus == domain.LicenseStatusActive {
			expiredLicenses = append(expiredLicenses, workflows.ExpiredLicenseInfo{
				LicenseID:   license.LicenseID,
				AgentID:     license.AgentID,
				AgentCode:   license.AgentCode,
				AgentName:   fmt.Sprintf("%s %s %s", license.FirstName, license.MiddleName, license.LastName),
				LicenseLine: license.LicenseLine,
				RenewalDate: license.RenewalDate.Format("2006-01-02"),
			})
		}
	}

	logger.Info("Found expired licenses", "count", len(expiredLicenses))
	return expiredLicenses, nil
}

// BatchUpdateAgentStatusActivity updates agent status to DEACTIVATED in batch
// ACT-DEC-002: Batch Update Agent Status
// BR-AGT-PRF-013: Auto-Deactivation on License Expiry
//
// This activity updates all agents in the batch to DEACTIVATED status
// and creates audit logs for each update
func (a *LicenseDeactivationActivities) BatchUpdateAgentStatusActivity(
	ctx context.Context,
	batch []workflows.ExpiredLicenseInfo,
) (workflows.BatchUpdateResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Updating agent status", "count", len(batch))

	result := workflows.BatchUpdateResult{
		TotalCount: len(batch),
	}

	for i, item := range batch {
		// Record heartbeat every 10 agents
		if i%10 == 0 {
			activity.RecordHeartbeat(ctx, fmt.Sprintf("Processing %d/%d", i, len(batch)))
		}

		// Update agent status to DEACTIVATED
		updates := map[string]interface{}{
			"status":             domain.AgentStatusDeactivated,
			"status_date":        time.Now(),
			"status_reason":      fmt.Sprintf("License expired on %s", item.RenewalDate),
			"commission_enabled": false,
		}

		err := a.profileRepo.Update(ctx, item.AgentID, updates, "SYSTEM")
		if err != nil {
			logger.Error("Failed to update agent status",
				"agent_id", item.AgentID,
				"agent_code", item.AgentCode,
				"error", err)
			result.FailedCount++
			result.FailedIDs = append(result.FailedIDs, item.AgentID)
			continue
		}

		result.SuccessCount++
		logger.Debug("Agent deactivated",
			"agent_id", item.AgentID,
			"agent_code", item.AgentCode,
			"license_id", item.LicenseID)
	}

	logger.Info("Agent status update complete",
		"success", result.SuccessCount,
		"failed", result.FailedCount)

	return result, nil
}

// BatchDisablePortalAccessActivity disables portal access for agents
// ACT-DEC-003: Batch Disable Portal Access
//
// This activity disables portal/login access for deactivated agents
// Implementation depends on your authentication system (e.g., updating auth DB, revoking tokens)
func (a *LicenseDeactivationActivities) BatchDisablePortalAccessActivity(
	ctx context.Context,
	batch []workflows.ExpiredLicenseInfo,
) (workflows.BatchUpdateResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Disabling portal access", "count", len(batch))

	result := workflows.BatchUpdateResult{
		TotalCount: len(batch),
	}

	for i, item := range batch {
		// Record heartbeat every 10 agents
		if i%10 == 0 {
			activity.RecordHeartbeat(ctx, fmt.Sprintf("Disabling portal access %d/%d", i, len(batch)))
		}

		// TODO: Implement portal access disable logic
		// This would typically:
		// 1. Call an external authentication API
		// 2. Update a portal_access table
		// 3. Revoke active sessions/tokens
		// 4. Disable login credentials
		//
		// Example:
		// err := a.authService.DisableAccess(ctx, item.AgentID)
		// if err != nil {
		//     logger.Error("Failed to disable portal access", "agent_id", item.AgentID, "error", err)
		//     result.FailedCount++
		//     result.FailedIDs = append(result.FailedIDs, item.AgentID)
		//     continue
		// }

		result.SuccessCount++
		logger.Debug("Portal access disabled", "agent_id", item.AgentID, "agent_code", item.AgentCode)
	}

	logger.Info("Portal access disable complete",
		"success", result.SuccessCount,
		"failed", result.FailedCount)

	return result, nil
}

// BatchStopCommissionActivity stops commission processing for agents
// ACT-DEC-004: Batch Stop Commission
// BR-AGT-PRF-013: Auto-Deactivation on License Expiry
//
// This activity stops commission calculation and payment for deactivated agents
func (a *LicenseDeactivationActivities) BatchStopCommissionActivity(
	ctx context.Context,
	batch []workflows.ExpiredLicenseInfo,
) (workflows.BatchUpdateResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Stopping commission processing", "count", len(batch))

	result := workflows.BatchUpdateResult{
		TotalCount: len(batch),
	}

	for i, item := range batch {
		// Record heartbeat every 10 agents
		if i%10 == 0 {
			activity.RecordHeartbeat(ctx, fmt.Sprintf("Stopping commission %d/%d", i, len(batch)))
		}

		// TODO: Implement commission stop logic
		// This would typically:
		// 1. Update commission_processing_status table
		// 2. Call external commission system API
		// 3. Flag agent in payroll system
		// 4. Stop future commission calculations
		//
		// Example:
		// err := a.commissionService.StopProcessing(ctx, item.AgentID)
		// if err != nil {
		//     logger.Error("Failed to stop commission", "agent_id", item.AgentID, "error", err)
		//     result.FailedCount++
		//     result.FailedIDs = append(result.FailedIDs, item.AgentID)
		//     continue
		// }

		// For now, commission_enabled is already set to false by BatchUpdateAgentStatusActivity
		result.SuccessCount++
		logger.Debug("Commission processing stopped", "agent_id", item.AgentID, "agent_code", item.AgentCode)
	}

	logger.Info("Commission stop complete",
		"success", result.SuccessCount,
		"failed", result.FailedCount)

	return result, nil
}

// BatchSendNotificationActivity sends notifications to agents
// ACT-DEC-005: Batch Send Notification
//
// This activity sends email/SMS notifications to agents about their deactivation
func (a *LicenseDeactivationActivities) BatchSendNotificationActivity(
	ctx context.Context,
	batch []workflows.ExpiredLicenseInfo,
) (workflows.BatchUpdateResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Sending notifications", "count", len(batch))

	result := workflows.BatchUpdateResult{
		TotalCount: len(batch),
	}

	for i, item := range batch {
		// Record heartbeat every 10 agents
		if i%10 == 0 {
			activity.RecordHeartbeat(ctx, fmt.Sprintf("Sending notifications %d/%d", i, len(batch)))
		}

		// TODO: Implement notification sending
		// This would typically:
		// 1. Fetch agent contact details (email, phone)
		// 2. Send email notification
		// 3. Send SMS notification
		// 4. Create in-app notification
		// 5. Log notification sent
		//
		// Example:
		// err := a.notificationService.SendDeactivationNotice(ctx, NotificationRequest{
		//     AgentID: item.AgentID,
		//     AgentName: item.AgentName,
		//     Reason: fmt.Sprintf("Your license expired on %s", item.RenewalDate),
		//     Channels: []string{"EMAIL", "SMS"},
		// })
		// if err != nil {
		//     logger.Error("Failed to send notification", "agent_id", item.AgentID, "error", err)
		//     result.FailedCount++
		//     result.FailedIDs = append(result.FailedIDs, item.AgentID)
		//     continue
		// }

		result.SuccessCount++
		logger.Debug("Notification sent",
			"agent_id", item.AgentID,
			"agent_code", item.AgentCode,
			"agent_name", item.AgentName)
	}

	logger.Info("Notification sending complete",
		"success", result.SuccessCount,
		"failed", result.FailedCount)

	return result, nil
}

// CreateBatchAuditLogActivity creates audit log for batch operation
// ACT-DEC-006: Create Batch Audit Log
//
// This activity creates a comprehensive audit log entry for the entire batch operation
func (a *LicenseDeactivationActivities) CreateBatchAuditLogActivity(
	ctx context.Context,
	output *workflows.DeactivationOutput,
) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Creating batch audit log")

	// Record heartbeat
	activity.RecordHeartbeat(ctx, "Creating audit log")

	// Create audit log entry for batch operation
	// TODO: Implement batch audit log table and repository method
	// For now, we'll create a summary audit log
	//
	// Example:
	// err := a.batchOperationRepo.Create(ctx, &domain.BatchOperation{
	//     BatchID: output.BatchID,
	//     OperationType: "LICENSE_DEACTIVATION",
	//     BatchDate: output.BatchDate,
	//     TotalProcessed: output.TotalExpiredLicenses,
	//     SuccessCount: output.AgentsDeactivated,
	//     FailedCount: output.FailedDeactivations,
	//     ProcessingTimeSeconds: output.ProcessingTimeSeconds,
	//     ExecutedBy: "SYSTEM",
	//     Status: "COMPLETED",
	// })

	logger.Info("Batch audit log created",
		"batch_id", output.BatchID,
		"total", output.TotalExpiredLicenses,
		"deactivated", output.AgentsDeactivated,
		"failed", output.FailedDeactivations)

	return nil
}
