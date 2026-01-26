package handler

import (
	"pli-agent-api/core/port"
	resp "pli-agent-api/handler/response"
	repo "pli-agent-api/repo/postgres"

	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"
)

// AgentValidationHandler handles all validation APIs
// AGT-012 to AGT-015: Validation APIs
type AgentValidationHandler struct {
	*serverHandler.Base
	profileRepo *repo.AgentProfileRepository
}

// NewAgentValidationHandler creates a new validation handler
func NewAgentValidationHandler(profileRepo *repo.AgentProfileRepository) *AgentValidationHandler {
	base := serverHandler.New("Agent Validation APIs").SetPrefix("/v1").AddPrefix("")
	return &AgentValidationHandler{
		Base:        base,
		profileRepo: profileRepo,
	}
}

// Routes defines all validation API routes
func (h *AgentValidationHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		serverRoute.POST("/validations/pan/check-uniqueness", h.CheckPANUniqueness).Name("Check PAN Uniqueness"),
		serverRoute.POST("/validations/hrms/employee-id", h.ValidateEmployeeID).Name("Validate Employee ID"),
		serverRoute.POST("/validations/bank/ifsc", h.ValidateIFSC).Name("Validate IFSC Code"),
		serverRoute.GET("/validations/office/:office_code", h.ValidateOfficeCode).Name("Validate Office Code"),
	}
}

// CheckPANUniqueness validates PAN uniqueness
// AGT-012: Check PAN Uniqueness
// BR-AGT-PRF-006: PAN Update with Format and Uniqueness Validation
// VR-AGT-PRF-002: PAN Uniqueness
func (h *AgentValidationHandler) CheckPANUniqueness(sctx *serverRoute.Context, req CheckPANUniquenessRequest) (*resp.PANUniquenessResponse, error) {
	log.Info(sctx.Ctx, "Checking PAN uniqueness for PAN: %s", req.PANNumber)

	// Check PAN uniqueness using repository
	isUnique, err := h.profileRepo.ValidatePANUniqueness(sctx.Ctx, req.PANNumber, req.ExcludeAgentID)
	if err != nil {
		log.Error(sctx.Ctx, "Error checking PAN uniqueness: %v", err)
		return nil, err
	}

	if isUnique {
		return &resp.PANUniquenessResponse{
			StatusCodeAndMessage: port.ValidationSuccess,
			IsUnique:             true,
			Message:              "PAN is unique and available for registration",
			ExistingAgent:        nil,
		}, nil
	}

	// If not unique, fetch existing agent details
	existingProfile, err := h.profileRepo.FindByPAN(sctx.Ctx, req.PANNumber)
	if err != nil {
		log.Error(sctx.Ctx, "Error fetching existing agent by PAN: %v", err)
		return nil, err
	}

	// Build existing agent info
	existingAgent := &resp.ExistingAgentInfo{
		AgentID:   existingProfile.AgentID,
		AgentCode: existingProfile.AgentCode.String,
		FirstName: existingProfile.FirstName,
		LastName:  existingProfile.LastName,
		Status:    existingProfile.Status,
	}

	return &resp.PANUniquenessResponse{
		StatusCodeAndMessage: port.ValidationSuccess,
		IsUnique:             false,
		Message:              "PAN is already registered with another agent",
		ExistingAgent:        existingAgent,
	}, nil
}

// ValidateEmployeeID validates employee ID against HRMS
// AGT-013: Validate Employee ID (HRMS)
// BR-AGT-PRF-003: HRMS Integration Mandatory for Departmental Employees
// VR-AGT-PRF-023: HRMS Employee ID Validation
// INT-AGT-001: HRMS Integration
func (h *AgentValidationHandler) ValidateEmployeeID(sctx *serverRoute.Context, req ValidateEmployeeIDRequest) (*resp.EmployeeIDValidationResponse, error) {
	log.Info(sctx.Ctx, "Validating employee ID: %s", req.EmployeeID)

	// TODO: INT-AGT-001 - Integrate with HRMS system
	// For now, return mock response for demonstration
	// In production, this should call external HRMS service

	// Mock validation logic
	isValid := len(req.EmployeeID) >= 6 // Simple mock validation

	if !isValid {
		return &resp.EmployeeIDValidationResponse{
			StatusCodeAndMessage: port.ValidationSuccess,
			IsValid:              false,
			EmployeeStatus:       "INVALID",
			EmployeeData:         nil,
		}, nil
	}

	// Mock employee data
	employeeData := &resp.EmployeeData{
		EmployeeID:   req.EmployeeID,
		FirstName:    "Rajesh",
		LastName:     "Kumar",
		DateOfBirth:  "1985-05-15",
		Gender:       "Male",
		Designation:  "Postmaster",
		OfficeCode:   "OFF-001",
		MobileNumber: "9876543210",
		Email:        "rajesh.kumar@indiapost.gov.in",
		Status:       "ACTIVE",
	}

	return &resp.EmployeeIDValidationResponse{
		StatusCodeAndMessage: port.ValidationSuccess,
		IsValid:              true,
		EmployeeStatus:       "ACTIVE",
		EmployeeData:         employeeData,
	}, nil
}

// ValidateIFSC validates IFSC code
// AGT-014: Validate IFSC Code
// BR-AGT-PRF-018: Bank Account Details for Commission Disbursement
// VR-AGT-PRF-017: IFSC Code Format Validation
func (h *AgentValidationHandler) ValidateIFSC(sctx *serverRoute.Context, req ValidateIFSCRequest) (*resp.IFSCValidationResponse, error) {
	log.Info(sctx.Ctx, "Validating IFSC code: %s", req.IFSCCode)

	// TODO: Integrate with bank master data service or external IFSC API
	// For now, return mock response

	// Mock validation logic - check format
	if len(req.IFSCCode) != 11 {
		return &resp.IFSCValidationResponse{
			StatusCodeAndMessage: port.ValidationSuccess,
			IsValid:              false,
			BankDetails:          nil,
		}, nil
	}

	// Mock bank details
	bankDetails := &resp.BankDetails{
		IFSCCode:   req.IFSCCode,
		BankName:   "State Bank of India",
		BranchName: "Connaught Place Branch",
		City:       "New Delhi",
		State:      "Delhi",
	}

	return &resp.IFSCValidationResponse{
		StatusCodeAndMessage: port.ValidationSuccess,
		IsValid:              true,
		BankDetails:          bankDetails,
	}, nil
}

// ValidateOfficeCode validates office code
// AGT-015: Validate Office Code
// BR-AGT-PRF-034: Office Association
// VR-AGT-PRF-027: Valid Office Code
func (h *AgentValidationHandler) ValidateOfficeCode(sctx *serverRoute.Context, req OfficeCodeUri) (*resp.OfficeValidationResponse, error) {
	log.Info(sctx.Ctx, "Validating office code: %s", req.OfficeCode)

	// TODO: Query office master table
	// For now, return mock response

	// Mock validation logic
	isValid := len(req.OfficeCode) > 0

	if !isValid {
		return &resp.OfficeValidationResponse{
			StatusCodeAndMessage: port.ValidationSuccess,
			IsValid:              false,
			OfficeDetails:        nil,
		}, nil
	}

	// Mock office details
	officeDetails := &resp.OfficeDetails{
		OfficeCode: req.OfficeCode,
		OfficeName: "Central Post Office",
		OfficeType: "HEAD_OFFICE",
		CircleID:   "CIR-001",
		DivisionID: "DIV-001",
		State:      "Delhi",
		City:       "New Delhi",
	}

	return &resp.OfficeValidationResponse{
		StatusCodeAndMessage: port.ValidationSuccess,
		IsValid:              true,
		OfficeDetails:        officeDetails,
	}, nil
}
