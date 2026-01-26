package workflows

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// ========================================================================
// APPROVAL CHILD WORKFLOW - Human-in-the-Loop Pattern
// ========================================================================
// This workflow implements human approval process with timeout
// Used by WF-002: Agent Onboarding Workflow for supervisor/management approvals
// Receives approval decisions via signals

const (
	ApprovalSignalName = "approval-decision"
	ApprovalTimeout    = 72 * time.Hour // 3 days for approval
)

// ApprovalWorkflowInput contains data for approval request
type ApprovalWorkflowInput struct {
	RequestID    string
	AgentType    string
	FirstName    string
	LastName     string
	PANNumber    string
	ApprovalType string // SUPERVISOR_APPROVAL, MANAGEMENT_APPROVAL
	InitiatedBy  string
	Approvers    []string // List of approver emails
}

// ApprovalDecisionSignal contains approval decision from approver
type ApprovalDecisionSignal struct {
	Decision   string    // APPROVED, REJECTED
	ApprovedBy string    // Email/ID of approver
	Comments   string    // Approval comments
	ApprovedAt time.Time // Timestamp of approval
}

// ApprovalWorkflowOutput contains approval result
type ApprovalWorkflowOutput struct {
	Approved   bool
	Decision   string
	ApprovedBy string
	Comments   string
	ApprovedAt time.Time
}

// AgentApprovalWorkflow implements human-in-the-loop approval pattern
// ACT-022: SendApprovalRequestActivity triggers this child workflow
// Waits for approval signal with timeout
// Business Rule: Approval required for ADVISOR and ADVISOR_COORDINATOR types
func AgentApprovalWorkflow(ctx workflow.Context, input ApprovalWorkflowInput) (*ApprovalWorkflowOutput, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("AgentApprovalWorkflow started",
		"RequestID", input.RequestID,
		"AgentType", input.AgentType,
		"ApprovalType", input.ApprovalType)

	// Set up signal channel to receive approval decision
	var approvalDecision ApprovalDecisionSignal
	signalChan := workflow.GetSignalChannel(ctx, ApprovalSignalName)

	// Send approval notification to approvers
	// This activity sends email/notification with approval link
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, ao)

	// Execute SendApprovalNotificationActivity
	// This sends notification to approvers with approval link
	var notificationSent bool
	err := workflow.ExecuteActivity(activityCtx, "SendApprovalNotificationActivity", input).Get(activityCtx, &notificationSent)
	if err != nil {
		logger.Error("Failed to send approval notification", "error", err)
		// Continue anyway - approver might check via dashboard
	}

	logger.Info("Waiting for approval decision", "Timeout", ApprovalTimeout)

	// Wait for approval signal with timeout
	// Selector allows waiting with timeout
	selector := workflow.NewSelector(ctx)

	// Case 1: Approval signal received
	selector.AddReceive(signalChan, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &approvalDecision)
		logger.Info("Approval decision received",
			"Decision", approvalDecision.Decision,
			"ApprovedBy", approvalDecision.ApprovedBy)
	})

	// Case 2: Timeout - auto-reject after 72 hours
	timerFuture := workflow.NewTimer(ctx, ApprovalTimeout)
	selector.AddFuture(timerFuture, func(f workflow.Future) {
		logger.Warn("Approval timeout - auto-rejecting", "Timeout", ApprovalTimeout)
		approvalDecision = ApprovalDecisionSignal{
			Decision:   "TIMEOUT",
			ApprovedBy: "SYSTEM",
			Comments:   "Approval timeout - automatically rejected after 72 hours",
			ApprovedAt: workflow.Now(ctx),
		}
	})

	// Wait for either signal or timeout
	selector.Select(ctx)

	// Process approval decision
	approved := approvalDecision.Decision == "APPROVED"

	logger.Info("Approval workflow completed",
		"Approved", approved,
		"Decision", approvalDecision.Decision,
		"ApprovedBy", approvalDecision.ApprovedBy)

	return &ApprovalWorkflowOutput{
		Approved:   approved,
		Decision:   approvalDecision.Decision,
		ApprovedBy: approvalDecision.ApprovedBy,
		Comments:   approvalDecision.Comments,
		ApprovedAt: approvalDecision.ApprovedAt,
	}, nil
}

// ========================================================================
// HELPER FUNCTIONS FOR APPROVAL WORKFLOW
// ========================================================================

// SendApprovalDecisionSignal sends approval decision to running workflow
// This function is called by approval API endpoint when approver makes decision
// External API: POST /agent-profiles/approvals/:request_id/decide
func SendApprovalDecisionSignal(
	ctx workflow.Context,
	workflowID string,
	decision ApprovalDecisionSignal,
) error {
	// Note: This is a helper function - actual signal sending happens via Temporal client
	// from the approval API handler
	return fmt.Errorf("this function should be called via Temporal client, not from workflow")
}
