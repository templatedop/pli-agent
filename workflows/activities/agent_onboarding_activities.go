package activities

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.temporal.io/sdk/activity"

	"pli-agent-api/core/domain"
	repo "pli-agent-api/repo/postgres"
)

// AgentOnboardingActivities contains all activities for Agent Onboarding Workflow (WF-002)
// Activities: ACT-011 to ACT-028 (18 activities)
type AgentOnboardingActivities struct {
	profileRepo *repo.AgentProfileRepository
	addressRepo *repo.AgentAddressRepository
	contactRepo *repo.AgentContactRepository
	emailRepo   *repo.AgentEmailRepository
	licenseRepo *repo.AgentLicenseRepository
	auditRepo   *repo.AgentAuditLogRepository
	sessionRepo *repo.AgentProfileSessionRepository // For workflow self-recording
}

// NewAgentOnboardingActivities creates a new AgentOnboardingActivities instance
func NewAgentOnboardingActivities(
	profileRepo *repo.AgentProfileRepository,
	addressRepo *repo.AgentAddressRepository,
	contactRepo *repo.AgentContactRepository,
	emailRepo *repo.AgentEmailRepository,
	licenseRepo *repo.AgentLicenseRepository,
	auditRepo *repo.AgentAuditLogRepository,
	sessionRepo *repo.AgentProfileSessionRepository,
) *AgentOnboardingActivities {
	return &AgentOnboardingActivities{
		profileRepo: profileRepo,
		addressRepo: addressRepo,
		contactRepo: contactRepo,
		emailRepo:   emailRepo,
		licenseRepo: licenseRepo,
		auditRepo:   auditRepo,
		sessionRepo: sessionRepo,
	}
}

// ==================== Input/Output Structs ====================

// OnboardingInput is the workflow input
type OnboardingInput struct {
	AgentType            string
	ProfileData          ProfileData
	EmployeeID           string
	AdvisorCoordinatorID string
	CircleID             string
	DivisionID           string
	Documents            []Document
	InitiatedBy          string
}

// ProfileData contains profile information
type ProfileData struct {
	FirstName    string
	MiddleName   string
	LastName     string
	PANNumber    string
	DateOfBirth  time.Time
	Gender       string
	MobileNumber string
	Email        string
	AadharNumber string
	OfficeCode   string
	Addresses    []Address
}

// Address represents an address
type Address struct {
	AddressType       string
	AddressLine1      string
	AddressLine2      string
	City              string
	State             string
	Country           string
	Pincode           string
	IsSameAsPermanent bool
}

// Document represents a document
type Document struct {
	DocumentType string
	DocumentURL  string
	FileName     string
}

// OnboardingOutput is the workflow output
type OnboardingOutput struct {
	AgentID          string
	AgentCode        string
	Status           string
	Message          string
	ProfileCreatedAt time.Time
}

// ApprovalDecision represents an approval decision signal
type ApprovalDecision struct {
	Approved        bool
	RejectionReason string
	ApprovedBy      string
}

// HRMSData represents data from HRMS
type HRMSData struct {
	EmployeeID    string
	FirstName     string
	MiddleName    string
	LastName      string
	DateOfBirth   time.Time
	Gender        string
	MobileNumber  string
	Email         string
	OfficeCode    string
	Designation   string
	ServiceNumber string
}

// ==================== RecordWorkflowStartActivity ====================
// This activity is called as the FIRST activity in the workflow
// It records the workflow start in the database
// CRITICAL: This ensures the workflow is self-recording and self-healing
// If the database update fails, Temporal will retry this activity
// The workflow doesn't proceed until the database knows about it

type RecordWorkflowStartInput struct {
	SessionID     string
	WorkflowID    string
	RunID         string
	WorkflowState string
	CurrentStep   string
	NextStep      string
	Progress      int
	SubmittedBy   string
}

type RecordWorkflowStartOutput struct {
	Recorded bool
	Message  string
}

// RecordWorkflowStartActivity records workflow start in database
// This is the FIRST activity called by the workflow
// Ensures atomicity: Either database knows about workflow OR workflow fails
func (a *AgentOnboardingActivities) RecordWorkflowStartActivity(ctx context.Context, input RecordWorkflowStartInput) (*RecordWorkflowStartOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("RecordWorkflowStartActivity started",
		"SessionID", input.SessionID,
		"WorkflowID", input.WorkflowID,
		"RunID", input.RunID)

	activity.RecordHeartbeat(ctx, "Recording workflow start in database")

	// ATOMIC: Link Temporal workflow + update state in single database round trip
	// This is the CRITICAL operation that makes the workflow self-recording
	// If this fails, Temporal will retry this activity
	// The workflow doesn't proceed until this succeeds
	_, err := a.sessionRepo.LinkTemporalWorkflowAndUpdateStateReturning(
		ctx,
		input.SessionID,
		input.WorkflowID,
		input.RunID,
		input.WorkflowState,
		input.CurrentStep,
		input.NextStep,
		input.Progress,
		input.SubmittedBy,
	)
	if err != nil {
		logger.Error("Failed to record workflow start in database", "error", err)
		return nil, fmt.Errorf("failed to record workflow start: %w", err)
	}

	logger.Info("Workflow start recorded successfully in database", "SessionID", input.SessionID)

	return &RecordWorkflowStartOutput{
		Recorded: true,
		Message:  "Workflow start recorded in database",
	}, nil
}

// ==================== ACT-011: ValidateAgentTypeActivity ====================

type ValidateAgentTypeInput struct {
	AgentType string
}

type ValidateAgentTypeOutput struct {
	Valid bool
}

// ValidateAgentTypeActivity validates the agent type selection
// ACT-011: Timeout 10s, Retry 3 attempts
// FR-AGT-PRF-001: New Profile Creation
// BR-AGT-PRF-001 to BR-AGT-PRF-004: Agent Type Rules
func (a *AgentOnboardingActivities) ValidateAgentTypeActivity(ctx context.Context, input ValidateAgentTypeInput) (*ValidateAgentTypeOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("ValidateAgentTypeActivity started", "AgentType", input.AgentType)

	validTypes := map[string]bool{
		domain.AgentTypeAdvisor:              true,
		domain.AgentTypeAdvisorCoordinator:   true,
		domain.AgentTypeDepartmentalEmployee: true,
		domain.AgentTypeFieldOfficer:         true,
	}

	if !validTypes[input.AgentType] {
		return nil, fmt.Errorf("invalid agent type: %s", input.AgentType)
	}

	return &ValidateAgentTypeOutput{Valid: true}, nil
}

// ==================== ACT-012: ValidateProfileDataActivity ====================

type ValidateProfileDataInput struct {
	ProfileData ProfileData
}

type ValidateProfileDataOutput struct {
	Valid            bool
	ValidationErrors []string
}

// ValidateProfileDataActivity validates all profile fields
// ACT-012: Timeout 30s, Retry 3 attempts
// VR-AGT-PRF-001 to VR-AGT-PRF-007: Personal Information Validation
// VR-AGT-PRF-003: PAN Format Validation
func (a *AgentOnboardingActivities) ValidateProfileDataActivity(ctx context.Context, input ValidateProfileDataInput) (*ValidateProfileDataOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("ValidateProfileDataActivity started")

	errors := []string{}

	// Validate first name
	if len(input.ProfileData.FirstName) < 2 || len(input.ProfileData.FirstName) > 50 {
		errors = append(errors, "first_name must be 2-50 characters")
	}

	// Validate last name
	if len(input.ProfileData.LastName) < 2 || len(input.ProfileData.LastName) > 50 {
		errors = append(errors, "last_name must be 2-50 characters")
	}

	// Validate PAN format (VR-AGT-PRF-003)
	panRegex := regexp.MustCompile(`^[A-Z]{5}[0-9]{4}[A-Z]{1}$`)
	if !panRegex.MatchString(input.ProfileData.PANNumber) {
		errors = append(errors, "invalid PAN format (must be AAAAA9999A)")
	}

	// Validate date of birth (VR-AGT-PRF-002)
	if input.ProfileData.DateOfBirth.After(time.Now().AddDate(-18, 0, 0)) {
		errors = append(errors, "agent must be at least 18 years old")
	}

	// Validate gender (VR-AGT-PRF-005)
	validGenders := map[string]bool{
		domain.GenderMale:   true,
		domain.GenderFemale: true,
		domain.GenderOther:  true,
	}
	if !validGenders[input.ProfileData.Gender] {
		errors = append(errors, "invalid gender")
	}

	// Validate mobile number (VR-AGT-PRF-011)
	mobileRegex := regexp.MustCompile(`^[6-9][0-9]{9}$`)
	if !mobileRegex.MatchString(input.ProfileData.MobileNumber) {
		errors = append(errors, "invalid mobile number (must be 10 digits starting with 6-9)")
	}

	// Validate email (VR-AGT-PRF-012)
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(input.ProfileData.Email) {
		errors = append(errors, "invalid email format")
	}

	// Validate Aadhar if provided (VR-AGT-PRF-004)
	if input.ProfileData.AadharNumber != "" {
		aadharRegex := regexp.MustCompile(`^[0-9]{12}$`)
		if !aadharRegex.MatchString(input.ProfileData.AadharNumber) {
			errors = append(errors, "invalid Aadhar number (must be 12 digits)")
		}
	}

	if len(errors) > 0 {
		return &ValidateProfileDataOutput{
			Valid:            false,
			ValidationErrors: errors,
		}, fmt.Errorf("validation failed: %s", strings.Join(errors, ", "))
	}

	return &ValidateProfileDataOutput{Valid: true}, nil
}

// ==================== ACT-013: ValidateEmployeeIDActivity ====================

type ValidateEmployeeIDInput struct {
	EmployeeID string
}

type ValidateEmployeeIDOutput struct {
	Valid      bool
	EmployeeID string
}

// ValidateEmployeeIDActivity validates employee ID with HRMS
// ACT-013: Timeout 30s, Retry 5 attempts (external API)
// BR-AGT-PRF-003: HRMS Integration
func (a *AgentOnboardingActivities) ValidateEmployeeIDActivity(ctx context.Context, input ValidateEmployeeIDInput) (*ValidateEmployeeIDOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("ValidateEmployeeIDActivity started", "EmployeeID", input.EmployeeID)

	// TODO: Integrate with actual HRMS API
	// For now, simulate HRMS validation
	if input.EmployeeID == "" {
		return nil, fmt.Errorf("employee ID is required")
	}

	// Simulate HRMS API call
	// In production: Call HRMS service to validate employee ID
	activity.RecordHeartbeat(ctx, "Validating with HRMS")

	// Mock validation: Check format EMP-XXXXX
	if !strings.HasPrefix(input.EmployeeID, "EMP-") {
		return nil, fmt.Errorf("invalid employee ID format (must start with EMP-)")
	}

	logger.Info("Employee ID validated successfully", "EmployeeID", input.EmployeeID)
	return &ValidateEmployeeIDOutput{
		Valid:      true,
		EmployeeID: input.EmployeeID,
	}, nil
}

// ==================== ACT-014: FetchHRMSDataActivity ====================

type FetchHRMSDataInput struct {
	EmployeeID string
}

type FetchHRMSDataOutput struct {
	HRMSData HRMSData
}

// FetchHRMSDataActivity fetches employee data from HRMS
// ACT-014: Timeout 1m, Retry 5 attempts (external API)
// BR-AGT-PRF-003: HRMS Integration
func (a *AgentOnboardingActivities) FetchHRMSDataActivity(ctx context.Context, input FetchHRMSDataInput) (*FetchHRMSDataOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("FetchHRMSDataActivity started", "EmployeeID", input.EmployeeID)

	// TODO: Integrate with actual HRMS API
	// For now, return mock data
	activity.RecordHeartbeat(ctx, "Fetching from HRMS")

	// Mock HRMS data
	mockData := HRMSData{
		EmployeeID:    input.EmployeeID,
		FirstName:     "Rajesh",
		MiddleName:    "Kumar",
		LastName:      "Sharma",
		DateOfBirth:   time.Date(1985, 5, 15, 0, 0, 0, 0, time.UTC),
		Gender:        domain.GenderMale,
		MobileNumber:  "9876543210",
		Email:         "rajesh.sharma@lic.in",
		OfficeCode:    "OFF-001",
		Designation:   "Assistant",
		ServiceNumber: "SRV-12345",
	}

	logger.Info("HRMS data fetched successfully", "EmployeeID", input.EmployeeID)
	return &FetchHRMSDataOutput{HRMSData: mockData}, nil
}

// ==================== ACT-015: AutoPopulateProfileActivity ====================

type AutoPopulateProfileInput struct {
	HRMSData    HRMSData
	ProfileData ProfileData
}

type AutoPopulateProfileOutput struct {
	PopulatedProfile ProfileData
}

// AutoPopulateProfileActivity auto-populates profile from HRMS data
// ACT-015: Timeout 30s, Retry 3 attempts
// BR-AGT-PRF-003: HRMS Integration
func (a *AgentOnboardingActivities) AutoPopulateProfileActivity(ctx context.Context, input AutoPopulateProfileInput) (*AutoPopulateProfileOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("AutoPopulateProfileActivity started")

	// Merge HRMS data with existing profile data
	// HRMS data takes precedence for certain fields
	populated := input.ProfileData

	if input.HRMSData.FirstName != "" {
		populated.FirstName = input.HRMSData.FirstName
	}
	if input.HRMSData.MiddleName != "" {
		populated.MiddleName = input.HRMSData.MiddleName
	}
	if input.HRMSData.LastName != "" {
		populated.LastName = input.HRMSData.LastName
	}
	if !input.HRMSData.DateOfBirth.IsZero() {
		populated.DateOfBirth = input.HRMSData.DateOfBirth
	}
	if input.HRMSData.Gender != "" {
		populated.Gender = input.HRMSData.Gender
	}
	if input.HRMSData.MobileNumber != "" {
		populated.MobileNumber = input.HRMSData.MobileNumber
	}
	if input.HRMSData.Email != "" {
		populated.Email = input.HRMSData.Email
	}
	if input.HRMSData.OfficeCode != "" {
		populated.OfficeCode = input.HRMSData.OfficeCode
	}

	logger.Info("Profile auto-populated successfully")
	return &AutoPopulateProfileOutput{PopulatedProfile: populated}, nil
}

// ==================== ACT-016: ValidateAdvisorCoordinatorActivity ====================

type ValidateAdvisorCoordinatorInput struct {
	CoordinatorID string
}

type ValidateAdvisorCoordinatorOutput struct {
	Valid         bool
	CoordinatorID string
}

// ValidateAdvisorCoordinatorActivity validates advisor coordinator linkage
// ACT-016: Timeout 20s, Retry 3 attempts
// BR-AGT-PRF-001: Advisor Coordinator Linkage Requirement
func (a *AgentOnboardingActivities) ValidateAdvisorCoordinatorActivity(ctx context.Context, input ValidateAdvisorCoordinatorInput) (*ValidateAdvisorCoordinatorOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("ValidateAdvisorCoordinatorActivity started", "CoordinatorID", input.CoordinatorID)

	if input.CoordinatorID == "" {
		return nil, fmt.Errorf("advisor coordinator ID is required for ADVISOR agent type")
	}

	// Check if coordinator exists and is active
	coordinator, err := a.profileRepo.FindByID(ctx, input.CoordinatorID)
	if err != nil {
		return nil, fmt.Errorf("coordinator not found: %w", err)
	}

	if coordinator.AgentType != domain.AgentTypeAdvisorCoordinator {
		return nil, fmt.Errorf("specified agent is not an advisor coordinator")
	}

	if !coordinator.IsActive() {
		return nil, fmt.Errorf("advisor coordinator is not active")
	}

	logger.Info("Advisor coordinator validated successfully", "CoordinatorID", input.CoordinatorID)
	return &ValidateAdvisorCoordinatorOutput{
		Valid:         true,
		CoordinatorID: input.CoordinatorID,
	}, nil
}

// ==================== ACT-017: ValidatePANUniquenessActivity ====================

type ValidatePANUniquenessInput struct {
	PANNumber string
}

type ValidatePANUniquenessOutput struct {
	IsUnique  bool
	PANNumber string
}

// ValidatePANUniquenessActivity checks PAN uniqueness
// ACT-017: Timeout 20s, Retry 3 attempts
// BR-AGT-PRF-006: PAN Update with Format and Uniqueness Validation
func (a *AgentOnboardingActivities) ValidatePANUniquenessActivity(ctx context.Context, input ValidatePANUniquenessInput) (*ValidatePANUniquenessOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("ValidatePANUniquenessActivity started", "PANNumber", input.PANNumber)

	// Check PAN uniqueness (exclude empty string for new profile)
	isUnique, err := a.profileRepo.ValidatePANUniqueness(ctx, input.PANNumber, "")
	if err != nil {
		return nil, fmt.Errorf("failed to validate PAN uniqueness: %w", err)
	}

	logger.Info("PAN uniqueness validated", "PANNumber", input.PANNumber, "IsUnique", isUnique)
	return &ValidatePANUniquenessOutput{
		IsUnique:  isUnique,
		PANNumber: input.PANNumber,
	}, nil
}

// ==================== ACT-018: ValidateMandatoryFieldsActivity ====================

type ValidateMandatoryFieldsInput struct {
	AgentType   string
	ProfileData ProfileData
}

type ValidateMandatoryFieldsOutput struct {
	Valid         bool
	MissingFields []string
}

// ValidateMandatoryFieldsActivity validates all mandatory fields
// ACT-018: Timeout 30s, Retry 3 attempts
// BR-AGT-PRF-007: Personal Information Update Rules
func (a *AgentOnboardingActivities) ValidateMandatoryFieldsActivity(ctx context.Context, input ValidateMandatoryFieldsInput) (*ValidateMandatoryFieldsOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("ValidateMandatoryFieldsActivity started", "AgentType", input.AgentType)

	missing := []string{}

	// Common mandatory fields
	if input.ProfileData.FirstName == "" {
		missing = append(missing, "first_name")
	}
	if input.ProfileData.LastName == "" {
		missing = append(missing, "last_name")
	}
	if input.ProfileData.PANNumber == "" {
		missing = append(missing, "pan_number")
	}
	if input.ProfileData.DateOfBirth.IsZero() {
		missing = append(missing, "date_of_birth")
	}
	if input.ProfileData.Gender == "" {
		missing = append(missing, "gender")
	}
	if input.ProfileData.MobileNumber == "" {
		missing = append(missing, "mobile_number")
	}
	if input.ProfileData.Email == "" {
		missing = append(missing, "email")
	}
	if input.ProfileData.OfficeCode == "" {
		missing = append(missing, "office_code")
	}

	// Validate addresses
	if len(input.ProfileData.Addresses) == 0 {
		missing = append(missing, "addresses")
	} else {
		// Check for at least one permanent address
		hasPermanent := false
		for _, addr := range input.ProfileData.Addresses {
			if addr.AddressType == domain.AddressTypePermanent {
				hasPermanent = true
				break
			}
		}
		if !hasPermanent {
			missing = append(missing, "permanent_address")
		}
	}

	if len(missing) > 0 {
		return &ValidateMandatoryFieldsOutput{
			Valid:         false,
			MissingFields: missing,
		}, fmt.Errorf("missing mandatory fields: %s", strings.Join(missing, ", "))
	}

	logger.Info("Mandatory fields validated successfully")
	return &ValidateMandatoryFieldsOutput{Valid: true}, nil
}

// ==================== ACT-019: UploadKYCDocumentsActivity ====================

type UploadKYCDocumentsInput struct {
	AgentType string
	Documents []Document
}

type UploadKYCDocumentsOutput struct {
	DocumentURLs []string
}

// UploadKYCDocumentsActivity uploads KYC documents
// ACT-019: Timeout 2m, Retry 3 attempts
// BR-AGT-PRF-027: External Identification Number Assignment
func (a *AgentOnboardingActivities) UploadKYCDocumentsActivity(ctx context.Context, input UploadKYCDocumentsInput) (*UploadKYCDocumentsOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("UploadKYCDocumentsActivity started", "DocumentCount", len(input.Documents))

	// TODO: Integrate with actual document storage service (S3, MinIO, etc.)
	// For now, simulate document upload
	activity.RecordHeartbeat(ctx, "Uploading documents")

	documentURLs := []string{}
	for i, doc := range input.Documents {
		// Simulate upload
		activity.RecordHeartbeat(ctx, fmt.Sprintf("Uploading document %d/%d", i+1, len(input.Documents)))

		// Mock URL
		url := fmt.Sprintf("https://storage.lic.in/kyc/%s/%s", input.AgentType, doc.FileName)
		documentURLs = append(documentURLs, url)
	}

	logger.Info("Documents uploaded successfully", "Count", len(documentURLs))
	return &UploadKYCDocumentsOutput{DocumentURLs: documentURLs}, nil
}

// ==================== ACT-020: ValidateDocumentsActivity ====================

type ValidateDocumentsInput struct {
	DocumentURLs []string
}

type ValidateDocumentsOutput struct {
	Valid bool
}

// ValidateDocumentsActivity validates uploaded documents
// ACT-020: Timeout 1m, Retry 3 attempts
func (a *AgentOnboardingActivities) ValidateDocumentsActivity(ctx context.Context, input ValidateDocumentsInput) (*ValidateDocumentsOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("ValidateDocumentsActivity started", "DocumentCount", len(input.DocumentURLs))

	// TODO: Integrate with document validation service
	// For now, simulate validation
	activity.RecordHeartbeat(ctx, "Validating documents")

	if len(input.DocumentURLs) == 0 {
		return nil, fmt.Errorf("no documents to validate")
	}

	// Mock validation: Check if URLs are valid
	for _, url := range input.DocumentURLs {
		if url == "" {
			return nil, fmt.Errorf("invalid document URL")
		}
	}

	logger.Info("Documents validated successfully")
	return &ValidateDocumentsOutput{Valid: true}, nil
}

// ==================== ACT-021: CheckApprovalRequiredActivity ====================

type CheckApprovalRequiredInput struct {
	AgentType string
}

type CheckApprovalRequiredOutput struct {
	ApprovalRequired bool
	ApprovalType     string
}

// CheckApprovalRequiredActivity checks if approval is required
// ACT-021: Timeout 10s, Retry 3 attempts
func (a *AgentOnboardingActivities) CheckApprovalRequiredActivity(ctx context.Context, input CheckApprovalRequiredInput) (*CheckApprovalRequiredOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("CheckApprovalRequiredActivity started", "AgentType", input.AgentType)

	// Business rule: Approval required for certain agent types
	// For ADVISOR and ADVISOR_COORDINATOR: Require supervisor approval
	approvalRequired := false
	approvalType := ""

	switch input.AgentType {
	case domain.AgentTypeAdvisor:
		approvalRequired = true
		approvalType = "SUPERVISOR_APPROVAL"
	case domain.AgentTypeAdvisorCoordinator:
		approvalRequired = true
		approvalType = "MANAGEMENT_APPROVAL"
	}

	logger.Info("Approval check completed", "ApprovalRequired", approvalRequired, "ApprovalType", approvalType)
	return &CheckApprovalRequiredOutput{
		ApprovalRequired: approvalRequired,
		ApprovalType:     approvalType,
	}, nil
}

// ==================== ACT-022: SendApprovalRequestActivity ====================

type SendApprovalRequestInput struct {
	AgentType   string
	ProfileData ProfileData
	InitiatedBy string
}

type SendApprovalRequestOutput struct {
	RequestID string
	SentTo    string
}

// SendApprovalRequestActivity sends approval request
// ACT-022: Timeout 30s, Retry 5 attempts
func (a *AgentOnboardingActivities) SendApprovalRequestActivity(ctx context.Context, input SendApprovalRequestInput) (*SendApprovalRequestOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("SendApprovalRequestActivity started", "AgentType", input.AgentType)

	// NOTE: This activity prepares approval request data
	// The actual approval workflow is started as a child workflow from main workflow
	// This ensures proper human-in-the-loop pattern with signal handling

	activity.RecordHeartbeat(ctx, "Preparing approval request")

	// Generate approval request ID
	requestID := fmt.Sprintf("APR-%d", time.Now().Unix())

	// Determine approvers based on agent type
	// Business Rule: Different approval authorities for different agent types
	var approvers []string
	switch input.AgentType {
	case domain.AgentTypeAdvisor:
		// ADVISOR: Requires supervisor approval
		approvers = []string{"supervisor@indiapost.gov.in"}
	case domain.AgentTypeAdvisorCoordinator:
		// ADVISOR_COORDINATOR: Requires management approval
		approvers = []string{"manager@indiapost.gov.in", "regional-head@indiapost.gov.in"}
	default:
		approvers = []string{"admin@indiapost.gov.in"}
	}

	logger.Info("Approval request prepared",
		"RequestID", requestID,
		"Approvers", approvers,
		"AgentType", input.AgentType)

	return &SendApprovalRequestOutput{
		RequestID: requestID,
		SentTo:    approvers[0], // Primary approver
	}, nil
}

// ==================== SendApprovalNotificationActivity ====================

type SendApprovalNotificationInput struct {
	RequestID    string
	AgentType    string
	FirstName    string
	LastName     string
	PANNumber    string
	ApprovalType string
	Approvers    []string
}

// SendApprovalNotificationActivity sends approval notification to approvers
// INT-AGT-005: Notification Service Integration
// Sends email with approval link
func (a *AgentOnboardingActivities) SendApprovalNotificationActivity(ctx context.Context, input SendApprovalNotificationInput) (bool, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("SendApprovalNotificationActivity started", "RequestID", input.RequestID)

	activity.RecordHeartbeat(ctx, "Sending approval notifications")

	// INT-AGT-005: Notification Service Integration
	// In production: Call notification service to send emails
	// For now, log the notification
	for _, approver := range input.Approvers {
		logger.Info("Sending approval notification",
			"To", approver,
			"RequestID", input.RequestID,
			"AgentName", fmt.Sprintf("%s %s", input.FirstName, input.LastName),
			"ApprovalType", input.ApprovalType)

		// Mock notification content
		// Subject: Agent Profile Approval Required - [FirstName LastName]
		// Body:
		//   New agent profile requires your approval
		//   Name: [FirstName LastName]
		//   Agent Type: [AgentType]
		//   PAN: [PANNumber]
		//   Approval Type: [ApprovalType]
		//
		//   Approve: [LINK]
		//   Reject: [LINK]
	}

	logger.Info("Approval notifications sent successfully", "Count", len(input.Approvers))
	return true, nil
}

// ==================== ACT-023: GenerateAgentCodeActivity ====================

type GenerateAgentCodeInput struct {
	AgentType string
}

type GenerateAgentCodeOutput struct {
	AgentCode string
}

// GenerateAgentCodeActivity generates unique agent code
// ACT-023: Timeout 20s, Retry 3 attempts
// BR-AGT-PRF-027: External Identification Number Assignment
func (a *AgentOnboardingActivities) GenerateAgentCodeActivity(ctx context.Context, input GenerateAgentCodeInput) (*GenerateAgentCodeOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("GenerateAgentCodeActivity started", "AgentType", input.AgentType)

	// Generate agent code: AGT-{YEAR}-{SEQUENCE}
	// TODO: Implement proper sequence generation from database
	year := time.Now().Year()
	sequence := time.Now().Unix() % 1000000 // Mock sequence

	agentCode := fmt.Sprintf("AGT-%d-%06d", year, sequence)

	logger.Info("Agent code generated", "AgentCode", agentCode)
	return &GenerateAgentCodeOutput{AgentCode: agentCode}, nil
}

// ==================== ACT-024: CreateAgentProfileActivity ====================

type CreateAgentProfileInput struct {
	AgentType            string
	AgentCode            string
	EmployeeID           string
	ProfileData          ProfileData
	AdvisorCoordinatorID string
	CircleID             string
	DivisionID           string
	CreatedBy            string
}

type CreateAgentProfileOutput struct {
	AgentID string
}

// CreateAgentProfileActivity creates agent profile in database
// ACT-024: Timeout 1m, Retry 3 attempts
// FR-AGT-PRF-001: New Profile Creation
// BR-AGT-PRF-001: Advisor Coordinator Linkage Requirement
// ATOMIC: Creates profile + addresses + contacts + emails in single database transaction
func (a *AgentOnboardingActivities) CreateAgentProfileActivity(ctx context.Context, input CreateAgentProfileInput) (*CreateAgentProfileOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("CreateAgentProfileActivity started", "AgentType", input.AgentType)

	activity.RecordHeartbeat(ctx, "Creating agent profile with all related entities")

	// Build agent profile
	// FR-AGT-PRF-001: New Profile Creation
	profile := domain.AgentProfile{
		AgentType:            input.AgentType,
		AgentCode:            sql.NullString{String: input.AgentCode, Valid: true},
		EmployeeID:           sql.NullString{String: input.EmployeeID, Valid: input.EmployeeID != ""},
		OfficeCode:           input.ProfileData.OfficeCode,
		CircleID:             sql.NullString{String: input.CircleID, Valid: input.CircleID != ""},
		DivisionID:           sql.NullString{String: input.DivisionID, Valid: input.DivisionID != ""},
		AdvisorCoordinatorID: sql.NullString{String: input.AdvisorCoordinatorID, Valid: input.AdvisorCoordinatorID != ""},
		FirstName:            input.ProfileData.FirstName,
		MiddleName:           sql.NullString{String: input.ProfileData.MiddleName, Valid: input.ProfileData.MiddleName != ""},
		LastName:             input.ProfileData.LastName,
		Gender:               input.ProfileData.Gender,
		DateOfBirth:          input.ProfileData.DateOfBirth,
		AadharNumber:         sql.NullString{String: input.ProfileData.AadharNumber, Valid: input.ProfileData.AadharNumber != ""},
		PANNumber:            input.ProfileData.PANNumber,
		Status:               domain.AgentStatusActive,
		StatusDate:           time.Now(),
		CreatedBy:            input.CreatedBy,
	}

	// Build addresses array for bulk insert (uses UNNEST)
	var addresses []domain.AgentAddress
	for _, addr := range input.ProfileData.Addresses {
		addresses = append(addresses, domain.AgentAddress{
			AddressType: addr.AddressType,
			Line1:       addr.AddressLine1,
			Line2:       sql.NullString{String: addr.AddressLine2, Valid: addr.AddressLine2 != ""},
			Line3:       sql.NullString{},
			City:        addr.City,
			District:    sql.NullString{},
			State:       addr.State,
			Country:     addr.Country,
			Pincode:     addr.Pincode,
			IsPrimary:   addr.AddressType == "PERMANENT",
			ValidFrom:   time.Now(),
			CreatedBy:   input.CreatedBy,
		})
	}

	// Build contacts array for bulk insert (uses UNNEST)
	var contacts []domain.AgentContact
	if input.ProfileData.MobileNumber != "" {
		contacts = append(contacts, domain.AgentContact{
			ContactType:   domain.ContactTypeMobile,
			ContactNumber: input.ProfileData.MobileNumber,
			IsPrimary:     true,
			IsVerified:    false,
			CreatedBy:     input.CreatedBy,
		})
	}

	// Build emails array for bulk insert (uses UNNEST)
	var emails []domain.AgentEmail
	if input.ProfileData.Email != "" {
		emails = append(emails, domain.AgentEmail{
			EmailAddress: input.ProfileData.Email,
			IsPrimary:    true,
			IsVerified:   false,
			CreatedBy:    input.CreatedBy,
		})
	}

	// ATOMIC: Create profile with all related entities in single database transaction
	// CRITICAL: Ensures atomicity - either all entities are created, or none
	// Prevents inconsistent state where:
	// - Profile exists but addresses/contacts/emails don't
	// - Profile create succeeds but address insert fails leaving partial data
	// Uses CTE pattern with UNNEST for efficient bulk inserts
	createInput := repo.CreateWithRelatedEntitiesInput{
		Profile:   profile,
		Addresses: addresses,
		Contacts:  contacts,
		Emails:    emails,
	}

	created, err := a.profileRepo.CreateWithRelatedEntities(ctx, createInput)
	if err != nil {
		logger.Error("Failed to create agent profile with related entities", "error", err)
		return nil, fmt.Errorf("failed to create agent profile: %w", err)
	}

	agentID := created.AgentID
	logger.Info("Agent profile created successfully with all related entities",
		"AgentID", agentID,
		"Addresses", len(addresses),
		"Contacts", len(contacts),
		"Emails", len(emails))

	return &CreateAgentProfileOutput{AgentID: agentID}, nil
}

// ==================== ACT-025: LinkToHierarchyActivity ====================

type LinkToHierarchyInput struct {
	AgentID              string
	AgentType            string
	AdvisorCoordinatorID string
	CircleID             string
	DivisionID           string
}

type LinkToHierarchyOutput struct {
	Linked bool
}

// LinkToHierarchyActivity links agent to organizational hierarchy
// ACT-025: Timeout 30s, Retry 3 attempts
// BR-AGT-PRF-001: Advisor Coordinator Linkage Requirement
// BR-AGT-PRF-002: Advisor Coordinator Geographic Assignment
func (a *AgentOnboardingActivities) LinkToHierarchyActivity(ctx context.Context, input LinkToHierarchyInput) (*LinkToHierarchyOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("LinkToHierarchyActivity started", "AgentID", input.AgentID, "AgentType", input.AgentType)

	// TODO: Integrate with hierarchy management service
	// For now, the linkage is already done during profile creation via foreign keys
	// In production: Update hierarchy tables, send events, etc.

	activity.RecordHeartbeat(ctx, "Linking to hierarchy")

	logger.Info("Agent linked to hierarchy successfully", "AgentID", input.AgentID)
	return &LinkToHierarchyOutput{Linked: true}, nil
}

// ==================== ACT-026: CreateLicenseRecordActivity ====================

type CreateLicenseRecordInput struct {
	AgentID     string
	LicenseType string
	CreatedBy   string
}

type CreateLicenseRecordOutput struct {
	LicenseID string
}

// CreateLicenseRecordActivity creates initial license record
// ACT-026: Timeout 30s, Retry 3 attempts
// BR-AGT-PRF-012: License Renewal Period Rules
func (a *AgentOnboardingActivities) CreateLicenseRecordActivity(ctx context.Context, input CreateLicenseRecordInput) (*CreateLicenseRecordOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("CreateLicenseRecordActivity started", "AgentID", input.AgentID)

	activity.RecordHeartbeat(ctx, "Creating license record")

	// Create provisional license (1 year validity)
	// BR-AGT-PRF-012: Provisional license valid for 1 year
	licenseDate := time.Now()
	renewalDate := licenseDate.AddDate(1, 0, 0) // 1 year from now

	license := domain.AgentLicense{
		AgentID:              input.AgentID,
		LicenseLine:          "LIFE",
		LicenseType:          input.LicenseType,
		LicenseNumber:        fmt.Sprintf("LIC-%s-%d", input.AgentID[:8], time.Now().Unix()%10000),
		ResidentStatus:       sql.NullString{String: "RESIDENT", Valid: true},
		LicenseDate:          licenseDate,
		RenewalDate:          renewalDate,
		AuthorityDate:        sql.NullTime{Time: licenseDate, Valid: true},
		RenewalCount:         0,
		LicenseStatus:        domain.LicenseStatusActive,
		LicentiateExamPassed: false,
		IsPrimary:            true,
		CreatedBy:            input.CreatedBy,
	}

	created, err := a.licenseRepo.Create(ctx, license)
	if err != nil {
		logger.Error("Failed to create license record", "error", err)
		return nil, fmt.Errorf("failed to create license record: %w", err)
	}

	logger.Info("License record created successfully", "LicenseID", created.LicenseID)
	return &CreateLicenseRecordOutput{LicenseID: created.LicenseID}, nil
}

// ==================== ACT-027: SendWelcomeEmailActivity ====================

type SendWelcomeEmailInput struct {
	AgentID   string
	AgentCode string
	Email     string
	FirstName string
	LastName  string
}

// SendWelcomeEmailActivity sends welcome email to agent
// ACT-027: Timeout 1m, Retry 5 attempts
func (a *AgentOnboardingActivities) SendWelcomeEmailActivity(ctx context.Context, input SendWelcomeEmailInput) error {
	logger := activity.GetLogger(ctx)
	logger.Info("SendWelcomeEmailActivity started", "AgentID", input.AgentID, "Email", input.Email)

	// TODO: Integrate with email service
	// For now, simulate sending email
	activity.RecordHeartbeat(ctx, "Sending welcome email")

	// Mock email content
	subject := "Welcome to LIC Agent Network"
	body := fmt.Sprintf(`
		Dear %s %s,

		Welcome to the Life Insurance Corporation of India!

		Your agent profile has been successfully created.
		Agent Code: %s

		Please login to the agent portal to complete your onboarding.

		Best regards,
		LIC India
	`, input.FirstName, input.LastName, input.AgentCode)

	logger.Info("Welcome email sent successfully", "AgentID", input.AgentID, "Subject", subject, "Body", body)
	return nil
}

// ==================== ACT-028: SendWelcomeSMSActivity ====================

type SendWelcomeSMSInput struct {
	AgentID      string
	AgentCode    string
	MobileNumber string
	FirstName    string
}

// SendWelcomeSMSActivity sends welcome SMS to agent
// ACT-028: Timeout 30s, Retry 5 attempts
func (a *AgentOnboardingActivities) SendWelcomeSMSActivity(ctx context.Context, input SendWelcomeSMSInput) error {
	logger := activity.GetLogger(ctx)
	logger.Info("SendWelcomeSMSActivity started", "AgentID", input.AgentID, "Mobile", input.MobileNumber)

	// TODO: Integrate with SMS service
	// For now, simulate sending SMS
	activity.RecordHeartbeat(ctx, "Sending welcome SMS")

	// Mock SMS content
	message := fmt.Sprintf("Dear %s, Welcome to LIC! Your Agent Code: %s. Login to agent portal for details.", input.FirstName, input.AgentCode)

	logger.Info("Welcome SMS sent successfully", "AgentID", input.AgentID, "Message", message)
	return nil
}
