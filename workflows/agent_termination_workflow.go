package workflows

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// TerminationWorkflowInput is the input for agent termination workflow
// WF-AGT-PRF-004: Agent Termination Workflow
type TerminationWorkflowInput struct {
	AgentID               string    `json:"agent_id"`
	TerminationReason     string    `json:"termination_reason"`
	TerminationReasonCode string    `json:"termination_reason_code"`
	EffectiveDate         time.Time `json:"effective_date"`
	TerminatedBy          string    `json:"terminated_by"`
	TerminationRecordID   string    `json:"termination_record_id"` // Created by handler
}

// TerminationWorkflowOutput is the output of termination workflow
type TerminationWorkflowOutput struct {
	AgentID                     string    `json:"agent_id"`
	TerminationRecordID         string    `json:"termination_record_id"`
	StatusUpdated               bool      `json:"status_updated"`
	PortalDisabled              bool      `json:"portal_disabled"`
	CommissionStopped           bool      `json:"commission_stopped"`
	LetterGenerated             bool      `json:"letter_generated"`
	LetterURL                   string    `json:"letter_url,omitempty"`
	DataArchived                bool      `json:"data_archived"`
	ArchiveID                   string    `json:"archive_id,omitempty"`
	NotificationsSent           bool      `json:"notifications_sent"`
	WorkflowCompletedAt         time.Time `json:"workflow_completed_at"`
	ProcessingTimeSeconds       int       `json:"processing_time_seconds"`
	Errors                      []string  `json:"errors,omitempty"`
}

// AgentTerminationWorkflow orchestrates the complete agent termination process
// WF-AGT-PRF-004: Agent Termination Workflow
// BR-AGT-PRF-017: Agent Termination Workflow
//
// This workflow handles:
// 1. Status update (done by handler before workflow)
// 2. Disable portal access
// 3. Stop commission processing
// 4. Generate termination letter
// 5. Archive agent data (7-year retention)
// 6. Send notifications to stakeholders
//
// The workflow ensures all termination actions are completed reliably with automatic retries
func AgentTerminationWorkflow(ctx workflow.Context, input TerminationWorkflowInput) (*TerminationWorkflowOutput, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting agent termination workflow",
		"agent_id", input.AgentID,
		"reason_code", input.TerminationReasonCode)

	startTime := workflow.Now(ctx)

	// Set activity options with retries
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		HeartbeatTimeout:    30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	output := &TerminationWorkflowOutput{
		AgentID:             input.AgentID,
		TerminationRecordID: input.TerminationRecordID,
		StatusUpdated:       true, // Already done by handler
	}

	// Activity 1: Disable Portal Access
	logger.Info("Step 1: Disabling portal access")
	err := workflow.ExecuteActivity(ctx, "DisablePortalAccessActivity", input.AgentID).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to disable portal access", "error", err)
		output.Errors = append(output.Errors, "Portal access: "+err.Error())
		// Continue with other activities even if this fails
	} else {
		output.PortalDisabled = true
		// Update termination record
		_ = workflow.ExecuteActivity(ctx, "UpdateTerminationRecordActivity",
			input.TerminationRecordID, map[string]interface{}{"portal_disabled": true}).Get(ctx, nil)
	}

	// Activity 2: Stop Commission Processing
	logger.Info("Step 2: Stopping commission processing")
	err = workflow.ExecuteActivity(ctx, "StopCommissionProcessingActivity", input.AgentID).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to stop commission", "error", err)
		output.Errors = append(output.Errors, "Commission: "+err.Error())
	} else {
		output.CommissionStopped = true
		_ = workflow.ExecuteActivity(ctx, "UpdateTerminationRecordActivity",
			input.TerminationRecordID, map[string]interface{}{"commission_stopped": true}).Get(ctx, nil)
	}

	// Activity 3: Generate Termination Letter
	logger.Info("Step 3: Generating termination letter")
	var letterURL string
	err = workflow.ExecuteActivity(ctx, "GenerateTerminationLetterActivity", input).Get(ctx, &letterURL)
	if err != nil {
		logger.Error("Failed to generate termination letter", "error", err)
		output.Errors = append(output.Errors, "Letter generation: "+err.Error())
	} else {
		output.LetterGenerated = true
		output.LetterURL = letterURL
		_ = workflow.ExecuteActivity(ctx, "UpdateTerminationRecordActivity",
			input.TerminationRecordID, map[string]interface{}{
				"letter_generated":             true,
				"termination_letter_url":       letterURL,
				"termination_letter_generated_at": workflow.Now(ctx),
			}).Get(ctx, nil)
	}

	// Activity 4: Archive Agent Data (7-year retention)
	logger.Info("Step 4: Archiving agent data")
	var archiveID string
	err = workflow.ExecuteActivity(ctx, "ArchiveAgentDataActivity", input.AgentID).Get(ctx, &archiveID)
	if err != nil {
		logger.Error("Failed to archive agent data", "error", err)
		output.Errors = append(output.Errors, "Data archival: "+err.Error())
	} else {
		output.DataArchived = true
		output.ArchiveID = archiveID
		_ = workflow.ExecuteActivity(ctx, "UpdateTerminationRecordActivity",
			input.TerminationRecordID, map[string]interface{}{"data_archived": true}).Get(ctx, nil)
	}

	// Activity 5: Send Notifications
	logger.Info("Step 5: Sending termination notifications")
	err = workflow.ExecuteActivity(ctx, "SendTerminationNotificationsActivity", input).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to send notifications", "error", err)
		output.Errors = append(output.Errors, "Notifications: "+err.Error())
	} else {
		output.NotificationsSent = true
		_ = workflow.ExecuteActivity(ctx, "UpdateTerminationRecordActivity",
			input.TerminationRecordID, map[string]interface{}{"notifications_sent": true}).Get(ctx, nil)
	}

	// Final workflow status update
	_ = workflow.ExecuteActivity(ctx, "UpdateTerminationRecordActivity",
		input.TerminationRecordID, map[string]interface{}{"workflow_status": "COMPLETED"}).Get(ctx, nil)

	output.WorkflowCompletedAt = workflow.Now(ctx)
	output.ProcessingTimeSeconds = int(workflow.Now(ctx).Sub(startTime).Seconds())

	logger.Info("Agent termination workflow completed",
		"agent_id", input.AgentID,
		"duration_seconds", output.ProcessingTimeSeconds,
		"portal_disabled", output.PortalDisabled,
		"letter_generated", output.LetterGenerated,
		"data_archived", output.DataArchived,
		"errors", len(output.Errors))

	return output, nil
}
