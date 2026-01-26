package workflows

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"pli-agent-api/core/domain"
	"pli-agent-api/workflows/activities"
)

// AgentOnboardingWorkflow implements WF-002: Agent Onboarding Workflow
// Duration: 1-7 days (depending on approvals)
// Business Rules: BR-AGT-PRF-001 to BR-AGT-PRF-004, BR-AGT-PRF-027
// Complexity: MEDIUM
//
// This workflow handles the complete agent onboarding process:
// - DEPARTMENTAL_EMPLOYEE: HRMS integration with auto-population
// - FIELD_OFFICER: HRMS auto-fetch or manual entry
// - ADVISOR: Coordinator linkage mandatory
// - ADVISOR_COORDINATOR: Geographic assignment
// - Validation: PAN uniqueness, mandatory fields, document upload
// - Approval: Supervisor approval for certain agent types
func AgentOnboardingWorkflow(ctx workflow.Context, input activities.OnboardingInput) (*activities.OnboardingOutput, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("AgentOnboardingWorkflow started",
		"AgentType", input.AgentType,
		"PANNumber", input.ProfileData.PANNumber,
		"InitiatedBy", input.InitiatedBy)

	// Setup activity options with standard retry policy
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Activity instance for executing activities
	var a *activities.AgentOnboardingActivities

	// Step 0: Record Workflow Start in Database (FIRST ACTIVITY - CRITICAL)
	// This makes the workflow self-recording and self-healing
	// If this fails, Temporal will retry it
	// The workflow doesn't proceed until the database knows about it
	logger.Info("Step 0: Recording workflow start in database")
	workflowInfo := workflow.GetInfo(ctx)
	var recordStartResult activities.RecordWorkflowStartOutput
	err := workflow.ExecuteActivity(ctx, a.RecordWorkflowStartActivity, activities.RecordWorkflowStartInput{
		SessionID:     input.SessionID,
		WorkflowID:    workflowInfo.WorkflowExecution.ID,
		RunID:         workflowInfo.WorkflowExecution.RunID,
		WorkflowState: domain.WorkflowStateProfileSubmitting,
		CurrentStep:   "PROFILE_SUBMITTED",
		NextStep:      "VALIDATION",
		Progress:      10,
		SubmittedBy:   input.SubmittedBy,
	}).Get(ctx, &recordStartResult)
	if err != nil {
		logger.Error("Failed to record workflow start in database", "error", err)
		// This is CRITICAL - if we can't record the workflow start, we should fail
		// Temporal will retry this activity automatically
		return nil, fmt.Errorf("failed to record workflow start: %w", err)
	}
	logger.Info("Workflow start recorded in database successfully")

	// Step 1: Validate Agent Type and Profile Data
	logger.Info("Step 1: Validating agent type and profile data")
	var validateTypeResult activities.ValidateAgentTypeOutput
	err := workflow.ExecuteActivity(ctx, a.ValidateAgentTypeActivity, activities.ValidateAgentTypeInput{
		AgentType: input.AgentType,
	}).Get(ctx, &validateTypeResult)
	if err != nil {
		logger.Error("Agent type validation failed", "error", err)
		return nil, fmt.Errorf("agent type validation failed: %w", err)
	}

	var validateProfileResult activities.ValidateProfileDataOutput
	err = workflow.ExecuteActivity(ctx, a.ValidateProfileDataActivity, activities.ValidateProfileDataInput{
		ProfileData: input.ProfileData,
	}).Get(ctx, &validateProfileResult)
	if err != nil {
		logger.Error("Profile data validation failed", "error", err)
		return nil, fmt.Errorf("profile data validation failed: %w", err)
	}

	// Step 2: Branch by Agent Type
	logger.Info("Step 2: Processing by agent type", "AgentType", input.AgentType)

	switch input.AgentType {
	case domain.AgentTypeDepartmentalEmployee, domain.AgentTypeFieldOfficer:
		if input.EmployeeID != "" {
			// HRMS integration flow
			logger.Info("Processing HRMS integration", "EmployeeID", input.EmployeeID)

			// Validate Employee ID with HRMS
			var validateEmpResult activities.ValidateEmployeeIDOutput
			err = workflow.ExecuteActivity(ctx, a.ValidateEmployeeIDActivity, activities.ValidateEmployeeIDInput{
				EmployeeID: input.EmployeeID,
			}).Get(ctx, &validateEmpResult)
			if err != nil {
				logger.Error("Employee ID validation failed", "error", err)
				return nil, fmt.Errorf("employee ID validation failed: %w", err)
			}

			// Fetch HRMS data
			var fetchHRMSResult activities.FetchHRMSDataOutput
			err = workflow.ExecuteActivity(ctx, a.FetchHRMSDataActivity, activities.FetchHRMSDataInput{
				EmployeeID: input.EmployeeID,
			}).Get(ctx, &fetchHRMSResult)
			if err != nil {
				logger.Error("HRMS data fetch failed", "error", err)
				return nil, fmt.Errorf("HRMS data fetch failed: %w", err)
			}

			// Auto-populate profile
			var autoPopulateResult activities.AutoPopulateProfileOutput
			err = workflow.ExecuteActivity(ctx, a.AutoPopulateProfileActivity, activities.AutoPopulateProfileInput{
				HRMSData:    fetchHRMSResult.HRMSData,
				ProfileData: input.ProfileData,
			}).Get(ctx, &autoPopulateResult)
			if err != nil {
				logger.Error("Profile auto-population failed", "error", err)
				return nil, fmt.Errorf("profile auto-population failed: %w", err)
			}

			// Wait for manual corrections signal (optional)
			logger.Info("Waiting for profile corrections (if needed)")
			correctionChannel := workflow.GetSignalChannel(ctx, "profile-corrections")
			selector := workflow.NewSelector(ctx)

			var correctedFields map[string]interface{}
			correctionReceived := false

			selector.AddReceive(correctionChannel, func(c workflow.ReceiveChannel, more bool) {
				c.Receive(ctx, &correctedFields)
				correctionReceived = true
				logger.Info("Profile corrections received", "fields", correctedFields)
			})

			// Wait for 5 seconds for corrections, then continue
			selector.AddFuture(workflow.NewTimer(ctx, 5*time.Second), func(f workflow.Future) {
				logger.Info("No corrections received, continuing with auto-populated data")
			})

			selector.Select(ctx)

			if correctionReceived && len(correctedFields) > 0 {
				// Apply corrections to profile data
				input.ProfileData = autoPopulateResult.PopulatedProfile
				// Apply corrected fields (merge logic)
				for k, v := range correctedFields {
					switch k {
					case "first_name":
						if s, ok := v.(string); ok {
							input.ProfileData.FirstName = s
						}
					case "mobile_number":
						if s, ok := v.(string); ok {
							input.ProfileData.MobileNumber = s
						}
						// Add more fields as needed
					}
				}
			} else {
				input.ProfileData = autoPopulateResult.PopulatedProfile
			}
		}

	case domain.AgentTypeAdvisor:
		// Validate Advisor Coordinator
		logger.Info("Validating advisor coordinator linkage", "CoordinatorID", input.AdvisorCoordinatorID)
		var validateCoordResult activities.ValidateAdvisorCoordinatorOutput
		err = workflow.ExecuteActivity(ctx, a.ValidateAdvisorCoordinatorActivity, activities.ValidateAdvisorCoordinatorInput{
			CoordinatorID: input.AdvisorCoordinatorID,
		}).Get(ctx, &validateCoordResult)
		if err != nil {
			logger.Error("Advisor coordinator validation failed", "error", err)
			return nil, fmt.Errorf("advisor coordinator validation failed: %w", err)
		}

	case domain.AgentTypeAdvisorCoordinator:
		// Validate geographic assignment
		logger.Info("Validating geographic assignment", "CircleID", input.CircleID, "DivisionID", input.DivisionID)
		if input.CircleID == "" || input.DivisionID == "" {
			return nil, fmt.Errorf("geographic assignment required for advisor coordinator: circle_id and division_id mandatory")
		}
	}

	// Step 3: Validate Business Rules
	logger.Info("Step 3: Validating business rules")

	// Validate PAN uniqueness
	var validatePANResult activities.ValidatePANUniquenessOutput
	err = workflow.ExecuteActivity(ctx, a.ValidatePANUniquenessActivity, activities.ValidatePANUniquenessInput{
		PANNumber: input.ProfileData.PANNumber,
	}).Get(ctx, &validatePANResult)
	if err != nil {
		logger.Error("PAN validation failed", "error", err)
		return nil, fmt.Errorf("PAN validation failed: %w", err)
	}
	if !validatePANResult.IsUnique {
		return nil, fmt.Errorf("PAN number already exists in system")
	}

	// Validate mandatory fields
	var validateMandatoryResult activities.ValidateMandatoryFieldsOutput
	err = workflow.ExecuteActivity(ctx, a.ValidateMandatoryFieldsActivity, activities.ValidateMandatoryFieldsInput{
		AgentType:   input.AgentType,
		ProfileData: input.ProfileData,
	}).Get(ctx, &validateMandatoryResult)
	if err != nil {
		logger.Error("Mandatory fields validation failed", "error", err)
		return nil, fmt.Errorf("mandatory fields validation failed: %w", err)
	}

	// Step 4: Handle Document Uploads
	logger.Info("Step 4: Processing document uploads")
	if len(input.Documents) > 0 {
		// Upload KYC documents
		var uploadDocsResult activities.UploadKYCDocumentsOutput
		err = workflow.ExecuteActivity(ctx, a.UploadKYCDocumentsActivity, activities.UploadKYCDocumentsInput{
			AgentType: input.AgentType,
			Documents: input.Documents,
		}).Get(ctx, &uploadDocsResult)
		if err != nil {
			logger.Error("Document upload failed", "error", err)
			return nil, fmt.Errorf("document upload failed: %w", err)
		}

		// Validate documents
		var validateDocsResult activities.ValidateDocumentsOutput
		err = workflow.ExecuteActivity(ctx, a.ValidateDocumentsActivity, activities.ValidateDocumentsInput{
			DocumentURLs: uploadDocsResult.DocumentURLs,
		}).Get(ctx, &validateDocsResult)
		if err != nil {
			logger.Error("Document validation failed", "error", err)
			return nil, fmt.Errorf("document validation failed: %w", err)
		}
	}

	// Step 5: Check Approval Requirement
	logger.Info("Step 5: Checking approval requirement")
	var approvalCheckResult activities.CheckApprovalRequiredOutput
	err = workflow.ExecuteActivity(ctx, a.CheckApprovalRequiredActivity, activities.CheckApprovalRequiredInput{
		AgentType: input.AgentType,
	}).Get(ctx, &approvalCheckResult)
	if err != nil {
		logger.Error("Approval check failed", "error", err)
		return nil, fmt.Errorf("approval check failed: %w", err)
	}

	if approvalCheckResult.ApprovalRequired {
		logger.Info("Approval required, sending approval request")

		// Send approval request
		var sendApprovalResult activities.SendApprovalRequestOutput
		err = workflow.ExecuteActivity(ctx, a.SendApprovalRequestActivity, activities.SendApprovalRequestInput{
			AgentType:   input.AgentType,
			ProfileData: input.ProfileData,
			InitiatedBy: input.InitiatedBy,
		}).Get(ctx, &sendApprovalResult)
		if err != nil {
			logger.Error("Failed to send approval request", "error", err)
			return nil, fmt.Errorf("failed to send approval request: %w", err)
		}

		// Wait for approval decision signal
		logger.Info("Waiting for approval decision")
		approvalChannel := workflow.GetSignalChannel(ctx, "approval-decision")
		cancelChannel := workflow.GetSignalChannel(ctx, "cancel-onboarding")

		selector := workflow.NewSelector(ctx)
		var approvalDecision activities.ApprovalDecision
		var cancelReason string
		signalReceived := false

		selector.AddReceive(approvalChannel, func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, &approvalDecision)
			signalReceived = true
			logger.Info("Approval decision received", "approved", approvalDecision.Approved)
		})

		selector.AddReceive(cancelChannel, func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, &cancelReason)
			logger.Info("Onboarding cancellation requested", "reason", cancelReason)
		})

		// 7-day timeout for approval
		selector.AddFuture(workflow.NewTimer(ctx, 7*24*time.Hour), func(f workflow.Future) {
			logger.Warn("Approval timeout reached (7 days)")
		})

		selector.Select(ctx)

		if cancelReason != "" {
			return &activities.OnboardingOutput{
				Status:  "CANCELLED",
				Message: fmt.Sprintf("Onboarding cancelled: %s", cancelReason),
			}, nil
		}

		if !signalReceived {
			logger.Warn("Approval timeout, marking as PENDING_APPROVAL")
			return &activities.OnboardingOutput{
				Status:  "PENDING_APPROVAL",
				Message: "Approval timeout - manual follow-up required",
			}, nil
		}

		if !approvalDecision.Approved {
			logger.Info("Onboarding rejected", "reason", approvalDecision.RejectionReason)
			return &activities.OnboardingOutput{
				Status:  "REJECTED",
				Message: approvalDecision.RejectionReason,
			}, nil
		}

		logger.Info("Onboarding approved by", "approver", approvalDecision.ApprovedBy)
	}

	// Step 6: Create Agent Profile
	logger.Info("Step 6: Creating agent profile")

	// Generate agent code
	var generateCodeResult activities.GenerateAgentCodeOutput
	err = workflow.ExecuteActivity(ctx, a.GenerateAgentCodeActivity, activities.GenerateAgentCodeInput{
		AgentType: input.AgentType,
	}).Get(ctx, &generateCodeResult)
	if err != nil {
		logger.Error("Agent code generation failed", "error", err)
		return nil, fmt.Errorf("agent code generation failed: %w", err)
	}

	// Create agent profile in database
	var createProfileResult activities.CreateAgentProfileOutput
	err = workflow.ExecuteActivity(ctx, a.CreateAgentProfileActivity, activities.CreateAgentProfileInput{
		AgentType:            input.AgentType,
		AgentCode:            generateCodeResult.AgentCode,
		EmployeeID:           input.EmployeeID,
		ProfileData:          input.ProfileData,
		AdvisorCoordinatorID: input.AdvisorCoordinatorID,
		CircleID:             input.CircleID,
		DivisionID:           input.DivisionID,
		CreatedBy:            input.InitiatedBy,
	}).Get(ctx, &createProfileResult)
	if err != nil {
		logger.Error("Profile creation failed", "error", err)
		return nil, fmt.Errorf("profile creation failed: %w", err)
	}

	agentID := createProfileResult.AgentID
	agentCode := generateCodeResult.AgentCode
	logger.Info("Agent profile created", "AgentID", agentID, "AgentCode", agentCode)

	// Link to hierarchy
	var linkHierarchyResult activities.LinkToHierarchyOutput
	err = workflow.ExecuteActivity(ctx, a.LinkToHierarchyActivity, activities.LinkToHierarchyInput{
		AgentID:              agentID,
		AgentType:            input.AgentType,
		AdvisorCoordinatorID: input.AdvisorCoordinatorID,
		CircleID:             input.CircleID,
		DivisionID:           input.DivisionID,
	}).Get(ctx, &linkHierarchyResult)
	if err != nil {
		logger.Error("Hierarchy linking failed", "error", err)
		// Compensate: Delete created profile
		logger.Info("Compensating: Deleting created agent profile")
		// Note: Compensation would be handled in production via saga pattern
		return nil, fmt.Errorf("hierarchy linking failed: %w", err)
	}

	// Step 7: Initialize License Tracking
	logger.Info("Step 7: Initializing license tracking")

	// Create license record
	var createLicenseResult activities.CreateLicenseRecordOutput
	err = workflow.ExecuteActivity(ctx, a.CreateLicenseRecordActivity, activities.CreateLicenseRecordInput{
		AgentID:     agentID,
		LicenseType: domain.LicenseTypeProvisional, // Start with provisional
		CreatedBy:   input.InitiatedBy,
	}).Get(ctx, &createLicenseResult)
	if err != nil {
		logger.Warn("License record creation failed (non-critical)", "error", err)
		// Non-critical: Log and continue
	} else {
		logger.Info("License record created", "LicenseID", createLicenseResult.LicenseID)

		// TODO: Start child workflow for license renewal tracking (WF-001)
		// This would be implemented in future when WF-001 is ready
		// childWorkflowOptions := workflow.ChildWorkflowOptions{
		// 	WorkflowID: fmt.Sprintf("license-renewal-%s", agentID),
		// }
		// childCtx := workflow.WithChildOptions(ctx, childWorkflowOptions)
		// workflow.ExecuteChildWorkflow(childCtx, LicenseRenewalTrackingWorkflow, input)
	}

	// Step 8: Send Welcome Notification
	logger.Info("Step 8: Sending welcome notifications")

	// Send welcome email
	err = workflow.ExecuteActivity(ctx, a.SendWelcomeEmailActivity, activities.SendWelcomeEmailInput{
		AgentID:   agentID,
		AgentCode: agentCode,
		Email:     input.ProfileData.Email,
		FirstName: input.ProfileData.FirstName,
		LastName:  input.ProfileData.LastName,
	}).Get(ctx, nil)
	if err != nil {
		logger.Warn("Welcome email failed (non-critical)", "error", err)
	}

	// Send welcome SMS
	err = workflow.ExecuteActivity(ctx, a.SendWelcomeSMSActivity, activities.SendWelcomeSMSInput{
		AgentID:      agentID,
		AgentCode:    agentCode,
		MobileNumber: input.ProfileData.MobileNumber,
		FirstName:    input.ProfileData.FirstName,
	}).Get(ctx, nil)
	if err != nil {
		logger.Warn("Welcome SMS failed (non-critical)", "error", err)
	}

	// Step 9: Complete Onboarding
	logger.Info("Step 9: Completing onboarding")

	result := &activities.OnboardingOutput{
		AgentID:          agentID,
		AgentCode:        agentCode,
		Status:           "ACTIVE",
		Message:          "Agent onboarding completed successfully",
		ProfileCreatedAt: time.Now(),
	}

	logger.Info("AgentOnboardingWorkflow completed successfully",
		"AgentID", agentID,
		"AgentCode", agentCode)

	return result, nil
}
