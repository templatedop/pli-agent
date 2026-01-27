package handler

import (
	"encoding/json"
	"fmt"
	"time"

	"pli-agent-api/core/domain"
	"pli-agent-api/core/port"
	resp "pli-agent-api/handler/response"
	repo "pli-agent-api/repo/postgres"

	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"
)

// AgentProfileUpdateHandler handles profile update and search APIs
// AGT-022 to AGT-028: Profile Update & Search
type AgentProfileUpdateHandler struct {
	*serverHandler.Base
	profileRepo       *repo.AgentProfileRepository
	auditLogRepo      *repo.AgentAuditLogRepository
	updateRequestRepo *repo.AgentProfileUpdateRequestRepository
	fieldMetadataRepo *repo.AgentProfileFieldMetadataRepository
}

// NewAgentProfileUpdateHandler creates a new profile update handler
func NewAgentProfileUpdateHandler(
	profileRepo *repo.AgentProfileRepository,
	auditLogRepo *repo.AgentAuditLogRepository,
	updateRequestRepo *repo.AgentProfileUpdateRequestRepository,
	fieldMetadataRepo *repo.AgentProfileFieldMetadataRepository,
) *AgentProfileUpdateHandler {
	base := serverHandler.New("Agent Profile Update & Search APIs").SetPrefix("/v1").AddPrefix("")
	return &AgentProfileUpdateHandler{
		Base:              base,
		profileRepo:       profileRepo,
		auditLogRepo:      auditLogRepo,
		updateRequestRepo: updateRequestRepo,
		fieldMetadataRepo: fieldMetadataRepo,
	}
}

// Routes defines profile update and search routes
func (h *AgentProfileUpdateHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		// Search & View APIs
		serverRoute.GET("/agents/search", h.SearchAgents).Name("Search Agents"),
		serverRoute.GET("/agents/:agent_id", h.GetAgentProfile).Name("Get Agent Profile Details"),
		serverRoute.GET("/agents/:agent_id/update-form", h.GetUpdateForm).Name("Get Update Form"),
		serverRoute.GET("/agents/:agent_id/audit-history", h.GetAuditHistory).Name("Get Audit History"),

		// Update APIs
		serverRoute.PUT("/agents/:agent_id/sections/:section", h.UpdateProfileSection).Name("Update Profile Section"),

		// Approval APIs (placeholders for now - full workflow in Phase 6.1)
		serverRoute.PUT("/approvals/:approval_request_id/approve", h.ApproveProfileUpdate).Name("Approve Profile Update"),
		serverRoute.PUT("/approvals/:approval_request_id/reject", h.RejectProfileUpdate).Name("Reject Profile Update"),
	}
}

// SearchAgents performs multi-criteria agent search with pagination
// AGT-022: Search Agents
// FR-AGT-PRF-004: Multi-criteria agent search
// BR-AGT-PRF-022: Multi-Criteria Agent Search
func (h *AgentProfileUpdateHandler) SearchAgents(sctx *serverRoute.Context, req SearchAgentsRequest) (*resp.SearchAgentsResponse, error) {
	log.Info(sctx.Ctx, "Searching agents with filters: %+v", req)

	// Perform search
	profiles, totalCount, err := h.profileRepo.Search(
		sctx.Ctx,
		req.AgentID,
		req.Name,
		req.PANNumber,
		req.MobileNumber,
		req.Email,
		req.Status,
		req.OfficeCode,
		req.Page,
		req.Limit,
	)
	if err != nil {
		log.Error(sctx.Ctx, "Error searching agents: %v", err)
		return nil, err
	}

	// Transform to search results
	results := make([]resp.AgentSearchResult, len(profiles))
	for i, p := range profiles {
		results[i] = resp.AgentSearchResult{
			AgentID:   p.AgentID,
			Name:      fmt.Sprintf("%s %s", p.FirstName, p.LastName),
			AgentType: p.AgentType,
			PAN:       p.PANNumber,
			Status:    p.Status,
			Office:    p.OfficeCode,
		}
	}

	// Calculate pagination
	totalPages := (totalCount + req.Limit - 1) / req.Limit
	pagination := resp.PaginationMetadata{
		CurrentPage:    req.Page,
		TotalPages:     totalPages,
		TotalResults:   totalCount,
		ResultsPerPage: req.Limit,
	}

	log.Info(sctx.Ctx, "Found %d agents (page %d/%d)", totalCount, req.Page, totalPages)

	return &resp.SearchAgentsResponse{
		StatusCodeAndMessage: port.SearchSuccess,
		Results:              results,
		Pagination:           pagination,
	}, nil
}

// GetAgentProfile retrieves complete agent profile with all related entities
// AGT-023: Get Agent Profile Details
// FR-AGT-PRF-005: Profile Dashboard View
func (h *AgentProfileUpdateHandler) GetAgentProfile(sctx *serverRoute.Context, req AgentIDUri) (*resp.AgentProfileResponse, error) {
	log.Info(sctx.Ctx, "Fetching profile for agent: %s", req.AgentID)

	// Get profile with related entities (single query)
	profile, addresses, contacts, emails, err := h.profileRepo.GetProfileWithRelatedEntities(sctx.Ctx, req.AgentID)
	if err != nil {
		log.Error(sctx.Ctx, "Error fetching profile: %v", err)
		return nil, err
	}

	// Build full name handling null middle name
	fullName := profile.FirstName
	if profile.MiddleName.Valid && profile.MiddleName.String != "" {
		fullName += " " + profile.MiddleName.String
	}
	fullName += " " + profile.LastName

	// Transform to DTO with proper type conversions
	dob := profile.DateOfBirth // Convert time.Time to *time.Time
	profileDTO := resp.AgentProfileDTO{
		AgentID:     profile.AgentID,
		ProfileType: profile.AgentType,
		FullName:    fullName,
		PANNumber:   profile.PANNumber,
		Status:      profile.Status,
		PersonalInfo: resp.PersonalInfoDTO{
			FirstName:     profile.FirstName,
			MiddleName:    profile.MiddleName.String,
			LastName:      profile.LastName,
			DateOfBirth:   &dob,
			Gender:        profile.Gender,
			AadharNumber:  profile.AadharNumber.String,
			MaritalStatus: profile.MaritalStatus.String,
			Category:      profile.Category.String,
			Title:         profile.Title.String,
		},
		Addresses: transformAddresses(addresses),
		Contacts:  transformContacts(contacts),
		Emails:    transformEmails(emails),
	}

	// Add office info if available
	if profile.OfficeCode != "" {
		profileDTO.Office = &resp.OfficeDTO{
			OfficeCode: profile.OfficeCode,
			OfficeName: profile.OfficeCode, // TODO: Fetch actual office name
		}
	}

	log.Info(sctx.Ctx, "Profile fetched successfully for agent: %s", req.AgentID)

	return &resp.AgentProfileResponse{
		StatusCodeAndMessage: port.ProfileFetchSuccess,
		AgentProfile:         profileDTO,
	}, nil
}

// GetUpdateForm returns pre-populated update form with metadata
// AGT-024: Get Update Form
// Phase 6.2: Dynamic field metadata from database
func (h *AgentProfileUpdateHandler) GetUpdateForm(sctx *serverRoute.Context, req AgentIDUri) (*resp.UpdateFormResponse, error) {
	log.Info(sctx.Ctx, "Fetching update form for agent: %s", req.AgentID)

	// Get current profile
	profile, err := h.profileRepo.FindByID(sctx.Ctx, req.AgentID)
	if err != nil {
		log.Error(sctx.Ctx, "Error fetching profile: %v", err)
		return nil, err
	}

	// Fetch all active field metadata from database
	allFieldMetadata, err := h.fieldMetadataRepo.GetAll(sctx.Ctx)
	if err != nil {
		log.Error(sctx.Ctx, "Error fetching field metadata: %v", err)
		return nil, err
	}

	// Group fields by section and build response structures
	sectionFieldsMap := make(map[string][]string)
	editableFields := make(map[string]resp.FieldMetadata)

	for _, fieldMeta := range allFieldMetadata {
		// Add field to section
		sectionFieldsMap[fieldMeta.Section] = append(sectionFieldsMap[fieldMeta.Section], fieldMeta.FieldName)

		// Parse validation rules from JSONB
		var validationRules map[string]interface{}
		if fieldMeta.ValidationRules.Valid && fieldMeta.ValidationRules.String != "" {
			if err := json.Unmarshal([]byte(fieldMeta.ValidationRules.String), &validationRules); err != nil {
				log.Warn(sctx.Ctx, "Failed to parse validation rules for field %s: %v", fieldMeta.FieldName, err)
				validationRules = make(map[string]interface{})
			}
		}

		// Parse select options from JSONB
		var selectOptions []map[string]interface{}
		if fieldMeta.SelectOptions.Valid && fieldMeta.SelectOptions.String != "" {
			if err := json.Unmarshal([]byte(fieldMeta.SelectOptions.String), &selectOptions); err != nil {
				log.Warn(sctx.Ctx, "Failed to parse select options for field %s: %v", fieldMeta.FieldName, err)
			}
		}

		// Build field metadata response
		fieldMetaResp := resp.FieldMetadata{
			Name:             fieldMeta.FieldName,
			DisplayName:      fieldMeta.DisplayName,
			Type:             fieldMeta.FieldType,
			Required:         fieldMeta.IsRequired,
			Editable:         fieldMeta.IsEditable,
			RequiresApproval: fieldMeta.RequiresApproval,
			ValidationRules:  validationRules,
			SelectOptions:    selectOptions,
		}

		// Add optional fields
		if fieldMeta.Placeholder.Valid {
			fieldMetaResp.Placeholder = fieldMeta.Placeholder.String
		}
		if fieldMeta.HelpText.Valid {
			fieldMetaResp.HelpText = fieldMeta.HelpText.String
		}

		editableFields[fieldMeta.FieldName] = fieldMetaResp
	}

	// Build sections with display names
	sectionDisplayNames := map[string]string{
		"personal_info": "Personal Information",
		"address":       "Address Information",
		"contact":       "Contact Information",
		"email":         "Email Information",
		"bank":          "Bank Details",
		"license":       "License Information",
	}

	sections := make([]resp.SectionDTO, 0)
	for section, fields := range sectionFieldsMap {
		displayName := sectionDisplayNames[section]
		if displayName == "" {
			displayName = section
		}
		sections = append(sections, resp.SectionDTO{
			Name:        section,
			DisplayName: displayName,
			Fields:      fields,
		})
	}

	// Build current data map dynamically from profile
	currentData := map[string]interface{}{
		// Personal info
		"title":              profile.Title,
		"first_name":         profile.FirstName,
		"middle_name":        profile.MiddleName,
		"last_name":          profile.LastName,
		"date_of_birth":      profile.DateOfBirth,
		"gender":             profile.Gender,
		"marital_status":     profile.MaritalStatus,
		"category":           profile.Category,
		"pan_number":         profile.PANNumber,
		"aadhar_number":      profile.AadharNumber,
		"professional_title": profile.ProfessionalTitle,
		// TODO: Add other sections (address, contact, email, bank) when needed
	}

	log.Info(sctx.Ctx, "Update form fetched successfully for agent: %s with %d fields across %d sections",
		req.AgentID, len(editableFields), len(sections))

	return &resp.UpdateFormResponse{
		StatusCodeAndMessage: port.FormFetchSuccess,
		AgentID:              profile.AgentID,
		Sections:             sections,
		CurrentData:          currentData,
		EditableFields:       editableFields,
	}, nil
}

// UpdateProfileSection updates a profile section
// AGT-025: Update Profile Section
// FR-AGT-PRF-006: Personal Information Update
// BR-AGT-PRF-005: Name Update with Audit Logging
// Phase 6.2: Dynamic approval logic based on field metadata
func (h *AgentProfileUpdateHandler) UpdateProfileSection(sctx *serverRoute.Context, req UpdateSectionRequest) (*resp.UpdateSectionResponse, error) {
	log.Info(sctx.Ctx, "Updating section %s for agent: %s", req.Section, req.AgentID)

	// Fetch critical fields from database (fields that require approval)
	criticalFieldsMetadata, err := h.fieldMetadataRepo.GetCriticalFields(sctx.Ctx)
	if err != nil {
		log.Error(sctx.Ctx, "Error fetching critical fields metadata: %v", err)
		return nil, err
	}

	// Build map of critical fields for quick lookup
	criticalFields := make(map[string]bool)
	for _, fieldMeta := range criticalFieldsMetadata {
		criticalFields[fieldMeta.FieldName] = true
	}

	// Check if any critical fields are being updated
	requiresApproval := false
	for field := range req.Updates {
		if criticalFields[field] {
			requiresApproval = true
			break
		}
	}

	// If requires approval, create approval request
	if requiresApproval {
		log.Info(sctx.Ctx, "Update requires approval for agent: %s", req.AgentID)

		// Create approval request
		updateRequest, err := h.updateRequestRepo.Create(
			sctx.Ctx,
			req.AgentID,
			req.Section,
			req.Updates,
			req.Reason,
			req.UpdatedBy,
		)
		if err != nil {
			log.Error(sctx.Ctx, "Error creating approval request: %v", err)
			return nil, err
		}

		updatedFields := make([]string, 0, len(req.Updates))
		for field := range req.Updates {
			updatedFields = append(updatedFields, field)
		}

		// Build changed fields preview
		changedFields := make(map[string]resp.ChangeInfo)
		for field, newValue := range req.Updates {
			changedFields[field] = resp.ChangeInfo{
				OldValue: "", // Will be applied on approval
				NewValue: fmt.Sprintf("%v", newValue),
			}
		}

		log.Info(sctx.Ctx, "Approval request created: %s", updateRequest.RequestID)

		return &resp.UpdateSectionResponse{
			StatusCodeAndMessage: port.PendingApproval,
			AgentID:              req.AgentID,
			Section:              req.Section,
			Status:               domain.UpdateRequestStatusPending,
			UpdatedFields:        updatedFields,
			ApprovalRequired:     true,
			ApprovalRequestID:    &updateRequest.RequestID,
			ChangedFields:        changedFields,
		}, nil
	}

	// Non-critical fields - update directly
	updatedProfile, err := h.profileRepo.UpdateSectionReturning(
		sctx.Ctx,
		req.AgentID,
		req.Updates,
		req.UpdatedBy,
	)
	if err != nil {
		log.Error(sctx.Ctx, "Error updating profile section: %v", err)
		return nil, err
	}

	// Build changed fields map
	changedFields := make(map[string]resp.ChangeInfo)
	for field, newValue := range req.Updates {
		changedFields[field] = resp.ChangeInfo{
			OldValue: "old_value", // TODO: Get from old profile
			NewValue: fmt.Sprintf("%v", newValue),
		}
	}

	updatedFields := make([]string, 0, len(req.Updates))
	for field := range req.Updates {
		updatedFields = append(updatedFields, field)
	}

	log.Info(sctx.Ctx, "Profile section updated successfully for agent: %s", req.AgentID)

	return &resp.UpdateSectionResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		AgentID:              req.AgentID,
		Section:              req.Section,
		Status:               "UPDATED",
		UpdatedFields:        updatedFields,
		ApprovalRequired:     false,
		UpdatedProfile: &resp.AgentProfileDTO{
			AgentID:     updatedProfile.AgentID,
			ProfileType: updatedProfile.AgentType,
			FullName:    fmt.Sprintf("%s %s %s", updatedProfile.FirstName, updatedProfile.MiddleName, updatedProfile.LastName),
			PANNumber:   updatedProfile.PANNumber,
			Status:      updatedProfile.Status,
		},
		ChangedFields: changedFields,
	}, nil
}

// ApproveProfileUpdate approves a profile update request and applies the changes
// AGT-026: Approve Profile Update
// BR-AGT-PRF-005: Name Update with Audit Logging
// BR-AGT-PRF-006: PAN Update with Validation
// ULTIMATE OPTIMIZATION: 1 database call (stored function does everything)
func (h *AgentProfileUpdateHandler) ApproveProfileUpdate(sctx *serverRoute.Context, req ApprovalRequest) (*resp.ApprovalResponse, error) {
	log.Info(sctx.Ctx, "Approving profile update request: %s", req.ApprovalRequestID)

	// Single database call: approve + update + audit logs
	updatedProfile, approvedRequest, err := h.profileRepo.ApproveRequestAndUpdateProfile(
		sctx.Ctx,
		req.ApprovalRequestID,
		req.ApprovedBy,
		req.Comments,
	)
	if err != nil {
		log.Error(sctx.Ctx, "Error approving and applying updates: %v", err)
		return nil, err
	}

	log.Info(sctx.Ctx, "Profile update approved and applied for agent: %s", approvedRequest.AgentID)

	return &resp.ApprovalResponse{
		StatusCodeAndMessage: port.ApprovalSuccess,
		ApprovalRequestID:    req.ApprovalRequestID,
		Status:               domain.UpdateRequestStatusApproved,
		AgentID:              approvedRequest.AgentID,
		ApprovedBy:           req.ApprovedBy,
		ProcessedAt:          time.Now(),
		Message:              "Profile update approved and applied successfully",
		UpdatedProfile: &resp.AgentProfileDTO{
			AgentID:     updatedProfile.AgentID,
			ProfileType: updatedProfile.AgentType,
			PANNumber:   updatedProfile.PANNumber,
			Status:      updatedProfile.Status,
		},
	}, nil
}

// RejectProfileUpdate rejects a profile update request
// AGT-027: Reject Profile Update
// BR-AGT-PRF-005: Name Update with Audit Logging (rejected requests also logged)
// OPTIMIZED: 1 database call (fetch+reject in single query)
func (h *AgentProfileUpdateHandler) RejectProfileUpdate(sctx *serverRoute.Context, req ApprovalRequest) (*resp.ApprovalResponse, error) {
	log.Info(sctx.Ctx, "Rejecting profile update request: %s", req.ApprovalRequestID)

	// Reject request and get rejected request data in SINGLE database call
	rejectedRequest, err := h.updateRequestRepo.RejectAndReturn(
		sctx.Ctx,
		req.ApprovalRequestID,
		req.RejectedBy,
		req.Comments,
	)
	if err != nil {
		log.Error(sctx.Ctx, "Error rejecting request: %v", err)
		return nil, err
	}

	log.Info(sctx.Ctx, "Profile update rejected for agent: %s", rejectedRequest.AgentID)

	return &resp.ApprovalResponse{
		StatusCodeAndMessage: port.RejectionSuccess,
		ApprovalRequestID:    req.ApprovalRequestID,
		Status:               domain.UpdateRequestStatusRejected,
		AgentID:              rejectedRequest.AgentID,
		RejectedBy:           req.RejectedBy,
		ProcessedAt:          time.Now(),
		Message:              "Profile update rejected - changes not applied",
	}, nil
}

// GetAuditHistory retrieves audit history with pagination
// AGT-028: Get Audit History
// FR-AGT-PRF-022: Profile Change History and Audit Trail
func (h *AgentProfileUpdateHandler) GetAuditHistory(sctx *serverRoute.Context, req AuditHistoryRequest) (*resp.AuditHistoryResponse, error) {
	log.Info(sctx.Ctx, "Fetching audit history for agent: %s", req.AgentID)

	// Parse date filters
	var fromDate, toDate *time.Time
	if req.FromDate != nil && *req.FromDate != "" {
		parsed, err := time.Parse("2006-01-02", *req.FromDate)
		if err == nil {
			fromDate = &parsed
		}
	}
	if req.ToDate != nil && *req.ToDate != "" {
		parsed, err := time.Parse("2006-01-02", *req.ToDate)
		if err == nil {
			toDate = &parsed
		}
	}

	// Get audit history
	auditLogs, totalCount, err := h.auditLogRepo.GetHistory(
		sctx.Ctx,
		req.AgentID,
		fromDate,
		toDate,
		req.Page,
		req.Limit,
	)
	if err != nil {
		log.Error(sctx.Ctx, "Error fetching audit history: %v", err)
		return nil, err
	}

	// Transform to DTOs
	auditLogDTOs := make([]resp.AuditLogDTO, len(auditLogs))
	for i, log := range auditLogs {
		auditLogDTOs[i] = resp.AuditLogDTO{
			AuditID:     log.AuditID,
			Action:      log.ActionType,
			FieldName:   log.FieldName,
			OldValue:    log.OldValue,
			NewValue:    log.NewValue,
			PerformedBy: log.PerformedBy,
			PerformedAt: log.PerformedAt,
			IPAddress:   log.IPAddress,
		}
	}

	// Calculate pagination
	totalPages := (totalCount + req.Limit - 1) / req.Limit
	pagination := resp.PaginationMetadata{
		CurrentPage:    req.Page,
		TotalPages:     totalPages,
		TotalResults:   totalCount,
		ResultsPerPage: req.Limit,
	}

	log.Info(sctx.Ctx, "Audit history fetched: %d logs (page %d/%d)", totalCount, req.Page, totalPages)

	return &resp.AuditHistoryResponse{
		StatusCodeAndMessage: port.AuditHistorySuccess,
		AgentID:              req.AgentID,
		AuditLogs:            auditLogDTOs,
		Pagination:           pagination,
	}, nil
}

// Helper functions to transform domain models to DTOs

func transformAddresses(addresses []domain.AgentAddress) []resp.AddressDTO {
	result := make([]resp.AddressDTO, len(addresses))
	for i, addr := range addresses {
		result[i] = resp.AddressDTO{
			AddressID:   addr.AddressID,
			AddressType: addr.AddressType,
			Line1:       addr.AddressLine1,
			Line2:       addr.AddressLine2,
			Line3:       addr.AddressLine3,
			City:        addr.City,
			District:    addr.District,
			State:       addr.State,
			Country:     addr.Country,
			Pincode:     addr.Pincode,
			IsPrimary:   addr.IsPrimary,
		}
	}
	return result
}

func transformContacts(contacts []domain.AgentContact) []resp.ContactDTO {
	result := make([]resp.ContactDTO, len(contacts))
	for i, contact := range contacts {
		result[i] = resp.ContactDTO{
			ContactID:       contact.ContactID,
			ContactType:     contact.ContactType,
			MobileNumber:    contact.MobileNumber,
			AlternateNumber: contact.AlternateNumber,
			IsPrimary:       contact.IsPrimary,
			IsVerified:      contact.IsVerified,
		}
	}
	return result
}

func transformEmails(emails []domain.AgentEmail) []resp.EmailDTO {
	result := make([]resp.EmailDTO, len(emails))
	for i, email := range emails {
		result[i] = resp.EmailDTO{
			EmailID:      email.EmailID,
			EmailType:    email.EmailType,
			EmailAddress: email.EmailAddress,
			IsPrimary:    email.IsPrimary,
			IsVerified:   email.IsVerified,
		}
	}
	return result
}
