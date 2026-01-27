package workflows

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// ReinstatementWorkflowInput is the input for agent reinstatement workflow
// WF-AGT-PRF-011: Reinstatement Workflow
type ReinstatementWorkflowInput struct {
	ReinstatementID     string `json:"reinstatement_id"`     // Created by handler
	AgentID             string `json:"agent_id"`
	ReinstatementReason string `json:"reinstatement_reason"`
	RequestedBy         string `json:"requested_by"`
}

// ReinstatementWorkflowOutput is the output of reinstatement workflow
type ReinstatementWorkflowOutput struct {
	ReinstatementID       string    `json:"reinstatement_id"`
	AgentID               string    `json:"agent_id"`
	ApprovalStatus        string    `json:"approval_status"` // PENDING, APPROVED, REJECTED
	ApprovedBy            string    `json:"approved_by,omitempty"`
	RejectedBy            string    `json:"rejected_by,omitempty"`
	RejectionReason       string    `json:"rejection_reason,omitempty"`
	StatusRestored        bool      `json:"status_restored"`
	CommissionReEnabled   bool      `json:"commission_re_enabled"`
	PortalAccessRestored  bool      `json:"portal_access_restored"`
	NotificationsSent     bool      `json:"notifications_sent"`
	WorkflowCompletedAt   time.Time `json:"workflow_completed_at"`
	ProcessingTimeSeconds int       `json:"processing_time_seconds"`
	Errors                []string  `json:"errors,omitempty"`
}

// ApprovalDecisionSignal is the signal sent to approve/reject reinstatement
type ApprovalDecisionSignal struct {
	Decision      string `json:"decision"` // APPROVED, REJECTED
	DecidedBy     string `json:"decided_by"`
	Reason        string `json:"reason,omitempty"`
	Conditions    string `json:"conditions,omitempty"`
	ProbationDays int    `json:"probation_days,omitempty"`
}

// AgentReinstatementWorkflow handles agent reinstatement with manager approval
// WF-AGT-PRF-011: Reinstatement Workflow
//
// This workflow implements human-in-the-loop approval pattern:
// 1. Request is created (done by handler)
// 2. Notification sent to approver (manager)
// 3. Wait for approval/rejection signal
// 4. On approval:
//    - Restore agent status to ACTIVE
//    - Re-enable commission processing
//    - Restore portal access
//    - Send confirmation notifications
// 5. On rejection:
//    - Update request status
//    - Notify requester with reason
//
// The workflow uses Temporal signals for approval decisions
func AgentReinstatementWorkflow(ctx workflow.Context, input ReinstatementWorkflowInput) (*ReinstatementWorkflowOutput, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting agent reinstatement workflow",
		"agent_id", input.AgentID,
		"reinstatement_id", input.ReinstatementID)

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

	output := &ReinstatementWorkflowOutput{
		ReinstatementID: input.ReinstatementID,
		AgentID:         input.AgentID,
		ApprovalStatus:  "PENDING",
	}

	// Step 1: Send approval request notification to manager
	logger.Info("Step 1: Sending approval request notification")
	err := workflow.ExecuteActivity(ctx, "SendApprovalRequestNotificationActivity", input).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to send approval request", "error", err)
		output.Errors = append(output.Errors, "Approval notification: "+err.Error())
		// Continue anyway - workflow can still be approved manually
	}

	// Step 2: Wait for approval decision signal (human-in-the-loop)
	// Timeout after 30 days - auto-reject if no decision
	logger.Info("Step 2: Waiting for approval decision")

	var decision ApprovalDecisionSignal
	signalChannel := workflow.GetSignalChannel(ctx, "approval-decision")

	selector := workflow.NewSelector(ctx)

	// Wait for approval signal OR timeout after 30 days
	approvalReceived := false

	selector.AddReceive(signalChannel, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &decision)
		approvalReceived = true
		logger.Info("Received approval decision", "decision", decision.Decision, "decided_by", decision.DecidedBy)
	})

	// Add timeout handler (30 days)
	timeoutTimer := workflow.NewTimer(ctx, 30*24*time.Hour)
	selector.AddFuture(timeoutTimer, func(f workflow.Future) {
		logger.Warn("Approval timeout - auto-rejecting reinstatement request")
		decision = ApprovalDecisionSignal{
			Decision:  "REJECTED",
			DecidedBy: "SYSTEM",
			Reason:    "Request timeout - no decision received within 30 days",
		}
	})

	selector.Select(ctx) // Block until either signal or timeout

	// Step 3: Process the decision
	if decision.Decision == "APPROVED" {
		logger.Info("Step 3: Processing approval")
		output.ApprovalStatus = "APPROVED"
		output.ApprovedBy = decision.DecidedBy

		// Activity 1: Approve reinstatement (updates database)
		err = workflow.ExecuteActivity(ctx, "ApproveReinstatementActivity",
			input.ReinstatementID,
			decision.DecidedBy,
			decision.Conditions,
			decision.ProbationDays,
		).Get(ctx, nil)
		if err != nil {
			logger.Error("Failed to approve reinstatement", "error", err)
			output.Errors = append(output.Errors, "Approval: "+err.Error())
			return output, err
		}
		output.StatusRestored = true
		output.CommissionReEnabled = true // Done by ApproveReinstatementActivity

		// Activity 2: Restore portal access
		err = workflow.ExecuteActivity(ctx, "RestorePortalAccessActivity", input.AgentID).Get(ctx, nil)
		if err != nil {
			logger.Error("Failed to restore portal access", "error", err)
			output.Errors = append(output.Errors, "Portal access: "+err.Error())
		} else {
			output.PortalAccessRestored = true
		}

		// Activity 3: Send approval confirmation notifications
		err = workflow.ExecuteActivity(ctx, "SendReinstatementApprovalNotificationActivity", input, decision).Get(ctx, nil)
		if err != nil {
			logger.Error("Failed to send approval notifications", "error", err)
			output.Errors = append(output.Errors, "Notifications: "+err.Error())
		} else {
			output.NotificationsSent = true
		}

	} else {
		// REJECTED
		logger.Info("Step 3: Processing rejection")
		output.ApprovalStatus = "REJECTED"
		output.RejectedBy = decision.DecidedBy
		output.RejectionReason = decision.Reason

		// Activity 1: Reject reinstatement (updates database)
		err = workflow.ExecuteActivity(ctx, "RejectReinstatementActivity",
			input.ReinstatementID,
			decision.DecidedBy,
			decision.Reason,
		).Get(ctx, nil)
		if err != nil {
			logger.Error("Failed to reject reinstatement", "error", err)
			output.Errors = append(output.Errors, "Rejection: "+err.Error())
			return output, err
		}

		// Activity 2: Send rejection notification
		err = workflow.ExecuteActivity(ctx, "SendReinstatementRejectionNotificationActivity", input, decision).Get(ctx, nil)
		if err != nil {
			logger.Error("Failed to send rejection notifications", "error", err)
			output.Errors = append(output.Errors, "Notifications: "+err.Error())
		} else {
			output.NotificationsSent = true
		}
	}

	output.WorkflowCompletedAt = workflow.Now(ctx)
	output.ProcessingTimeSeconds = int(workflow.Now(ctx).Sub(startTime).Seconds())

	logger.Info("Agent reinstatement workflow completed",
		"agent_id", input.AgentID,
		"status", output.ApprovalStatus,
		"duration_seconds", output.ProcessingTimeSeconds)

	return output, nil
}
