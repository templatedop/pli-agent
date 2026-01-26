package handler

import (
	"encoding/json"
	"fmt"

	"pli-agent-api/core/domain"
	"pli-agent-api/core/port"
	resp "pli-agent-api/handler/response"
	repo "pli-agent-api/repo/postgres"

	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"
)

// AgentProfileUpdateHandler handles profile update and approval APIs
// AGT-022 to AGT-028: Profile Update Journey
type AgentProfileUpdateHandler struct {
	*serverHandler.Base
	profileRepo  *repo.AgentProfileRepository
	approvalRepo *repo.ApprovalRequestRepository
	auditRepo    *repo.AgentAuditLogRepository
}

// NewAgentProfileUpdateHandler creates a new profile update handler
func NewAgentProfileUpdateHandler(
	profileRepo *repo.AgentProfileRepository,
	approvalRepo *repo.ApprovalRequestRepository,
	auditRepo *repo.AgentAuditLogRepository,
) *AgentProfileUpdateHandler {
	base := serverHandler.New("Agent Profile Update & Approval APIs").SetPrefix("/v1").AddPrefix("")
	return &AgentProfileUpdateHandler{
		Base:         base,
		profileRepo:  profileRepo,
		approvalRepo: approvalRepo,
		auditRepo:    auditRepo,
	}
}

// Routes defines profile update and approval routes
func (h *AgentProfileUpdateHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		// Agent Search & View (AGT-022 to AGT-024)
		serverRoute.GET("/agents/search", h.SearchAgents).Name("Search Agents"),
		serverRoute.GET("/agents/:agent_id", h.GetAgentProfile).Name("Get Agent Profile"),
		serverRoute.GET("/agents/:agent_id/update-form", h.GetUpdateForm).Name("Get Profile Update Form"),

		// Profile Update (AGT-025)
		serverRoute.PUT("/agents/:agent_id/sections/:section", h.UpdateProfileSection).Name("Update Profile Section"),

		// Approval Workflow (AGT-026, AGT-027)
		serverRoute.PUT("/approvals/:approval_request_id/approve", h.ApproveProfileUpdate).Name("Approve Profile Update"),
		serverRoute.PUT("/approvals/:approval_request_id/reject", h.RejectProfileUpdate).Name("Reject Profile Update"),

		// Audit History (AGT-028)
		serverRoute.GET("/agents/:agent_id/audit-history", h.GetAuditHistory).Name("Get Audit History"),
	}
}

// SearchAgents performs multi-criteria agent search
// AGT-022: Multi-criteria Agent Search
// FR-AGT-PRF-004: Agent Search Functionality
// BR-AGT-PRF-022: Multi-Criteria Search Support
func (h *AgentProfileUpdateHandler) SearchAgents(sctx *serverRoute.Context, req SearchAgentsQuery) (*resp.AgentSearchResponse, error) {
	log.Info(sctx.Ctx, "Searching agents with criteria: name=%s, pan=%s, status=%s", req.Name, req.PANNumber, req.Status)

	// Default pagination
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 20
	}

	// Execute search - single database round trip
	agents, totalCount, err := h.profileRepo.SearchAgents(
		sctx.Ctx,
		req.AgentID,
		req.Name,
		req.PANNumber,
		req.MobileNumber,
		req.Status,
		req.OfficeCode,
		req.Page,
		req.Limit,
	)
	if err != nil {
		log.Error(sctx.Ctx, "Error searching agents: %v", err)
		return nil, err
	}

	// Build response
	agentList := make([]resp.AgentSummary, 0, len(agents))
	for _, a := range agents {
		agentList = append(agentList, resp.AgentSummary{
			AgentID:      a.AgentID,
			AgentCode:    a.AgentCode.String,
			FirstName:    a.FirstName,
			MiddleName:   a.MiddleName.String,
			LastName:     a.LastName,
			AgentType:    string(a.AgentType),
			Status:       string(a.Status),
			OfficeCode:   a.OfficeCode,
			PANNumber:    a.PANNumber,
			MobileNumber: "", // Will be populated from contacts in real implementation
		})
	}

	log.Info(sctx.Ctx, "Found %d agents (total: %d)", len(agentList), totalCount)

	return &resp.AgentSearchResponse{
		StatusCodeAndMessage: port.ListSuccess,
		MetaDataResponse:     port.NewMetaDataResponse((req.Page-1)*req.Limit, req.Limit, uint64(totalCount)),
		Data:                 agentList,
	}, nil
}

// GetAgentProfile retrieves complete agent profile with related data
// AGT-023: Get Agent Profile Details
// FR-AGT-PRF-005: Profile Dashboard View
// BR-AGT-PRF-023: Dashboard Profile View
func (h *AgentProfileUpdateHandler) GetAgentProfile(sctx *serverRoute.Context, req AgentIDURI) (*resp.AgentProfileDetailsResponse, error) {
	log.Info(sctx.Ctx, "Fetching agent profile for ID: %s", req.AgentID)

	// Fetch profile with related data - single database round trip using JSON aggregation
	profileData, err := h.profileRepo.GetAgentProfileWithDetails(sctx.Ctx, req.AgentID)
	if err != nil {
		log.Error(sctx.Ctx, "Error fetching agent profile: %v", err)
		return nil, err
	}

	// Extract profile, addresses, contacts, emails from the aggregated result
	var profile domain.AgentProfile
	var addresses []domain.AgentAddress
	var contacts []domain.AgentContact
	var emails []domain.AgentEmail

	// Parse the profile section
	if profileMap, ok := profileData["profile"].(map[string]interface{}); ok {
		profileJSON, _ := json.Marshal(profileMap)
		json.Unmarshal(profileJSON, &profile)
	}

	// Parse addresses
	if addressesData, ok := profileData["addresses"].([]interface{}); ok {
		addressesJSON, _ := json.Marshal(addressesData)
		json.Unmarshal(addressesJSON, &addresses)
	}

	// Parse contacts
	if contactsData, ok := profileData["contacts"].([]interface{}); ok {
		contactsJSON, _ := json.Marshal(contactsData)
		json.Unmarshal(contactsJSON, &contacts)
	}

	// Parse emails
	if emailsData, ok := profileData["emails"].([]interface{}); ok {
		emailsJSON, _ := json.Marshal(emailsData)
		json.Unmarshal(emailsJSON, &emails)
	}

	log.Info(sctx.Ctx, "Successfully fetched profile for agent: %s %s", profile.FirstName, profile.LastName)

	return &resp.AgentProfileDetailsResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		Profile:              profile,
		Addresses:            addresses,
		Contacts:             contacts,
		Emails:               emails,
	}, nil
}

// GetUpdateForm retrieves form data for updating a specific section
// AGT-024: Get Profile Update Form
// FR-AGT-PRF-006: Personal Information Update
func (h *AgentProfileUpdateHandler) GetUpdateForm(sctx *serverRoute.Context, req GetUpdateFormRequest) (*resp.UpdateFormResponse, error) {
	log.Info(sctx.Ctx, "Fetching update form for agent: %s, section: %s", req.AgentID, req.Section)

	// Validate section
	validSections := map[string]bool{
		"personal_info": true,
		"address":       true,
		"contact":       true,
		"license":       true,
		"bank_details":  true,
		"status":        true,
	}

	if !validSections[req.Section] {
		return nil, fmt.Errorf("invalid section: %s. Valid sections: personal_info, address, contact, license, bank_details, status", req.Section)
	}

	// Fetch form data - single database round trip
	formData, err := h.profileRepo.GetAgentUpdateFormData(sctx.Ctx, req.AgentID, req.Section)
	if err != nil {
		log.Error(sctx.Ctx, "Error fetching update form: %v", err)
		return nil, err
	}

	log.Info(sctx.Ctx, "Successfully fetched update form for section: %s", req.Section)

	return &resp.UpdateFormResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		Section:              req.Section,
		FormData:             formData,
		EditableFields:       getEditableFieldsForSection(req.Section),
		CriticalFields:       getCriticalFieldsForSection(req.Section),
	}, nil
}

// UpdateProfileSection updates a specific section of agent profile
// AGT-025: Update Profile Section
// FR-AGT-PRF-006: Personal Information Update
// BR-AGT-PRF-005: Name Update with Audit Logging
// BR-AGT-PRF-006: PAN Update with Format and Uniqueness Validation
func (h *AgentProfileUpdateHandler) UpdateProfileSection(sctx *serverRoute.Context, req UpdateProfileSectionRequest) (*resp.UpdateProfileResponse, error) {
	log.Info(sctx.Ctx, "Updating profile section for agent: %s, section: %s", req.AgentID, req.Section)

	// Check if update requires approval
	requiresApproval := false
	criticalFields := getCriticalFieldsForSection(req.Section)
	for field := range req.Changes {
		for _, criticalField := range criticalFields {
			if field == criticalField {
				requiresApproval = true
				break
			}
		}
	}

	if requiresApproval {
		// Create approval request
		changesJSON, _ := json.Marshal(req.Changes)
		approvalReq := domain.ApprovalRequest{
			AgentID:          req.AgentID,
			Section:          req.Section,
			RequestedChanges: string(changesJSON),
			RequestedBy:      req.UpdatedBy,
			Status:           domain.ApprovalStatusPending,
		}

		createdApproval, err := h.approvalRepo.Create(sctx.Ctx, approvalReq)
		if err != nil {
			log.Error(sctx.Ctx, "Error creating approval request: %v", err)
			return nil, err
		}

		log.Info(sctx.Ctx, "Created approval request: %s", createdApproval.ApprovalRequestID)

		return &resp.UpdateProfileResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: "APPROVAL_REQUIRED",
				Message:    "Update requires approval. Approval request created.",
			},
			ApprovalRequired:  true,
			ApprovalRequestID: &createdApproval.ApprovalRequestID,
			Status:            "PENDING_APPROVAL",
		}, nil
	}

	// Direct update without approval
	var updatedProfile *domain.AgentProfile
	var err error

	switch req.Section {
	case "personal_info":
		// Extract update map from changes
		updateMap := make(map[string]interface{})
		for field, changeObj := range req.Changes {
			if change, ok := changeObj.(map[string]interface{}); ok {
				if newValue, exists := change["new_value"]; exists {
					updateMap[field] = newValue
				}
			}
		}

		// Atomic update with RETURNING
		updatedProfile, err = h.profileRepo.UpdateAgentPersonalInfoReturning(
			sctx.Ctx,
			req.AgentID,
			updateMap,
			req.UpdatedBy,
		)
	default:
		return nil, fmt.Errorf("section update not implemented: %s", req.Section)
	}

	if err != nil {
		log.Error(sctx.Ctx, "Error updating profile section: %v", err)
		return nil, err
	}

	log.Info(sctx.Ctx, "Successfully updated profile section: %s, new version: %d", req.Section, updatedProfile.Version)

	return &resp.UpdateProfileResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		ApprovalRequired:     false,
		Status:               "UPDATED",
		UpdatedFields:        getUpdatedFieldsList(req.Changes),
		Version:              updatedProfile.Version,
	}, nil
}

// ApproveProfileUpdate approves a pending profile update request
// AGT-026: Approve Profile Update
// CRITICAL: Single atomic operation
func (h *AgentProfileUpdateHandler) ApproveProfileUpdate(sctx *serverRoute.Context, req ApprovalActionRequest) (*resp.ApprovalActionResponse, error) {
	log.Info(sctx.Ctx, "Approving profile update request: %s by %s", req.ApprovalRequestID, req.ReviewedBy)

	// Approve the request - single atomic UPDATE...RETURNING
	approvedRequest, err := h.approvalRepo.ApproveReturning(
		sctx.Ctx,
		req.ApprovalRequestID,
		req.ReviewedBy,
		req.ReviewComments,
	)
	if err != nil {
		log.Error(sctx.Ctx, "Error approving request: %v", err)
		return nil, err
	}

	// Apply the approved changes to agent profile
	var changes map[string]interface{}
	if err := json.Unmarshal([]byte(approvedRequest.RequestedChanges), &changes); err != nil {
		return nil, fmt.Errorf("failed to parse requested changes: %w", err)
	}

	// Extract update map
	updateMap := make(map[string]interface{})
	for field, changeObj := range changes {
		if change, ok := changeObj.(map[string]interface{}); ok {
			if newValue, exists := change["new_value"]; exists {
				updateMap[field] = newValue
			}
		}
	}

	// Apply updates based on section
	switch approvedRequest.Section {
	case "personal_info":
		_, err = h.profileRepo.UpdateAgentPersonalInfoReturning(
			sctx.Ctx,
			approvedRequest.AgentID,
			updateMap,
			req.ReviewedBy,
		)
	default:
		return nil, fmt.Errorf("section update not implemented: %s", approvedRequest.Section)
	}

	if err != nil {
		log.Error(sctx.Ctx, "Error applying approved changes: %v", err)
		return nil, err
	}

	log.Info(sctx.Ctx, "Successfully approved and applied profile update request: %s", req.ApprovalRequestID)

	return &resp.ApprovalActionResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: "APPROVAL_SUCCESS",
			Message:    "Profile update approved and applied successfully",
		},
		ApprovalRequestID: approvedRequest.ApprovalRequestID,
		Status:            string(approvedRequest.Status),
		ReviewedBy:        approvedRequest.ReviewedBy.String,
		ReviewedAt:        &approvedRequest.ReviewedAt.Time,
	}, nil
}

// RejectProfileUpdate rejects a pending profile update request
// AGT-027: Reject Profile Update
// CRITICAL: Single atomic operation
func (h *AgentProfileUpdateHandler) RejectProfileUpdate(sctx *serverRoute.Context, req ApprovalActionRequest) (*resp.ApprovalActionResponse, error) {
	log.Info(sctx.Ctx, "Rejecting profile update request: %s by %s", req.ApprovalRequestID, req.ReviewedBy)

	// Validate review comments are provided
	if req.ReviewComments == nil || *req.ReviewComments == "" {
		return nil, fmt.Errorf("review comments are required for rejection")
	}

	// Reject the request - single atomic UPDATE...RETURNING
	rejectedRequest, err := h.approvalRepo.RejectReturning(
		sctx.Ctx,
		req.ApprovalRequestID,
		req.ReviewedBy,
		*req.ReviewComments,
	)
	if err != nil {
		log.Error(sctx.Ctx, "Error rejecting request: %v", err)
		return nil, err
	}

	log.Info(sctx.Ctx, "Successfully rejected profile update request: %s", req.ApprovalRequestID)

	return &resp.ApprovalActionResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: "REJECTION_SUCCESS",
			Message:    "Profile update rejected successfully",
		},
		ApprovalRequestID: rejectedRequest.ApprovalRequestID,
		Status:            string(rejectedRequest.Status),
		ReviewedBy:        rejectedRequest.ReviewedBy.String,
		ReviewedAt:        &rejectedRequest.ReviewedAt.Time,
		ReviewComments:    rejectedRequest.ReviewComments.String,
	}, nil
}

// GetAuditHistory retrieves audit history for an agent
// AGT-028: Get Audit History
// FR-AGT-PRF-022: Audit History Tracking
func (h *AgentProfileUpdateHandler) GetAuditHistory(sctx *serverRoute.Context, req GetAuditHistoryRequest) (*resp.AuditHistoryResponse, error) {
	log.Info(sctx.Ctx, "Fetching audit history for agent: %s, page: %d", req.AgentID, req.Page)

	// Default pagination
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 50
	}

	// Fetch audit history - single database round trip with CTE
	auditLogs, totalCount, err := h.auditRepo.GetAuditHistoryWithCount(
		sctx.Ctx,
		req.AgentID,
		req.Page,
		req.Limit,
	)
	if err != nil {
		log.Error(sctx.Ctx, "Error fetching audit history: %v", err)
		return nil, err
	}

	// Build response
	auditEntries := make([]resp.AuditLogEntry, 0, len(auditLogs))
	for _, log := range auditLogs {
		auditEntries = append(auditEntries, resp.AuditLogEntry{
			AuditID:      log.AuditID,
			ActionType:   string(log.ActionType),
			ActionReason: log.ActionReason.String,
			FieldName:    log.FieldName.String,
			OldValue:     log.OldValue.String,
			NewValue:     log.NewValue.String,
			PerformedBy:  log.PerformedBy,
			PerformedAt:  log.PerformedAt,
			IPAddress:    log.IPAddress.String,
			UserAgent:    log.UserAgent.String,
		})
	}

	log.Info(sctx.Ctx, "Found %d audit log entries (total: %d)", len(auditEntries), totalCount)

	return &resp.AuditHistoryResponse{
		StatusCodeAndMessage: port.ListSuccess,
		MetaDataResponse:     port.NewMetaDataResponse((req.Page-1)*req.Limit, req.Limit, uint64(totalCount)),
		Data:                 auditEntries,
	}, nil
}

// Helper functions

func getEditableFieldsForSection(section string) []string {
	editableFields := map[string][]string{
		"personal_info": {"title", "first_name", "middle_name", "last_name", "gender", "date_of_birth", "marital_status", "aadhar_number", "pan_number"},
		"address":       {"address_line1", "address_line2", "address_line3", "city", "district", "state", "country", "pincode"},
		"contact":       {"contact_number", "is_primary"},
		"bank_details":  {"account_number", "ifsc_code", "bank_name", "branch_name"},
		"status":        {"status", "status_reason"},
	}
	return editableFields[section]
}

func getCriticalFieldsForSection(section string) []string {
	criticalFields := map[string][]string{
		"personal_info": {"first_name", "last_name", "pan_number"},
		"bank_details":  {"account_number", "ifsc_code"},
		"status":        {"status"},
	}
	return criticalFields[section]
}

func getUpdatedFieldsList(changes map[string]interface{}) []string {
	fields := make([]string, 0, len(changes))
	for field := range changes {
		fields = append(fields, field)
	}
	return fields
}
