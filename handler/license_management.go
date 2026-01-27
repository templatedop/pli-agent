package handler

import (
	"fmt"
	"time"

	"pli-agent-api/core/domain"
	"pli-agent-api/core/port"
	req "pli-agent-api/handler/request"
	resp "pli-agent-api/handler/response"
	repo "pli-agent-api/repo/postgres"

	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"
)

// AgentLicenseHandler handles license management APIs
// AGT-029 to AGT-038: License Management
// FR-AGT-PRF-010: License Management
// FR-AGT-PRF-011: License Renewal
// BR-AGT-PRF-012: License Renewal Period Rules
// BR-AGT-PRF-013: Auto-Deactivation on License Expiry
// BR-AGT-PRF-014: License Renewal Reminder Schedule
type AgentLicenseHandler struct {
	*serverHandler.Base
	licenseRepo  *repo.AgentLicenseRepository
	profileRepo  *repo.AgentProfileRepository
	auditLogRepo *repo.AgentAuditLogRepository
}

// NewAgentLicenseHandler creates a new license management handler
func NewAgentLicenseHandler(
	licenseRepo *repo.AgentLicenseRepository,
	profileRepo *repo.AgentProfileRepository,
	auditLogRepo *repo.AgentAuditLogRepository,
) *AgentLicenseHandler {
	return &AgentLicenseHandler{
		Base:         &serverHandler.Base{},
		licenseRepo:  licenseRepo,
		profileRepo:  profileRepo,
		auditLogRepo: auditLogRepo,
	}
}

// RegisterRoutes registers all license management routes
func (h *AgentLicenseHandler) RegisterRoutes() []serverRoute.Route {
	return []serverRoute.Route{
		// AGT-029: Get Agent Licenses
		serverRoute.NewRoute("GET", "/agents/:agent_id/licenses", h.GetAgentLicenses),
		// AGT-030: Add License
		serverRoute.NewRoute("POST", "/agents/:agent_id/licenses", h.AddLicense),
		// AGT-031: Get License Details
		serverRoute.NewRoute("GET", "/agents/:agent_id/licenses/:license_id", h.GetLicenseDetails),
		// AGT-032: Update License
		serverRoute.NewRoute("PUT", "/agents/:agent_id/licenses/:license_id", h.UpdateLicense),
		// AGT-033: Renew License
		serverRoute.NewRoute("PUT", "/agents/:agent_id/licenses/:license_id/renew", h.RenewLicense),
		// AGT-034: Delete License
		serverRoute.NewRoute("DELETE", "/agents/:agent_id/licenses/:license_id", h.DeleteLicense),
		// AGT-035: Get License Types
		serverRoute.NewRoute("GET", "/license-types", h.GetLicenseTypes),
		// AGT-036: Get Expiring Licenses
		serverRoute.NewRoute("GET", "/licenses/expiring", h.GetExpiringLicenses),
		// AGT-037: Get License Reminders
		serverRoute.NewRoute("GET", "/licenses/:license_id/reminders", h.GetLicenseReminders),
		// AGT-038: Batch Deactivate Expired Licenses
		serverRoute.NewRoute("POST", "/licenses/expired", h.BatchDeactivateExpired),
	}
}

// AgentIDUri represents agent_id path parameter
type AgentIDUri struct {
	AgentID string `uri:"agent_id" validate:"required"`
}

// LicenseIDUri represents license_id path parameter
type LicenseIDUri struct {
	LicenseID string `uri:"license_id" validate:"required,uuid"`
}

// AgentLicenseIDUri represents both agent_id and license_id path parameters
type AgentLicenseIDUri struct {
	AgentID   string `uri:"agent_id" validate:"required"`
	LicenseID string `uri:"license_id" validate:"required,uuid"`
}

// GetAgentLicenses retrieves all licenses for an agent
// AGT-029: Get Agent Licenses
// FR-AGT-PRF-010: License Management
func (h *AgentLicenseHandler) GetAgentLicenses(sctx *serverRoute.Context, uri AgentIDUri) (*resp.AgentLicensesResponse, error) {
	log.Info(sctx.Ctx, "Fetching licenses for agent: %s", uri.AgentID)

	// Verify agent exists
	_, err := h.profileRepo.FindByID(sctx.Ctx, uri.AgentID)
	if err != nil {
		log.Error(sctx.Ctx, "Agent not found: %v", err)
		return nil, err
	}

	// Fetch all licenses for agent
	licenses, err := h.licenseRepo.FindByAgentID(sctx.Ctx, uri.AgentID)
	if err != nil {
		log.Error(sctx.Ctx, "Error fetching licenses: %v", err)
		return nil, err
	}

	// Convert to DTOs with computed fields
	licenseDTOs := make([]resp.LicenseDTO, len(licenses))
	hasPrimary := false
	for i, license := range licenses {
		licenseDTOs[i] = h.toLicenseDTO(license)
		if license.IsPrimary {
			hasPrimary = true
		}
	}

	log.Info(sctx.Ctx, "Found %d licenses for agent: %s", len(licenses), uri.AgentID)

	return &resp.AgentLicensesResponse{
		StatusCodeAndMessage: port.LicenseListSuccess,
		AgentID:              uri.AgentID,
		Licenses:             licenseDTOs,
		TotalCount:           len(licenses),
		HasPrimaryLicense:    hasPrimary,
	}, nil
}

// AddLicense adds a new license for an agent
// AGT-030: Add License
// FR-AGT-PRF-010: License Management
// BR-AGT-PRF-012: License Renewal Period Rules
// VR-AGT-PRF-031 to VR-AGT-PRF-036: License validations
func (h *AgentLicenseHandler) AddLicense(sctx *serverRoute.Context, uri AgentIDUri, request req.AddLicenseRequest) (*resp.AddLicenseResponse, error) {
	log.Info(sctx.Ctx, "Adding license for agent: %s", uri.AgentID)

	// Verify agent exists
	_, err := h.profileRepo.FindByID(sctx.Ctx, uri.AgentID)
	if err != nil {
		log.Error(sctx.Ctx, "Agent not found: %v", err)
		return nil, err
	}

	// Validate license number uniqueness
	isUnique, err := h.licenseRepo.ValidateLicenseNumberUniqueness(sctx.Ctx, request.LicenseNumber, "")
	if err != nil {
		log.Error(sctx.Ctx, "Error validating license number uniqueness: %v", err)
		return nil, err
	}
	if !isUnique {
		return nil, fmt.Errorf("license number %s already exists", request.LicenseNumber)
	}

	// Calculate renewal date based on license type
	// BR-AGT-PRF-012: Provisional = 1 year, Permanent (after exam) = 5 years
	renewalDate := h.calculateRenewalDate(request.LicenseType, request.LicenseDate, request.LicentiateExamPassed)

	// Create license entity
	license := domain.AgentLicense{
		AgentID:                     uri.AgentID,
		LicenseLine:                 request.LicenseLine,
		LicenseType:                 request.LicenseType,
		LicenseNumber:               request.LicenseNumber,
		ResidentStatus:              request.ResidentStatus,
		LicenseDate:                 request.LicenseDate,
		RenewalDate:                 renewalDate,
		AuthorityDate:               request.AuthorityDate,
		RenewalCount:                0,
		LicenseStatus:               domain.LicenseStatusActive,
		IsPrimary:                   request.IsPrimary,
		LicentiateExamPassed:        request.LicentiateExamPassed,
		LicentiateExamDate:          request.LicentiateExamDate,
		LicentiateCertificateNumber: request.LicentiateCertificateNumber,
		Metadata:                    request.Metadata,
		CreatedBy:                   request.CreatedBy,
	}

	// Create license (repository handles audit logging via CTE)
	createdLicense, err := h.licenseRepo.Create(sctx.Ctx, license)
	if err != nil {
		log.Error(sctx.Ctx, "Error creating license: %v", err)
		return nil, err
	}

	log.Info(sctx.Ctx, "License created successfully: %s for agent: %s", createdLicense.LicenseID, uri.AgentID)

	return &resp.AddLicenseResponse{
		StatusCodeAndMessage: port.LicenseCreateSuccess,
		License:              h.toLicenseDTO(*createdLicense),
	}, nil
}

// GetLicenseDetails retrieves detailed license information
// AGT-031: Get License Details
func (h *AgentLicenseHandler) GetLicenseDetails(sctx *serverRoute.Context, uri AgentLicenseIDUri) (*resp.LicenseDetailsResponse, error) {
	log.Info(sctx.Ctx, "Fetching license details: %s for agent: %s", uri.LicenseID, uri.AgentID)

	// Fetch license
	license, err := h.licenseRepo.FindByID(sctx.Ctx, uri.LicenseID)
	if err != nil {
		log.Error(sctx.Ctx, "License not found: %v", err)
		return nil, err
	}

	// Verify license belongs to agent
	if license.AgentID != uri.AgentID {
		return nil, fmt.Errorf("license %s does not belong to agent %s", uri.LicenseID, uri.AgentID)
	}

	// Fetch renewal history from audit logs
	auditLogs, err := h.auditLogRepo.FindByAgentIDAndAction(sctx.Ctx, uri.AgentID, domain.AuditActionLicenseUpdate)
	if err != nil {
		log.Warn(sctx.Ctx, "Could not fetch audit logs: %v", err)
		auditLogs = []domain.AgentAuditLog{} // Continue without history
	}

	// Build renewal history (filter for this license)
	renewalHistory := h.buildRenewalHistory(auditLogs, uri.LicenseID)

	// Determine renewal eligibility
	canRenew, reason := h.canRenewLicense(*license)

	log.Info(sctx.Ctx, "License details fetched successfully: %s", uri.LicenseID)

	return &resp.LicenseDetailsResponse{
		StatusCodeAndMessage:     port.LicenseDetailSuccess,
		License:                  h.toLicenseDTO(*license),
		RenewalHistory:           renewalHistory,
		CanRenew:                 canRenew,
		RenewalEligibilityReason: reason,
	}, nil
}

// UpdateLicense updates an existing license
// AGT-032: Update License
// FR-AGT-PRF-010: License Management
func (h *AgentLicenseHandler) UpdateLicense(sctx *serverRoute.Context, uri AgentLicenseIDUri, request req.UpdateLicenseRequest) (*resp.UpdateLicenseResponse, error) {
	log.Info(sctx.Ctx, "Updating license: %s for agent: %s", uri.LicenseID, uri.AgentID)

	// Fetch existing license
	license, err := h.licenseRepo.FindByID(sctx.Ctx, uri.LicenseID)
	if err != nil {
		log.Error(sctx.Ctx, "License not found: %v", err)
		return nil, err
	}

	// Verify license belongs to agent
	if license.AgentID != uri.AgentID {
		return nil, fmt.Errorf("license %s does not belong to agent %s", uri.LicenseID, uri.AgentID)
	}

	// Build updates map
	updates := make(map[string]interface{})
	if request.LicenseLine != nil {
		updates["license_line"] = *request.LicenseLine
	}
	if request.LicenseType != nil {
		updates["license_type"] = *request.LicenseType
	}
	if request.LicenseNumber != nil {
		// Validate uniqueness if changing license number
		if *request.LicenseNumber != license.LicenseNumber {
			isUnique, err := h.licenseRepo.ValidateLicenseNumberUniqueness(sctx.Ctx, *request.LicenseNumber, uri.LicenseID)
			if err != nil {
				log.Error(sctx.Ctx, "Error validating license number uniqueness: %v", err)
				return nil, err
			}
			if !isUnique {
				return nil, fmt.Errorf("license number %s already exists", *request.LicenseNumber)
			}
		}
		updates["license_number"] = *request.LicenseNumber
	}
	if request.ResidentStatus != nil {
		updates["resident_status"] = *request.ResidentStatus
	}
	if request.LicenseDate != nil {
		updates["license_date"] = *request.LicenseDate
	}
	if request.AuthorityDate != nil {
		updates["authority_date"] = *request.AuthorityDate
	}
	if request.LicenseStatus != nil {
		updates["license_status"] = *request.LicenseStatus
	}
	if request.LicentiateExamPassed != nil {
		updates["licentiate_exam_passed"] = *request.LicentiateExamPassed
	}
	if request.LicentiateExamDate != nil {
		updates["licentiate_exam_date"] = *request.LicentiateExamDate
	}
	if request.LicentiateCertificateNumber != nil {
		updates["licentiate_certificate_number"] = *request.LicentiateCertificateNumber
	}
	if request.IsPrimary != nil {
		updates["is_primary"] = *request.IsPrimary
	}
	if request.Metadata != nil {
		updates["metadata"] = *request.Metadata
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	// Update license and get updated license in SINGLE database hit
	// Repository uses RETURNING clause to eliminate extra SELECT
	updatedLicense, err := h.licenseRepo.Update(sctx.Ctx, uri.LicenseID, updates, request.UpdatedBy)
	if err != nil {
		log.Error(sctx.Ctx, "Error updating license: %v", err)
		return nil, err
	}

	log.Info(sctx.Ctx, "License updated successfully: %s", uri.LicenseID)

	return &resp.UpdateLicenseResponse{
		StatusCodeAndMessage: port.LicenseUpdateSuccess,
		License:              h.toLicenseDTO(*updatedLicense),
	}, nil
}

// RenewLicense renews a license with complex renewal rules
// AGT-033: Renew License
// FR-AGT-PRF-011: License Renewal
// BR-AGT-PRF-012: License Renewal Period Rules (COMPLEX)
func (h *AgentLicenseHandler) RenewLicense(sctx *serverRoute.Context, uri AgentLicenseIDUri, request req.RenewLicenseRequest) (*resp.RenewLicenseResponse, error) {
	log.Info(sctx.Ctx, "Renewing license: %s for agent: %s (type: %s)", uri.LicenseID, uri.AgentID, request.RenewalType)

	// Fetch existing license
	license, err := h.licenseRepo.FindByID(sctx.Ctx, uri.LicenseID)
	if err != nil {
		log.Error(sctx.Ctx, "License not found: %v", err)
		return nil, err
	}

	// Verify license belongs to agent
	if license.AgentID != uri.AgentID {
		return nil, fmt.Errorf("license %s does not belong to agent %s", uri.LicenseID, uri.AgentID)
	}

	// Check renewal eligibility
	canRenew, reason := h.canRenewLicense(*license)
	if !canRenew {
		return nil, fmt.Errorf("license cannot be renewed: %s", reason)
	}

	previousExpiry := license.RenewalDate
	var renewedLicense *domain.AgentLicense
	var newRenewalDate time.Time
	var renewalMessage string

	// BR-AGT-PRF-012: Apply complex renewal rules
	switch request.RenewalType {
	case "PROVISIONAL_RENEWAL":
		// Provisional license: 1 year validity, max 2 renewals
		if license.LicenseType != domain.LicenseTypeProvisional {
			return nil, fmt.Errorf("can only do provisional renewal on provisional licenses")
		}
		if license.RenewalCount >= 2 {
			return nil, fmt.Errorf("provisional license can only be renewed 2 times")
		}
		newRenewalDate = time.Now().AddDate(1, 0, 0)
		renewalMessage = fmt.Sprintf("Provisional license renewed for 1 year (renewal %d/2)", license.RenewalCount+1)

		// Repository returns renewed license using RETURNING (single database hit)
		renewedLicense, err = h.licenseRepo.RenewLicense(sctx.Ctx, uri.LicenseID, request.UpdatedBy, newRenewalDate)

	case "CONVERT_TO_PERMANENT":
		// Convert provisional to permanent after passing exam
		if license.LicenseType != domain.LicenseTypeProvisional {
			return nil, fmt.Errorf("can only convert provisional licenses to permanent")
		}
		if !request.ExamPassed || request.ExamDate == nil || request.ExamCertificateNumber == nil {
			return nil, fmt.Errorf("exam details required for conversion to permanent")
		}
		// Check if within 3 years
		yearsSinceLicense := time.Since(license.LicenseDate).Hours() / 24 / 365
		if yearsSinceLicense > 3 {
			return nil, fmt.Errorf("exam must be passed within 3 years of provisional license")
		}
		// Permanent license: 5 years validity after exam
		newRenewalDate = request.ExamDate.AddDate(5, 0, 0)
		renewalMessage = "License converted to permanent after passing exam. Valid for 5 years, renewable annually."

		// Repository returns converted license using RETURNING (single database hit)
		renewedLicense, err = h.licenseRepo.ConvertToPermanent(sctx.Ctx, uri.LicenseID, request.UpdatedBy, *request.ExamDate, *request.ExamCertificateNumber)

	case "PERMANENT_RENEWAL":
		// Permanent license: Annual renewal required
		if license.LicenseType != domain.LicenseTypePermanent {
			return nil, fmt.Errorf("can only do permanent renewal on permanent licenses")
		}
		newRenewalDate = time.Now().AddDate(1, 0, 0)
		renewalMessage = "Permanent license renewed for 1 year"

		// Repository returns renewed license using RETURNING (single database hit)
		renewedLicense, err = h.licenseRepo.RenewLicense(sctx.Ctx, uri.LicenseID, request.UpdatedBy, newRenewalDate)

	default:
		return nil, fmt.Errorf("invalid renewal type: %s", request.RenewalType)
	}

	if err != nil {
		log.Error(sctx.Ctx, "Error renewing license: %v", err)
		return nil, err
	}

	log.Info(sctx.Ctx, "License renewed successfully: %s (type: %s)", uri.LicenseID, request.RenewalType)

	return &resp.RenewLicenseResponse{
		StatusCodeAndMessage: port.LicenseRenewSuccess,
		License:              h.toLicenseDTO(*renewedLicense),
		RenewalType:          request.RenewalType,
		PreviousExpiry:       previousExpiry,
		NewExpiry:            newRenewalDate,
		RenewalMessage:       renewalMessage,
	}, nil
}

// DeleteLicense soft-deletes a license
// AGT-034: Delete License
func (h *AgentLicenseHandler) DeleteLicense(sctx *serverRoute.Context, uri AgentLicenseIDUri) (*serverHandler.EmptyResponse, error) {
	log.Info(sctx.Ctx, "Deleting license: %s for agent: %s", uri.LicenseID, uri.AgentID)

	// Fetch license to verify it exists and belongs to agent
	license, err := h.licenseRepo.FindByID(sctx.Ctx, uri.LicenseID)
	if err != nil {
		log.Error(sctx.Ctx, "License not found: %v", err)
		return nil, err
	}

	if license.AgentID != uri.AgentID {
		return nil, fmt.Errorf("license %s does not belong to agent %s", uri.LicenseID, uri.AgentID)
	}

	// Soft delete license (repository handles audit logging via CTE)
	err = h.licenseRepo.Delete(sctx.Ctx, uri.LicenseID, "SYSTEM") // TODO: Get user from context
	if err != nil {
		log.Error(sctx.Ctx, "Error deleting license: %v", err)
		return nil, err
	}

	log.Info(sctx.Ctx, "License deleted successfully: %s", uri.LicenseID)

	return &serverHandler.EmptyResponse{}, nil
}

// GetLicenseTypes returns available license types
// AGT-035: Get License Types
func (h *AgentLicenseHandler) GetLicenseTypes(sctx *serverRoute.Context) (*resp.LicenseTypesResponse, error) {
	log.Info(sctx.Ctx, "Fetching license types")

	// Return static license type metadata
	// BR-AGT-PRF-012: License type rules
	licenseTypes := []resp.LicenseTypeDTO{
		{
			Code:                domain.LicenseTypeProvisional,
			Name:                "Provisional License",
			Description:         "Provisional license valid for 1 year, renewable up to 2 times. Must pass exam within 3 years.",
			ValidityYears:       1,
			RenewalIntervalDays: 365,
			MaxRenewals:         2,
		},
		{
			Code:                domain.LicenseTypePermanent,
			Name:                "Permanent License",
			Description:         "Permanent license valid for 5 years after passing exam. Annual renewal required.",
			ValidityYears:       5,
			RenewalIntervalDays: 365,
			MaxRenewals:         -1, // Unlimited
		},
	}

	return &resp.LicenseTypesResponse{
		StatusCodeAndMessage: port.LicenseTypeListSuccess,
		LicenseTypes:         licenseTypes,
	}, nil
}

// GetExpiringLicenses retrieves licenses expiring within specified days
// AGT-036: Get Expiring Licenses
// BR-AGT-PRF-014: License Renewal Reminder Schedule
// OPTIMIZED: Single database hit with JOIN (no N+1 query problem)
func (h *AgentLicenseHandler) GetExpiringLicenses(sctx *serverRoute.Context, query req.GetExpiringLicensesQuery) (*resp.ExpiringLicensesResponse, error) {
	log.Info(sctx.Ctx, "Fetching licenses expiring within %d days", query.Days)

	// Set defaults
	if query.Days == 0 {
		query.Days = 30
	}
	if query.Page == 0 {
		query.Page = 1
	}
	if query.Limit == 0 {
		query.Limit = 50
	}

	// Fetch expiring licenses WITH agent details in SINGLE database hit
	// Uses JOIN to eliminate N+1 query problem
	licensesWithProfiles, err := h.licenseRepo.FindExpiringLicensesWithAgentDetails(sctx.Ctx, query.Days)
	if err != nil {
		log.Error(sctx.Ctx, "Error fetching expiring licenses: %v", err)
		return nil, err
	}

	// Convert to DTOs (no additional database calls needed!)
	expiringDTOs := make([]resp.ExpiringLicenseDTO, 0, len(licensesWithProfiles))
	for _, lp := range licensesWithProfiles {
		daysRemaining := int(time.Until(lp.RenewalDate).Hours() / 24)

		expiringDTOs = append(expiringDTOs, resp.ExpiringLicenseDTO{
			LicenseID:     lp.LicenseID,
			AgentID:       lp.AgentID,
			AgentCode:     lp.AgentCode,
			AgentName:     fmt.Sprintf("%s %s %s", lp.FirstName, lp.MiddleName, lp.LastName),
			LicenseLine:   lp.LicenseLine,
			LicenseType:   lp.LicenseType,
			LicenseNumber: lp.LicenseNumber,
			RenewalDate:   lp.RenewalDate,
			DaysRemaining: daysRemaining,
			RenewalCount:  lp.RenewalCount,
			OfficeCode:    lp.OfficeCode,
			OfficeName:    lp.OfficeCode, // TODO: Fetch actual office name from office table
			ContactMobile: "",            // TODO: Add to JOIN if contact info in same table
			ContactEmail:  "",            // TODO: Add to JOIN if email info in same table
		})
	}

	// Calculate summary
	summary := struct {
		TotalExpiring    int `json:"total_expiring"`
		ExpiringIn7Days  int `json:"expiring_in_7_days"`
		ExpiringIn15Days int `json:"expiring_in_15_days"`
		ExpiringIn30Days int `json:"expiring_in_30_days"`
	}{
		TotalExpiring: len(expiringDTOs),
	}

	for _, dto := range expiringDTOs {
		if dto.DaysRemaining <= 7 {
			summary.ExpiringIn7Days++
		}
		if dto.DaysRemaining <= 15 {
			summary.ExpiringIn15Days++
		}
		if dto.DaysRemaining <= 30 {
			summary.ExpiringIn30Days++
		}
	}

	// Apply pagination
	totalCount := len(expiringDTOs)
	totalPages := (totalCount + query.Limit - 1) / query.Limit
	startIdx := (query.Page - 1) * query.Limit
	endIdx := startIdx + query.Limit
	if endIdx > totalCount {
		endIdx = totalCount
	}
	if startIdx >= totalCount {
		expiringDTOs = []resp.ExpiringLicenseDTO{}
	} else {
		expiringDTOs = expiringDTOs[startIdx:endIdx]
	}

	log.Info(sctx.Ctx, "Found %d expiring licenses", totalCount)

	return &resp.ExpiringLicensesResponse{
		StatusCodeAndMessage: port.LicenseExpiringListSuccess,
		Licenses:             expiringDTOs,
		Pagination: resp.PaginationMetadata{
			Page:       query.Page,
			Limit:      query.Limit,
			TotalCount: totalCount,
			TotalPages: totalPages,
		},
		Summary: summary,
	}, nil
}

// GetLicenseReminders retrieves reminder schedule for a license
// AGT-037: Get License Reminders
// BR-AGT-PRF-014: License Renewal Reminder Schedule
func (h *AgentLicenseHandler) GetLicenseReminders(sctx *serverRoute.Context, uri LicenseIDUri) (*resp.LicenseRemindersResponse, error) {
	log.Info(sctx.Ctx, "Fetching reminders for license: %s", uri.LicenseID)

	// Fetch license
	license, err := h.licenseRepo.FindByID(sctx.Ctx, uri.LicenseID)
	if err != nil {
		log.Error(sctx.Ctx, "License not found: %v", err)
		return nil, err
	}

	daysUntilExpiry := int(time.Until(license.RenewalDate).Hours() / 24)

	// BR-AGT-PRF-014: Reminder schedule at 30, 15, 7 days before and on expiry day
	reminderDays := []int{30, 15, 7, 0}
	reminders := make([]resp.ReminderScheduleDTO, 0)

	for _, daysBefore := range reminderDays {
		reminderDate := license.RenewalDate.AddDate(0, 0, -daysBefore)
		status := "PENDING"

		// If reminder date has passed
		if time.Now().After(reminderDate) {
			status = "SENT" // TODO: Check actual sent status from notifications table
		}

		reminders = append(reminders, resp.ReminderScheduleDTO{
			ReminderDate: reminderDate,
			DaysBefore:   daysBefore,
			Status:       status,
			SentAt:       nil, // TODO: Fetch from notifications table
		})
	}

	log.Info(sctx.Ctx, "Reminder schedule fetched for license: %s", uri.LicenseID)

	return &resp.LicenseRemindersResponse{
		StatusCodeAndMessage: port.LicenseReminderSuccess,
		LicenseID:            license.LicenseID,
		RenewalDate:          license.RenewalDate,
		DaysUntilExpiry:      daysUntilExpiry,
		Reminders:            reminders,
	}, nil
}

// BatchDeactivateExpired batch deactivates agents with expired licenses
// AGT-038: Batch Deactivate Expired Licenses
// FR-AGT-PRF-012: License Auto-Deactivation
// BR-AGT-PRF-013: Auto-Deactivation on License Expiry
func (h *AgentLicenseHandler) BatchDeactivateExpired(sctx *serverRoute.Context, request req.BatchDeactivateExpiredRequest) (*resp.BatchDeactivateExpiredResponse, error) {
	log.Info(sctx.Ctx, "Batch deactivating expired licenses (dry_run: %v)", request.DryRun)

	// Find all expired licenses
	expiredLicenses, err := h.licenseRepo.FindExpiredLicenses(sctx.Ctx)
	if err != nil {
		log.Error(sctx.Ctx, "Error finding expired licenses: %v", err)
		return nil, err
	}

	totalProcessed := len(expiredLicenses)
	affectedAgentIDs := make([]string, 0)

	// Collect unique agent IDs
	agentIDMap := make(map[string]bool)
	for _, license := range expiredLicenses {
		agentIDMap[license.AgentID] = true
	}
	for agentID := range agentIDMap {
		affectedAgentIDs = append(affectedAgentIDs, agentID)
	}

	if request.DryRun {
		log.Info(sctx.Ctx, "Dry run: Would deactivate %d expired licenses affecting %d agents", totalProcessed, len(affectedAgentIDs))

		return &resp.BatchDeactivateExpiredResponse{
			StatusCodeAndMessage: port.LicenseBatchDeactivateSuccess,
			Summary: resp.BatchDeactivationSummary{
				TotalProcessed:      totalProcessed,
				SuccessfullyUpdated: 0,
				Failed:              0,
				AffectedAgentIDs:    affectedAgentIDs,
				ProcessedAt:         time.Now(),
			},
			DryRun:  true,
			Message: fmt.Sprintf("Dry run completed. Would deactivate %d licenses", totalProcessed),
		}, nil
	}

	// Actual deactivation - use batch operation
	licenseIDs := make([]string, len(expiredLicenses))
	for i, license := range expiredLicenses {
		licenseIDs[i] = license.LicenseID
	}

	// Use repository batch method for efficient processing
	err = h.licenseRepo.BatchMarkAsExpired(sctx.Ctx, licenseIDs, request.ProcessedBy)
	successCount := totalProcessed
	failedCount := 0

	if err != nil {
		log.Error(sctx.Ctx, "Error batch marking licenses as expired: %v", err)
		// Try individual marking if batch fails
		successCount = 0
		failedCount = 0
		for _, licenseID := range licenseIDs {
			err := h.licenseRepo.MarkAsExpired(sctx.Ctx, licenseID, request.ProcessedBy)
			if err != nil {
				failedCount++
				log.Error(sctx.Ctx, "Failed to mark license %s as expired: %v", licenseID, err)
			} else {
				successCount++
			}
		}
	}

	log.Info(sctx.Ctx, "Batch deactivation completed: %d successful, %d failed", successCount, failedCount)

	// TODO: Trigger additional actions:
	// 1. Update agent status to DEACTIVATED
	// 2. Disable portal access
	// 3. Stop commission processing
	// 4. Send notifications

	return &resp.BatchDeactivateExpiredResponse{
		StatusCodeAndMessage: port.LicenseBatchDeactivateSuccess,
		Summary: resp.BatchDeactivationSummary{
			TotalProcessed:      totalProcessed,
			SuccessfullyUpdated: successCount,
			Failed:              failedCount,
			AffectedAgentIDs:    affectedAgentIDs,
			ProcessedAt:         time.Now(),
		},
		DryRun:  false,
		Message: fmt.Sprintf("Deactivated %d licenses successfully, %d failed", successCount, failedCount),
	}, nil
}

// Helper methods

// toLicenseDTO converts domain license to DTO with computed fields
func (h *AgentLicenseHandler) toLicenseDTO(license domain.AgentLicense) resp.LicenseDTO {
	now := time.Now()
	daysUntilExpiry := int(time.Until(license.RenewalDate).Hours() / 24)

	expiryStatus := "VALID"
	if daysUntilExpiry < 0 {
		expiryStatus = "EXPIRED"
	} else if daysUntilExpiry <= 30 {
		expiryStatus = "EXPIRING_SOON"
	}

	canRenew, _ := h.canRenewLicense(license)

	return resp.LicenseDTO{
		LicenseID:                   license.LicenseID,
		AgentID:                     license.AgentID,
		LicenseLine:                 license.LicenseLine,
		LicenseType:                 license.LicenseType,
		LicenseNumber:               license.LicenseNumber,
		ResidentStatus:              license.ResidentStatus,
		LicenseDate:                 license.LicenseDate,
		RenewalDate:                 license.RenewalDate,
		AuthorityDate:               license.AuthorityDate,
		RenewalCount:                license.RenewalCount,
		LicenseStatus:               license.LicenseStatus,
		IsPrimary:                   license.IsPrimary,
		LicentiateExamPassed:        license.LicentiateExamPassed,
		LicentiateExamDate:          license.LicentiateExamDate,
		LicentiateCertificateNumber: license.LicentiateCertificateNumber,
		Metadata:                    license.Metadata,
		DaysUntilExpiry:             daysUntilExpiry,
		ExpiryStatus:                expiryStatus,
		CanRenew:                    canRenew,
		CreatedAt:                   license.CreatedAt,
		UpdatedAt:                   license.UpdatedAt,
	}
}

// calculateRenewalDate calculates renewal date based on license type
// BR-AGT-PRF-012: Provisional = 1 year, Permanent (after exam) = 5 years
func (h *AgentLicenseHandler) calculateRenewalDate(licenseType string, licenseDate time.Time, examPassed bool) time.Time {
	if licenseType == domain.LicenseTypePermanent && examPassed {
		// Permanent license after exam: 5 years
		return licenseDate.AddDate(5, 0, 0)
	}
	// Provisional or permanent renewal: 1 year
	return licenseDate.AddDate(1, 0, 0)
}

// canRenewLicense determines if a license can be renewed and why
func (h *AgentLicenseHandler) canRenewLicense(license domain.AgentLicense) (bool, string) {
	// Cannot renew expired licenses
	if license.LicenseStatus == domain.LicenseStatusExpired {
		return false, "License is expired"
	}

	// Cannot renew terminated licenses
	if license.LicenseStatus == domain.LicenseStatusTerminated {
		return false, "License is terminated"
	}

	// Cannot renew suspended licenses
	if license.LicenseStatus == domain.LicenseStatusSuspended {
		return false, "License is suspended"
	}

	// Check provisional renewal limits
	if license.LicenseType == domain.LicenseTypeProvisional {
		if license.RenewalCount >= 2 {
			return false, "Provisional license can only be renewed 2 times. Must pass exam to convert to permanent."
		}
		// Check if within 3 years for exam eligibility
		yearsSinceLicense := time.Since(license.LicenseDate).Hours() / 24 / 365
		if yearsSinceLicense > 3 && !license.LicentiateExamPassed {
			return false, "Must pass exam within 3 years of provisional license"
		}
	}

	return true, "License is eligible for renewal"
}

// buildRenewalHistory builds renewal history from audit logs
func (h *AgentLicenseHandler) buildRenewalHistory(auditLogs []domain.AgentAuditLog, licenseID string) []resp.RenewalRecordDTO {
	// TODO: Implement proper parsing of audit logs to extract renewal events
	// This requires correlating multiple audit log entries for a single renewal
	return []resp.RenewalRecordDTO{}
}
