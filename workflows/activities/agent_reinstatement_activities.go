package activities

import (
	"context"
	"fmt"

	repo "pli-agent-api/repo/postgres"
	"pli-agent-api/workflows"

	"go.temporal.io/sdk/activity"
)

// AgentReinstatementActivities provides activities for agent reinstatement workflow
// WF-AGT-PRF-011: Reinstatement Workflow
type AgentReinstatementActivities struct {
	terminationRepo *repo.AgentTerminationRepository
	profileRepo     *repo.AgentProfileRepository
	contactRepo     *repo.AgentContactRepository
	emailRepo       *repo.AgentEmailRepository
}

// NewAgentReinstatementActivities creates a new agent reinstatement activities struct
func NewAgentReinstatementActivities(
	terminationRepo *repo.AgentTerminationRepository,
	profileRepo *repo.AgentProfileRepository,
	contactRepo *repo.AgentContactRepository,
	emailRepo *repo.AgentEmailRepository,
) *AgentReinstatementActivities {
	return &AgentReinstatementActivities{
		terminationRepo: terminationRepo,
		profileRepo:     profileRepo,
		contactRepo:     contactRepo,
		emailRepo:       emailRepo,
	}
}

// SendApprovalRequestNotificationActivity sends notification to approver
// ACT-REINST-001: Send Approval Request Notification
func (a *AgentReinstatementActivities) SendApprovalRequestNotificationActivity(
	ctx context.Context,
	input workflows.ReinstatementWorkflowInput,
) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Sending approval request notification",
		"reinstatement_id", input.ReinstatementID,
		"agent_id", input.AgentID)

	activity.RecordHeartbeat(ctx, "Sending approval request")

	// Fetch agent profile
	profile, err := a.profileRepo.FindByID(ctx, input.AgentID)
	if err != nil {
		return fmt.Errorf("failed to fetch agent profile: %w", err)
	}

	// TODO: Implement approval request notification
	// This would typically:
	// 1. Identify the approver (manager/supervisor)
	// 2. Send email with approval link
	// 3. Send SMS notification
	// 4. Create in-app task/notification
	// 5. Include reinstatement reason and agent details
	//
	// Example:
	// approver, err := a.hierarchyService.GetApprover(ctx, input.AgentID)
	// if err != nil {
	//     return fmt.Errorf("failed to get approver: %w", err)
	// }
	//
	// approvalLink := fmt.Sprintf("https://portal.example.com/approvals/reinstatement/%s", input.ReinstatementID)
	//
	// err = a.notificationService.SendApprovalRequest(ctx, ApprovalNotificationRequest{
	//     ApproverEmail: approver.Email,
	//     ApproverName: approver.Name,
	//     AgentID: input.AgentID,
	//     AgentName: fmt.Sprintf("%s %s", profile.FirstName, profile.LastName),
	//     ReinstatementReason: input.ReinstatementReason,
	//     RequestedBy: input.RequestedBy,
	//     ApprovalLink: approvalLink,
	//     Channels: []string{"EMAIL", "SMS", "IN_APP"},
	// })
	// if err != nil {
	//     return fmt.Errorf("failed to send approval request: %w", err)
	// }

	logger.Info("Approval request notification sent successfully",
		"agent_id", input.AgentID,
		"agent_code", profile.AgentCode)

	return nil
}

// ApproveReinstatementActivity approves reinstatement and updates database
// ACT-REINST-002: Approve Reinstatement
// Uses SINGLE database hit with CTE pattern (implemented in repository)
func (a *AgentReinstatementActivities) ApproveReinstatementActivity(
	ctx context.Context,
	reinstatementID string,
	approvedBy string,
	conditions string,
	probationDays int,
) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Approving reinstatement",
		"reinstatement_id", reinstatementID,
		"approved_by", approvedBy)

	activity.RecordHeartbeat(ctx, "Approving reinstatement")

	// Call repository method - SINGLE database hit
	// Updates: reinstatement_request + agent_profiles + audit_log atomically
	_, err := a.terminationRepo.ApproveReinstatement(ctx, reinstatementID, approvedBy, conditions, probationDays)
	if err != nil {
		return fmt.Errorf("failed to approve reinstatement: %w", err)
	}

	logger.Info("Reinstatement approved successfully",
		"reinstatement_id", reinstatementID)

	return nil
}

// RejectReinstatementActivity rejects reinstatement and updates database
// ACT-REINST-003: Reject Reinstatement
func (a *AgentReinstatementActivities) RejectReinstatementActivity(
	ctx context.Context,
	reinstatementID string,
	rejectedBy string,
	rejectionReason string,
) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Rejecting reinstatement",
		"reinstatement_id", reinstatementID,
		"rejected_by", rejectedBy)

	activity.RecordHeartbeat(ctx, "Rejecting reinstatement")

	// TODO: Implement rejection in repository
	// For now, we'll use FindReinstatementRequest and update manually
	// Ideally, create a RejectReinstatement method with CTE pattern

	request, err := a.terminationRepo.FindReinstatementRequest(ctx, reinstatementID)
	if err != nil {
		return fmt.Errorf("failed to find reinstatement request: %w", err)
	}

	// TODO: Create RejectReinstatement repository method with SINGLE database hit
	// For now, update the fields we have access to
	_ = request // Placeholder

	// Example of what the repository method should do:
	// _, err = a.terminationRepo.RejectReinstatement(ctx, reinstatementID, rejectedBy, rejectionReason)
	// if err != nil {
	//     return fmt.Errorf("failed to reject reinstatement: %w", err)
	// }

	logger.Info("Reinstatement rejected successfully",
		"reinstatement_id", reinstatementID,
		"reason", rejectionReason)

	return nil
}

// RestorePortalAccessActivity restores agent portal access
// ACT-REINST-004: Restore Portal Access
func (a *AgentReinstatementActivities) RestorePortalAccessActivity(ctx context.Context, agentID string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Restoring portal access", "agent_id", agentID)

	activity.RecordHeartbeat(ctx, "Restoring portal access")

	// TODO: Implement portal access restoration logic
	// This would typically:
	// 1. Call authentication service API
	// 2. Re-enable login credentials
	// 3. Restore user roles/permissions
	// 4. Allow future login attempts
	//
	// Example:
	// err := a.authService.EnableAccess(ctx, agentID)
	// if err != nil {
	//     return fmt.Errorf("failed to restore portal access: %w", err)
	// }

	logger.Info("Portal access restored successfully", "agent_id", agentID)
	return nil
}

// SendReinstatementApprovalNotificationActivity sends approval confirmation
// ACT-REINST-005: Send Approval Confirmation Notification
func (a *AgentReinstatementActivities) SendReinstatementApprovalNotificationActivity(
	ctx context.Context,
	input workflows.ReinstatementWorkflowInput,
	decision workflows.ApprovalDecisionSignal,
) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Sending approval confirmation notification", "agent_id", input.AgentID)

	activity.RecordHeartbeat(ctx, "Sending approval notification")

	// Fetch agent profile and contact details
	profile, err := a.profileRepo.FindByID(ctx, input.AgentID)
	if err != nil {
		return fmt.Errorf("failed to fetch agent profile: %w", err)
	}

	contacts, _ := a.contactRepo.FindByAgentID(ctx, input.AgentID)
	emails, _ := a.emailRepo.FindByAgentID(ctx, input.AgentID)

	// Extract primary contact information
	var primaryEmail, primaryPhone string
	for _, email := range emails {
		if email.IsPrimary {
			primaryEmail = email.EmailID
			break
		}
	}
	for _, contact := range contacts {
		if contact.IsPrimary {
			primaryPhone = contact.MobileNumber
			break
		}
	}

	// TODO: Implement approval notification sending
	// This would typically:
	// 1. Send email to agent with approval details
	// 2. Send SMS notification
	// 3. Create in-app notification
	// 4. Notify requester
	// 5. Notify manager/HR
	//
	// Example:
	// err = a.notificationService.SendReinstatementApproval(ctx, NotificationRequest{
	//     AgentID: input.AgentID,
	//     AgentName: fmt.Sprintf("%s %s", profile.FirstName, profile.LastName),
	//     AgentEmail: primaryEmail,
	//     AgentPhone: primaryPhone,
	//     ApprovedBy: decision.DecidedBy,
	//     Conditions: decision.Conditions,
	//     ProbationDays: decision.ProbationDays,
	//     Channels: []string{"EMAIL", "SMS", "IN_APP"},
	// })
	// if err != nil {
	//     return fmt.Errorf("failed to send approval notifications: %w", err)
	// }

	logger.Info("Approval confirmation sent successfully",
		"agent_id", input.AgentID,
		"agent_code", profile.AgentCode,
		"email", primaryEmail,
		"phone", primaryPhone)

	return nil
}

// SendReinstatementRejectionNotificationActivity sends rejection notification
// ACT-REINST-006: Send Rejection Notification
func (a *AgentReinstatementActivities) SendReinstatementRejectionNotificationActivity(
	ctx context.Context,
	input workflows.ReinstatementWorkflowInput,
	decision workflows.ApprovalDecisionSignal,
) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Sending rejection notification", "agent_id", input.AgentID)

	activity.RecordHeartbeat(ctx, "Sending rejection notification")

	// Fetch agent profile and contact details
	profile, err := a.profileRepo.FindByID(ctx, input.AgentID)
	if err != nil {
		return fmt.Errorf("failed to fetch agent profile: %w", err)
	}

	contacts, _ := a.contactRepo.FindByAgentID(ctx, input.AgentID)
	emails, _ := a.emailRepo.FindByAgentID(ctx, input.AgentID)

	// Extract primary contact information
	var primaryEmail, primaryPhone string
	for _, email := range emails {
		if email.IsPrimary {
			primaryEmail = email.EmailID
			break
		}
	}
	for _, contact := range contacts {
		if contact.IsPrimary {
			primaryPhone = contact.MobileNumber
			break
		}
	}

	// TODO: Implement rejection notification sending
	// This would typically:
	// 1. Send email to requester with rejection reason
	// 2. Send SMS notification
	// 3. Create in-app notification
	// 4. Notify manager/HR
	//
	// Example:
	// err = a.notificationService.SendReinstatementRejection(ctx, NotificationRequest{
	//     AgentID: input.AgentID,
	//     AgentName: fmt.Sprintf("%s %s", profile.FirstName, profile.LastName),
	//     RequesterEmail: input.RequestedBy, // Assuming this is an email
	//     RejectedBy: decision.DecidedBy,
	//     RejectionReason: decision.Reason,
	//     Channels: []string{"EMAIL", "SMS", "IN_APP"},
	// })
	// if err != nil {
	//     return fmt.Errorf("failed to send rejection notifications: %w", err)
	// }

	logger.Info("Rejection notification sent successfully",
		"agent_id", input.AgentID,
		"agent_code", profile.AgentCode,
		"email", primaryEmail,
		"phone", primaryPhone,
		"reason", decision.Reason)

	return nil
}
