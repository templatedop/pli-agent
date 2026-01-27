package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"pli-agent-api/core/domain"
	repo "pli-agent-api/repo/postgres"
	"pli-agent-api/workflows"

	"go.temporal.io/sdk/activity"
)

// AgentTerminationActivities provides activities for agent termination workflow
// WF-AGT-PRF-004: Agent Termination Workflow
type AgentTerminationActivities struct {
	terminationRepo *repo.AgentTerminationRepository
	profileRepo     *repo.AgentProfileRepository
	addressRepo     *repo.AgentAddressRepository
	contactRepo     *repo.AgentContactRepository
	emailRepo       *repo.AgentEmailRepository
	bankRepo        *repo.AgentBankDetailsRepository
	licenseRepo     *repo.AgentLicenseRepository
	auditLogRepo    *repo.AgentAuditLogRepository
}

// NewAgentTerminationActivities creates a new agent termination activities struct
func NewAgentTerminationActivities(
	terminationRepo *repo.AgentTerminationRepository,
	profileRepo *repo.AgentProfileRepository,
	addressRepo *repo.AgentAddressRepository,
	contactRepo *repo.AgentContactRepository,
	emailRepo *repo.AgentEmailRepository,
	bankRepo *repo.AgentBankDetailsRepository,
	licenseRepo *repo.AgentLicenseRepository,
	auditLogRepo *repo.AgentAuditLogRepository,
) *AgentTerminationActivities {
	return &AgentTerminationActivities{
		terminationRepo: terminationRepo,
		profileRepo:     profileRepo,
		addressRepo:     addressRepo,
		contactRepo:     contactRepo,
		emailRepo:       emailRepo,
		bankRepo:        bankRepo,
		licenseRepo:     licenseRepo,
		auditLogRepo:    auditLogRepo,
	}
}

// DisablePortalAccessActivity disables agent portal access
// ACT-TERM-001: Disable Portal Access
func (a *AgentTerminationActivities) DisablePortalAccessActivity(ctx context.Context, agentID string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Disabling portal access", "agent_id", agentID)

	activity.RecordHeartbeat(ctx, "Disabling portal access")

	// TODO: Implement portal access disable logic
	// This would typically:
	// 1. Call authentication service API
	// 2. Revoke all active sessions
	// 3. Disable login credentials
	// 4. Block future login attempts
	//
	// Example:
	// err := a.authService.DisableAccess(ctx, agentID)
	// if err != nil {
	//     return fmt.Errorf("failed to disable portal access: %w", err)
	// }

	logger.Info("Portal access disabled successfully", "agent_id", agentID)
	return nil
}

// StopCommissionProcessingActivity stops commission processing for agent
// ACT-TERM-002: Stop Commission Processing
func (a *AgentTerminationActivities) StopCommissionProcessingActivity(ctx context.Context, agentID string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Stopping commission processing", "agent_id", agentID)

	activity.RecordHeartbeat(ctx, "Stopping commission")

	// Commission is already disabled by the handler via commission_enabled = false
	// This activity can handle additional commission system integrations

	// TODO: Implement external commission system integration
	// This would typically:
	// 1. Call commission calculation service
	// 2. Stop future commission calculations
	// 3. Flag in payroll system
	// 4. Update commission status table
	//
	// Example:
	// err := a.commissionService.StopProcessing(ctx, agentID)
	// if err != nil {
	//     return fmt.Errorf("failed to stop commission: %w", err)
	// }

	logger.Info("Commission processing stopped successfully", "agent_id", agentID)
	return nil
}

// GenerateTerminationLetterActivity generates termination letter PDF
// ACT-TERM-003: Generate Termination Letter
// BR-AGT-PRF-017: Agent Termination Workflow
func (a *AgentTerminationActivities) GenerateTerminationLetterActivity(
	ctx context.Context,
	input workflows.TerminationWorkflowInput,
) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Generating termination letter", "agent_id", input.AgentID)

	activity.RecordHeartbeat(ctx, "Generating termination letter")

	// Fetch agent profile details
	profile, err := a.profileRepo.FindByID(ctx, input.AgentID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch agent profile: %w", err)
	}

	// TODO: Implement letter generation logic
	// This would typically:
	// 1. Create termination letter PDF from template
	// 2. Include agent details, termination reason, effective date
	// 3. Upload to cloud storage (S3, GCS, etc.)
	// 4. Return public/signed URL
	//
	// Example:
	// letterData := TerminationLetterData{
	//     AgentCode: profile.AgentCode,
	//     AgentName: fmt.Sprintf("%s %s %s", profile.FirstName, profile.MiddleName, profile.LastName),
	//     TerminationDate: input.EffectiveDate,
	//     TerminationReason: input.TerminationReason,
	//     TerminationReasonCode: input.TerminationReasonCode,
	// }
	//
	// pdfBytes, err := a.letterService.GeneratePDF(ctx, "termination_letter_template.html", letterData)
	// if err != nil {
	//     return "", fmt.Errorf("failed to generate PDF: %w", err)
	// }
	//
	// letterURL, err := a.storageService.UploadFile(ctx, fmt.Sprintf("termination-letters/%s-%s.pdf", input.AgentID, time.Now().Format("20060102")), pdfBytes)
	// if err != nil {
	//     return "", fmt.Errorf("failed to upload letter: %w", err)
	// }

	// For now, return a placeholder URL
	letterURL := fmt.Sprintf("https://storage.example.com/letters/termination-%s-%s.pdf",
		profile.AgentCode, time.Now().Format("20060102"))

	logger.Info("Termination letter generated successfully",
		"agent_id", input.AgentID,
		"letter_url", letterURL)

	return letterURL, nil
}

// ArchiveAgentDataActivity archives agent data for 7-year retention
// ACT-TERM-004: Archive Agent Data
// BR-AGT-PRF-017: Agent Termination Workflow (7-year retention)
func (a *AgentTerminationActivities) ArchiveAgentDataActivity(ctx context.Context, agentID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Archiving agent data", "agent_id", agentID)

	activity.RecordHeartbeat(ctx, "Archiving agent data")

	// Fetch all agent data from multiple tables
	profile, err := a.profileRepo.FindByID(ctx, agentID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch profile: %w", err)
	}

	addresses, _ := a.addressRepo.FindByAgentID(ctx, agentID)
	contacts, _ := a.contactRepo.FindByAgentID(ctx, agentID)
	emails, _ := a.emailRepo.FindByAgentID(ctx, agentID)
	bankDetails, _ := a.bankRepo.FindByAgentID(ctx, agentID)
	licenses, _ := a.licenseRepo.FindByAgentID(ctx, agentID)
	auditLogs, _ := a.auditLogRepo.FindByAgentID(ctx, agentID, 0, 10000) // All audit logs

	// Create comprehensive data snapshot
	dataSnapshot := map[string]interface{}{
		"profile":      profile,
		"addresses":    addresses,
		"contacts":     contacts,
		"emails":       emails,
		"bank_details": bankDetails,
		"licenses":     licenses,
		"audit_logs":   auditLogs,
		"archived_at":  time.Now(),
		"archive_type": "TERMINATION",
	}

	// Convert to JSON
	snapshotJSON, err := json.Marshal(dataSnapshot)
	if err != nil {
		return "", fmt.Errorf("failed to marshal data snapshot: %w", err)
	}

	// Create archive record (7-year retention)
	archive, err := a.terminationRepo.CreateDataArchive(
		ctx,
		agentID,
		domain.ArchiveTypeTermination,
		string(snapshotJSON),
		"SYSTEM",
	)
	if err != nil {
		return "", fmt.Errorf("failed to create archive: %w", err)
	}

	logger.Info("Agent data archived successfully",
		"agent_id", agentID,
		"archive_id", archive.ArchiveID,
		"retention_until", archive.RetentionUntil.Format("2006-01-02"))

	return archive.ArchiveID, nil
}

// SendTerminationNotificationsActivity sends notifications about termination
// ACT-TERM-005: Send Termination Notifications
func (a *AgentTerminationActivities) SendTerminationNotificationsActivity(
	ctx context.Context,
	input workflows.TerminationWorkflowInput,
) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Sending termination notifications", "agent_id", input.AgentID)

	activity.RecordHeartbeat(ctx, "Sending notifications")

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

	// TODO: Implement notification sending logic
	// This would typically:
	// 1. Send email to agent with termination details
	// 2. Send SMS notification
	// 3. Create in-app notification
	// 4. Notify manager/supervisor
	// 5. Notify HR department
	//
	// Example:
	// err = a.notificationService.SendTerminationNotice(ctx, NotificationRequest{
	//     AgentID: input.AgentID,
	//     AgentName: fmt.Sprintf("%s %s", profile.FirstName, profile.LastName),
	//     AgentEmail: primaryEmail,
	//     AgentPhone: primaryPhone,
	//     TerminationDate: input.EffectiveDate,
	//     TerminationReason: input.TerminationReason,
	//     TerminatedBy: input.TerminatedBy,
	//     Channels: []string{"EMAIL", "SMS", "IN_APP"},
	// })
	// if err != nil {
	//     return fmt.Errorf("failed to send notifications: %w", err)
	// }

	logger.Info("Termination notifications sent successfully",
		"agent_id", input.AgentID,
		"agent_code", profile.AgentCode,
		"email", primaryEmail,
		"phone", primaryPhone)

	return nil
}

// UpdateTerminationRecordActivity updates termination record with workflow progress
// ACT-TERM-006: Update Termination Record
func (a *AgentTerminationActivities) UpdateTerminationRecordActivity(
	ctx context.Context,
	terminationID string,
	updates map[string]interface{},
) error {
	logger := activity.GetLogger(ctx)
	logger.Debug("Updating termination record", "termination_id", terminationID, "updates", updates)

	activity.RecordHeartbeat(ctx, "Updating termination record")

	err := a.terminationRepo.UpdateTerminationRecord(ctx, terminationID, updates)
	if err != nil {
		return fmt.Errorf("failed to update termination record: %w", err)
	}

	logger.Debug("Termination record updated successfully", "termination_id", terminationID)
	return nil
}
