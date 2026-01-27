package handler

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"pli-agent-api/core/domain"
	"pli-agent-api/core/port"
	resp "pli-agent-api/handler/response"
	repo "pli-agent-api/repo/postgres"

	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"
)

// LicenseManagementHandler handles license management APIs
// AGT-029 to AGT-038: License Management Journey
type LicenseManagementHandler struct {
	*serverHandler.Base
	licenseRepo  *repo.AgentLicenseRepository
	reminderRepo *repo.LicenseReminderRepository
	profileRepo  *repo.AgentProfileRepository
}

// NewLicenseManagementHandler creates a new license management handler
func NewLicenseManagementHandler(
	licenseRepo *repo.AgentLicenseRepository,
	reminderRepo *repo.LicenseReminderRepository,
	profileRepo *repo.AgentProfileRepository,
) *LicenseManagementHandler {
	base := serverHandler.New("License Management APIs").SetPrefix("/v1").AddPrefix("")
	return &LicenseManagementHandler{
		Base:         base,
		licenseRepo:  licenseRepo,
		reminderRepo: reminderRepo,
		profileRepo:  profileRepo,
	}
}

// Routes defines license management routes
func (h *LicenseManagementHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		// License CRUD
		serverRoute.GET("/agents/:agent_id/licenses", h.GetAgentLicenses).Name("Get Agent Licenses"),
		serverRoute.POST("/agents/:agent_id/licenses", h.AddLicense).Name("Add License"),
		serverRoute.GET("/agents/:agent_id/licenses/:license_id", h.GetLicenseDetails).Name("Get License Details"),
		serverRoute.PUT("/agents/:agent_id/licenses/:license_id", h.UpdateLicense).Name("Update License"),
		serverRoute.PUT("/agents/:agent_id/licenses/:license_id/renew", h.RenewLicense).Name("Renew License"),
		serverRoute.DELETE("/agents/:agent_id/licenses/:license_id", h.DeleteLicense).Name("Delete License"),

		// Lookups and Reports
		serverRoute.GET("/license-types", h.GetLicenseTypes).Name("Get License Types"),
		serverRoute.GET("/licenses/expiring", h.GetExpiringLicenses).Name("Get Expiring Licenses"),
		serverRoute.GET("/licenses/:license_id/reminders", h.GetLicenseReminders).Name("Get License Reminders"),

		// Batch Operations
		serverRoute.POST("/licenses/expired", h.TriggerExpiryDeactivation).Name("Trigger Expiry Deactivation"),
	}
}

// GetAgentLicenses retrieves all licenses for an agent
// AGT-029: Get Agent Licenses
// FR-AGT-PRF-010: License Management Interface
// BR-AGT-PRF-012: License Renewal Period Rules
// BR-AGT-PRF-014: License Renewal Reminders
func (h *LicenseManagementHandler) GetAgentLicenses(sctx *serverRoute.Context, uri AgentIDUri, query GetAgentLicensesQuery) (*resp.AgentLicensesResponse, error) {
	log.Info(sctx.Ctx, "Getting licenses for agent: %s, status filter: %s", uri.AgentID, query.Status)

	var statusFilter *string
	if query.Status != "" {
		statusFilter = &query.Status
	}

	licenses, err := h.licenseRepo.FindByAgentID(sctx.Ctx, uri.AgentID, statusFilter)
	if err != nil {
		log.Error(sctx.Ctx, "Error fetching licenses: %v", err)
		return nil, err
	}

	// Convert to response format
	licenseList := make([]resp.LicenseWithExpiryInfo, 0, len(licenses))
	for i := range licenses {
		licenseList = append(licenseList, resp.ToLicenseWithExpiryInfo(&licenses[i]))
	}

	return &resp.AgentLicensesResponse{
		StatusCodeAndMessage: port.GetSuccess,
		AgentID:              uri.AgentID,
		Licenses:             licenseList,
	}, nil
}

// AddLicense adds a new license with automatic renewal date calculation
// AGT-030: Add License
// FR-AGT-PRF-010: License Management Interface
// FR-AGT-PRF-011: License Renewal Automation
// BR-AGT-PRF-012: License Renewal Period Rules
// BR-AGT-PRF-014: License Renewal Reminders
// BR-AGT-PRF-030: License Date Tracking
// VR-AGT-PRF-031: License Type Mandatory
// VR-AGT-PRF-032: Resident Status Mandatory
// VR-AGT-PRF-036: Authority Date Validation
// WF-AGT-PRF-003: License Renewal Workflow
func (h *LicenseManagementHandler) AddLicense(sctx *serverRoute.Context, req AddLicenseRequest) (*resp.AddLicenseResponse, error) {
	log.Info(sctx.Ctx, "Adding license for agent: %s, type: %s", req.AgentID, req.LicenseType)

	// Parse dates
	licenseDate, err := time.Parse("2006-01-02", req.LicenseDate)
	if err != nil {
		return nil, fmt.Errorf("invalid license_date format: %w", err)
	}

	authorityDate, err := time.Parse("2006-01-02", req.AuthorityDate)
	if err != nil {
		return nil, fmt.Errorf("invalid authority_date format: %w", err)
	}

	// Create license entity
	license := domain.AgentLicense{
		AgentID:                      req.AgentID,
		LicenseLine:                  req.LicenseLine,
		LicenseType:                  req.LicenseType,
		LicenseNumber:                req.LicenseNumber,
		ResidentStatus:               req.ResidentStatus,
		LicenseDate:                  licenseDate,
		AuthorityDate:                authorityDate,
		LicentiatExamPassed:          req.LicentiateExamPassed,
		IsPrimary:                    req.IsPrimary,
		CreatedBy:                    "SYSTEM", // TODO: Get from auth context
	}

	// Create license (repository calculates renewal date)
	createdLicense, err := h.licenseRepo.Create(sctx.Ctx, license)
	if err != nil {
		log.Error(sctx.Ctx, "Error creating license: %v", err)
		return nil, err
	}

	// Create reminders for renewal (BR-AGT-PRF-014)
	err = h.reminderRepo.CreateBatch(sctx.Ctx, createdLicense.LicenseID, createdLicense.RenewalDate, "SYSTEM")
	if err != nil {
		log.Error(sctx.Ctx, "Error creating reminders: %v", err)
		// Continue - license created successfully
	}

	// Fetch reminders for response
	reminders, _ := h.reminderRepo.FindByLicenseID(sctx.Ctx, createdLicense.LicenseID)
	reminderList := make([]resp.ReminderInfo, 0, len(reminders))
	for _, r := range reminders {
		reminderList = append(reminderList, resp.ReminderInfo{
			ReminderType:  r.ReminderType,
			ScheduledDate: r.ReminderDate.Format("2006-01-02"),
			Status:        r.SentStatus,
		})
	}

	// Calculate renewal reason
	renewalPeriod := "1 year"
	renewalReason := "Provisional license"
	if createdLicense.LicenseType == domain.LicenseTypePermanent {
		if createdLicense.LicentiatExamPassed {
			renewalPeriod = "5 years"
			renewalReason = "Permanent license after exam"
		} else {
			renewalReason = "Annual renewal for permanent license"
		}
	}

	log.Info(sctx.Ctx, "License created successfully: %s, renewal date: %s", createdLicense.LicenseID, createdLicense.RenewalDate.Format("2006-01-02"))

	return &resp.AddLicenseResponse{
		StatusCodeAndMessage: port.CreateSuccess,
		License:              resp.ToLicenseWithExpiryInfo(createdLicense),
		RenewalCalculation: resp.RenewalCalculationInfo{
			LicenseType:   createdLicense.LicenseType,
			RenewalPeriod: renewalPeriod,
			Reason:        renewalReason,
		},
		RemindersScheduled: reminderList,
	}, nil
}

// GetLicenseDetails retrieves complete license details
// AGT-031: Get License Details
// FR-AGT-PRF-010: License Management Interface
// BR-AGT-PRF-014: License Renewal Reminders
// FR-AGT-PRF-022: Audit History Tracking
func (h *LicenseManagementHandler) GetLicenseDetails(sctx *serverRoute.Context, uri struct {
	AgentID   string `uri:"agent_id" validate:"required,uuid4"`
	LicenseID string `uri:"license_id" validate:"required,uuid4"`
}) (*resp.LicenseDetailsResponse, error) {
	log.Info(sctx.Ctx, "Getting license details: %s for agent: %s", uri.LicenseID, uri.AgentID)

	license, err := h.licenseRepo.FindByID(sctx.Ctx, uri.LicenseID)
	if err != nil {
		log.Error(sctx.Ctx, "Error fetching license: %v", err)
		return nil, err
	}

	// Verify license belongs to agent
	if license.AgentID != uri.AgentID {
		return nil, fmt.Errorf("license does not belong to agent")
	}

	// Fetch reminders
	reminders, err := h.reminderRepo.FindByLicenseID(sctx.Ctx, uri.LicenseID)
	if err != nil {
		log.Error(sctx.Ctx, "Error fetching reminders: %v", err)
		// Continue without reminders
	}

	reminderList := make([]resp.ReminderInfo, 0, len(reminders))
	for _, r := range reminders {
		sentDate := ""
		if r.SentDate.Valid {
			sentDate = r.SentDate.Time.Format("2006-01-02")
		}
		reminderList = append(reminderList, resp.ReminderInfo{
			ReminderID:    r.ReminderID,
			ReminderType:  r.ReminderType,
			ScheduledDate: r.ReminderDate.Format("2006-01-02"),
			Status:        r.SentStatus,
			SentDate:      sentDate,
			EmailSent:     r.EmailSent,
			SMSSent:       r.SMSSent,
		})
	}

	// TODO: Fetch renewal history from audit logs

	return &resp.LicenseDetailsResponse{
		StatusCodeAndMessage: port.GetSuccess,
		License:              resp.ToLicenseWithExpiryInfo(license),
		RenewalHistory:       []resp.RenewalHistoryEntry{}, // TODO: Implement
		ReminderSchedule:     reminderList,
	}, nil
}

// UpdateLicense updates license details
// AGT-032: Update License
// FR-AGT-PRF-010: License Management Interface
// VR-AGT-PRF-032: Resident Status Mandatory
// VR-AGT-PRF-036: Authority Date Validation
// FR-AGT-PRF-022: Audit History Tracking
func (h *LicenseManagementHandler) UpdateLicense(sctx *serverRoute.Context, req UpdateLicenseRequest) (*resp.UpdateLicenseResponse, error) {
	log.Info(sctx.Ctx, "Updating license: %s for agent: %s", req.LicenseID, req.AgentID)

	// Build updates map
	updates := make(map[string]interface{})
	updatedFields := []string{}

	if req.LicenseNumber != nil {
		updates["license_number"] = *req.LicenseNumber
		updatedFields = append(updatedFields, "license_number")
	}
	if req.ResidentStatus != nil {
		updates["resident_status"] = *req.ResidentStatus
		updatedFields = append(updatedFields, "resident_status")
	}
	if req.AuthorityDate != nil {
		authorityDate, err := time.Parse("2006-01-02", *req.AuthorityDate)
		if err != nil {
			return nil, fmt.Errorf("invalid authority_date format: %w", err)
		}
		updates["authority_date"] = authorityDate
		updatedFields = append(updatedFields, "authority_date")
	}
	if req.IsPrimary != nil {
		updates["is_primary"] = *req.IsPrimary
		updatedFields = append(updatedFields, "is_primary")
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	updatedLicense, err := h.licenseRepo.Update(sctx.Ctx, req.LicenseID, updates, req.UpdatedBy)
	if err != nil {
		log.Error(sctx.Ctx, "Error updating license: %v", err)
		return nil, err
	}

	// Verify license belongs to agent
	if updatedLicense.AgentID != req.AgentID {
		return nil, fmt.Errorf("license does not belong to agent")
	}

	log.Info(sctx.Ctx, "License updated successfully: %s", req.LicenseID)

	return &resp.UpdateLicenseResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		License:              resp.ToLicenseWithExpiryInfo(updatedLicense),
		UpdatedFields:        updatedFields,
	}, nil
}

// RenewLicense renews a license with period calculation
// AGT-033: Renew License
// FR-AGT-PRF-010: License Management Interface
// FR-AGT-PRF-011: License Renewal Automation
// BR-AGT-PRF-012: Complex Renewal Rules (Provisional to Permanent Conversion)
// WF-AGT-PRF-003: License Renewal Workflow
// FR-AGT-PRF-022: Audit History Tracking
func (h *LicenseManagementHandler) RenewLicense(sctx *serverRoute.Context, req RenewLicenseRequest) (*resp.RenewLicenseResponse, error) {
	log.Info(sctx.Ctx, "Renewing license: %s for agent: %s, type: %s", req.LicenseID, req.AgentID, req.RenewalType)

	// Get current license to capture previous renewal date
	currentLicense, err := h.licenseRepo.FindByID(sctx.Ctx, req.LicenseID)
	if err != nil {
		log.Error(sctx.Ctx, "Error fetching license: %v", err)
		return nil, err
	}

	// Verify license belongs to agent
	if currentLicense.AgentID != req.AgentID {
		return nil, fmt.Errorf("license does not belong to agent")
	}

	previousRenewalDate := currentLicense.RenewalDate

	// Parse exam date if provided
	var examDate *time.Time
	if req.ExamDate != nil {
		parsedDate, err := time.Parse("2006-01-02", *req.ExamDate)
		if err != nil {
			return nil, fmt.Errorf("invalid exam_date format: %w", err)
		}
		examDate = &parsedDate
	}

	// Renew license
	renewedLicense, err := h.licenseRepo.Renew(
		sctx.Ctx,
		req.LicenseID,
		req.RenewalType,
		req.LicentiateExamPassed,
		examDate,
		req.ExamCertificateNumber,
		req.RenewedBy,
	)
	if err != nil {
		log.Error(sctx.Ctx, "Error renewing license: %v", err)
		return nil, err
	}

	// Determine if converted to permanent
	convertedToPermanent := currentLicense.IsProvisional() && renewedLicense.IsPermanent()

	// Calculate renewal reason
	renewalPeriod := "1 year"
	renewalReason := "Provisional renewal"
	if renewedLicense.IsPermanent() {
		if convertedToPermanent {
			renewalPeriod = "5 years"
			renewalReason = "Converted to permanent after exam"
		} else {
			renewalReason = "Annual renewal for permanent license"
		}
	}

	log.Info(sctx.Ctx, "License renewed successfully: %s, new renewal date: %s, converted: %v",
		req.LicenseID, renewedLicense.RenewalDate.Format("2006-01-02"), convertedToPermanent)

	return &resp.RenewLicenseResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		License:              resp.ToLicenseWithExpiryInfo(renewedLicense),
		RenewalCalculation: resp.RenewalCalculationInfo{
			LicenseType:   renewedLicense.LicenseType,
			RenewalPeriod: renewalPeriod,
			Reason:        renewalReason,
		},
		PreviousRenewalDate:  previousRenewalDate.Format("2006-01-02"),
		NewRenewalDate:       renewedLicense.RenewalDate.Format("2006-01-02"),
		ConvertedToPermanent: convertedToPermanent,
	}, nil
}

// DeleteLicense soft deletes a license
// AGT-034: Delete License
// FR-AGT-PRF-010: License Management Interface
// FR-AGT-PRF-022: Audit History Tracking
func (h *LicenseManagementHandler) DeleteLicense(sctx *serverRoute.Context, uri struct {
	AgentID   string `uri:"agent_id" validate:"required,uuid4"`
	LicenseID string `uri:"license_id" validate:"required,uuid4"`
}) error {
	log.Info(sctx.Ctx, "Deleting license: %s for agent: %s", uri.LicenseID, uri.AgentID)

	// Verify license exists and belongs to agent
	license, err := h.licenseRepo.FindByID(sctx.Ctx, uri.LicenseID)
	if err != nil {
		log.Error(sctx.Ctx, "Error fetching license: %v", err)
		return err
	}

	if license.AgentID != uri.AgentID {
		return fmt.Errorf("license does not belong to agent")
	}

	err = h.licenseRepo.Delete(sctx.Ctx, uri.LicenseID, "SYSTEM") // TODO: Get from auth
	if err != nil {
		log.Error(sctx.Ctx, "Error deleting license: %v", err)
		return err
	}

	log.Info(sctx.Ctx, "License deleted successfully: %s", uri.LicenseID)
	return nil
}

// GetLicenseTypes returns available license types (lookup)
// AGT-035: Get License Types
// FR-AGT-PRF-010: License Management Interface
// BR-AGT-PRF-012: License Renewal Period Rules
func (h *LicenseManagementHandler) GetLicenseTypes(sctx *serverRoute.Context) (*resp.LicenseTypesResponse, error) {
	log.Info(sctx.Ctx, "Getting license types")

	licenseTypes := []resp.LicenseTypeInfo{
		{
			Code:        domain.LicenseTypeProvisional,
			Name:        "Provisional License",
			Description: "1-year validity, renewable up to 2 times before exam required",
		},
		{
			Code:        domain.LicenseTypePermanent,
			Name:        "Permanent License",
			Description: "5-year validity after exam, annual renewal required",
		},
	}

	return &resp.LicenseTypesResponse{
		StatusCodeAndMessage: port.GetSuccess,
		LicenseTypes:         licenseTypes,
	}, nil
}

// GetExpiringLicenses retrieves licenses expiring within specified days
// AGT-036: Get Expiring Licenses
// FR-AGT-PRF-011: License Renewal Automation
// BR-AGT-PRF-012: License Renewal Period Rules
// BR-AGT-PRF-014: License Renewal Reminders
func (h *LicenseManagementHandler) GetExpiringLicenses(sctx *serverRoute.Context, query GetExpiringLicensesQuery) (*resp.ExpiringLicensesResponse, error) {
	log.Info(sctx.Ctx, "Getting expiring licenses within %d days", query.Days)

	// Default values
	if query.Days == 0 {
		query.Days = 30
	}
	if query.Page == 0 {
		query.Page = 1
	}
	if query.Limit == 0 {
		query.Limit = 20
	}

	var officeFilter *string
	if query.OfficeCode != "" {
		officeFilter = &query.OfficeCode
	}

	licenses, totalCount, err := h.licenseRepo.FindExpiring(sctx.Ctx, query.Days, officeFilter, query.Page, query.Limit)
	if err != nil {
		log.Error(sctx.Ctx, "Error fetching expiring licenses: %v", err)
		return nil, err
	}

	// Convert to response format with agent details
	expiringList := make([]resp.ExpiringLicenseInfo, 0, len(licenses))
	for _, license := range licenses {
		// TODO: Fetch agent details (name, mobile, email)
		expiringList = append(expiringList, resp.ExpiringLicenseInfo{
			LicenseID:       license.LicenseID,
			AgentID:         license.AgentID,
			AgentName:       "Agent Name", // TODO: Join with agent_profiles
			LicenseNumber:   license.LicenseNumber,
			LicenseType:     license.LicenseType,
			RenewalDate:     license.RenewalDate.Format("2006-01-02"),
			DaysUntilExpiry: license.DaysUntilExpiry(),
			AgentMobile:     "",           // TODO: Join with agent_contacts
			AgentEmail:      "",           // TODO: Join with agent_emails
			OfficeCode:      "",           // TODO: Join with agent_profiles
		})
	}

	totalPages := int(totalCount) / query.Limit
	if int(totalCount)%query.Limit > 0 {
		totalPages++
	}

	return &resp.ExpiringLicensesResponse{
		StatusCodeAndMessage: port.GetSuccess,
		MetaDataResponse: port.MetaDataResponse{
			Page:        query.Page,
			Limit:       query.Limit,
			TotalCount:  int(totalCount),
			TotalPages:  totalPages,
			HasNext:     query.Page < totalPages,
			HasPrevious: query.Page > 1,
		},
		Licenses: expiringList,
	}, nil
}

// GetLicenseReminders retrieves reminder schedule for a license
// AGT-037: Get License Reminders
// FR-AGT-PRF-011: License Renewal Automation
// BR-AGT-PRF-014: License Renewal Reminders
// WF-AGT-PRF-003: License Renewal Workflow
func (h *LicenseManagementHandler) GetLicenseReminders(sctx *serverRoute.Context, uri LicenseIDUri) (*resp.LicenseRemindersResponse, error) {
	log.Info(sctx.Ctx, "Getting reminders for license: %s", uri.LicenseID)

	reminders, err := h.reminderRepo.FindByLicenseID(sctx.Ctx, uri.LicenseID)
	if err != nil {
		log.Error(sctx.Ctx, "Error fetching reminders: %v", err)
		return nil, err
	}

	reminderList := make([]resp.ReminderInfo, 0, len(reminders))
	for _, r := range reminders {
		sentDate := ""
		if r.SentDate.Valid {
			sentDate = r.SentDate.Time.Format("2006-01-02")
		}
		reminderList = append(reminderList, resp.ReminderInfo{
			ReminderID:    r.ReminderID,
			ReminderType:  r.ReminderType,
			ScheduledDate: r.ReminderDate.Format("2006-01-02"),
			Status:        r.SentStatus,
			SentDate:      sentDate,
			EmailSent:     r.EmailSent,
			SMSSent:       r.SMSSent,
		})
	}

	return &resp.LicenseRemindersResponse{
		StatusCodeAndMessage: port.GetSuccess,
		LicenseID:            uri.LicenseID,
		Reminders:            reminderList,
	}, nil
}

// TriggerExpiryDeactivation batch job to deactivate agents with expired licenses
// AGT-038: Trigger License Expiry Deactivation
// FR-AGT-PRF-012: License Auto-Deactivation
// BR-AGT-PRF-013: Auto-Deactivation on License Expiry
// WF-AGT-PRF-007: License Deactivation Workflow
func (h *LicenseManagementHandler) TriggerExpiryDeactivation(sctx *serverRoute.Context, req TriggerExpiryDeactivationRequest) (*resp.ExpiryDeactivationResponse, error) {
	log.Info(sctx.Ctx, "Triggering expiry deactivation for batch date: %s, dry_run: %v", req.BatchDate, req.DryRun)

	batchDate, err := time.Parse("2006-01-02", req.BatchDate)
	if err != nil {
		return nil, fmt.Errorf("invalid batch_date format: %w", err)
	}

	agentIDs, err := h.licenseRepo.DeactivateExpiredAgents(sctx.Ctx, batchDate, req.DryRun)
	if err != nil {
		log.Error(sctx.Ctx, "Error deactivating expired agents: %v", err)
		return nil, err
	}

	batchID := uuid.New().String()
	notificationsSent := 0

	// TODO: Implement actual agent status deactivation
	// TODO: Disable portal access
	// TODO: Send notifications (email/SMS)

	if !req.DryRun {
		log.Info(sctx.Ctx, "Batch %s completed: %d agents deactivated", batchID, len(agentIDs))
	} else {
		log.Info(sctx.Ctx, "Dry run completed: %d agents would be deactivated", len(agentIDs))
	}

	return &resp.ExpiryDeactivationResponse{
		StatusCodeAndMessage: port.CreateSuccess,
		BatchID:              batchID,
		BatchDate:            req.BatchDate,
		AgentsDeactivated:    len(agentIDs),
		AgentIDs:             agentIDs,
		NotificationsSent:    notificationsSent,
		DryRun:               req.DryRun,
	}, nil
}
