package bootstrap

import (
	"context"

	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"

	handler "pli-agent-api/handler"
	repo "pli-agent-api/repo/postgres"
	"pli-agent-api/workflows"
	"pli-agent-api/workflows/activities"
)

// FxRepo module provides all repository implementations
// Repositories handle database operations for all entities
var FxRepo = fx.Module(
	"Repomodule",
	fx.Provide(
		repo.NewAgentProfileRepository,
		repo.NewAgentAddressRepository,
		repo.NewAgentContactRepository,
		repo.NewAgentEmailRepository,
		repo.NewAgentBankDetailsRepository,
		repo.NewAgentLicenseRepository,
		repo.NewAgentAuditLogRepository,
		repo.NewAgentProfileSessionRepository,       // Phase 5: Session management
		repo.NewAgentProfileUpdateRequestRepository, // Phase 6.1: Approval workflow
		repo.NewAgentProfileFieldMetadataRepository, // Phase 6.2: Dynamic field metadata
		repo.NewAgentTerminationRepository,          // Phase 8: Status management
		repo.NewAgentNotificationRepository,         // Phase 9: Notification history
		repo.NewAgentExportRepository,               // Phase 10: Export operations
		repo.NewHRMSWebhookRepository,               // Phase 10: HRMS webhook events
		// Add more repository constructors here as needed
	),
)

// FxHandler module provides all HTTP handlers
// Each handler must be annotated to implement serverHandler.Handler interface
var FxHandler = fx.Module(
	"Handlermodule",
	fx.Provide(
		// PHASE 4: Lookup & Validation APIs (AGT-007 to AGT-021)
		// Agent Lookup Handler (AGT-007 to AGT-011)
		fx.Annotate(
			handler.NewAgentLookupHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),
		// Agent Validation Handler (AGT-012 to AGT-015)
		fx.Annotate(
			handler.NewAgentValidationHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),
		// Agent Workflow Handler (AGT-016 to AGT-021)
		fx.Annotate(
			handler.NewAgentWorkflowHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),
		// PHASE 5: Profile Creation & Session Management (AGT-001 to AGT-006, AGT-016 to AGT-019)
		// Agent Profile Creation Handler with Temporal WF-002 integration
		fx.Annotate(
			handler.NewAgentProfileCreationHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),
		// PHASE 6: Profile Update APIs (AGT-022 to AGT-028)
		// Agent Profile Update Handler with multi-criteria search and audit history
		fx.Annotate(
			handler.NewAgentProfileUpdateHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),
		// PHASE 7: License Management APIs (AGT-029 to AGT-038)
		// Agent License Handler with renewal rules and expiry management
		fx.Annotate(
			handler.NewAgentLicenseHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),
		// PHASE 8: Status Management APIs (AGT-039 to AGT-041)
		// Agent Status Management Handler with termination and reinstatement workflows
		fx.Annotate(
			handler.NewAgentStatusManagementHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),
		// PHASE 9: Search & Dashboard APIs (AGT-022, AGT-023, AGT-028, AGT-068, AGT-073, AGT-076, AGT-077)
		// Agent Search & Dashboard Handler with multi-criteria search, hierarchy, timeline, notifications
		fx.Annotate(
			handler.NewAgentSearchDashboardHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),
		// PHASE 10: Batch & Webhook APIs (AGT-064 to AGT-067, AGT-078)
		// Agent Batch & Webhook Handler with export operations and HRMS integration
		fx.Annotate(
			handler.NewAgentBatchWebhookHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),
	),
)

// FxTemporal module provides Temporal client and worker for Agent Profile Management workflows
// Phase 5: WF-002 - Agent Onboarding Workflow
var FxTemporal = fx.Module(
	"Temporalmodule",
	fx.Provide(
		// Provide Temporal client
		func(cfg *config.Config) (client.Client, error) {
			hostPort := cfg.GetString("temporal.hostport")
			if hostPort == "" {
				hostPort = "localhost:7233" // Default Temporal server address
			}

			return client.NewClient(client.Options{
				HostPort:  hostPort,
				Namespace: cfg.GetString("temporal.namespace"),
			})
		},

		// Provide activity structs with repository dependencies
		// Phase 5: Agent Onboarding Activities (WF-002)
		activities.NewAgentOnboardingActivities,

		// Phase 8: License Deactivation Activities (WF-AGT-PRF-007)
		activities.NewLicenseDeactivationActivities,

		// Phase 8: Agent Termination Activities (WF-AGT-PRF-004)
		activities.NewAgentTerminationActivities,

		// Phase 8: Agent Reinstatement Activities (WF-AGT-PRF-011)
		activities.NewAgentReinstatementActivities,
	),

	fx.Invoke(
		// Register workflows and activities with worker
		func(lc fx.Lifecycle, c client.Client, cfg *config.Config,
			onboardingActivities *activities.AgentOnboardingActivities,
			deactivationActivities *activities.LicenseDeactivationActivities,
			terminationActivities *activities.AgentTerminationActivities,
			reinstatementActivities *activities.AgentReinstatementActivities,
		) error {
			taskQueue := cfg.GetString("temporal.taskqueue")
			if taskQueue == "" {
				taskQueue = "agent-profile-task-queue" // Default task queue
			}

			w := worker.New(c, taskQueue, worker.Options{
				MaxConcurrentWorkflowTaskExecutionSize:  cfg.GetInt("temporal.worker.max_concurrent_workflow"),
				MaxConcurrentActivityExecutionSize:      cfg.GetInt("temporal.worker.max_concurrent_activities"),
				MaxConcurrentLocalActivityExecutionSize: cfg.GetInt("temporal.worker.max_concurrent_local_activities"),
				MaxConcurrentActivityTaskPollers:        cfg.GetInt("temporal.worker.max_pollers"),
			})

			// ========================================
			// Register All Workflows
			// ========================================

			// PHASE 5: WF-002 - Agent Onboarding Workflow
			w.RegisterWorkflow(workflows.AgentOnboardingWorkflow)
			// Agent Approval Child Workflow (human-in-the-loop pattern)
			w.RegisterWorkflow(workflows.AgentApprovalWorkflow)

			// PHASE 8: WF-AGT-PRF-007 - License Deactivation Workflow (Scheduled)
			w.RegisterWorkflow(workflows.LicenseDeactivationWorkflow)

			// PHASE 8: WF-AGT-PRF-004 - Agent Termination Workflow
			w.RegisterWorkflow(workflows.AgentTerminationWorkflow)

			// PHASE 8: WF-AGT-PRF-011 - Agent Reinstatement Workflow
			w.RegisterWorkflow(workflows.AgentReinstatementWorkflow)

			// ========================================
			// Register All Activities
			// ========================================

			// PHASE 5: Agent Onboarding Activities (WF-002)
			// RecordWorkflowStartActivity (FIRST activity - makes workflow self-recording)
			w.RegisterActivity(onboardingActivities.RecordWorkflowStartActivity)
			// ACT-011: ValidateAgentTypeActivity
			w.RegisterActivity(onboardingActivities.ValidateAgentTypeActivity)
			// ACT-012: ValidateProfileDataActivity
			w.RegisterActivity(onboardingActivities.ValidateProfileDataActivity)
			// ACT-013: ValidateEmployeeIDActivity
			w.RegisterActivity(onboardingActivities.ValidateEmployeeIDActivity)
			// ACT-014: FetchHRMSDataActivity
			w.RegisterActivity(onboardingActivities.FetchHRMSDataActivity)
			// ACT-015: AutoPopulateProfileActivity
			w.RegisterActivity(onboardingActivities.AutoPopulateProfileActivity)
			// ACT-016: ValidateAdvisorCoordinatorActivity
			w.RegisterActivity(onboardingActivities.ValidateAdvisorCoordinatorActivity)
			// ACT-017: ValidatePANUniquenessActivity
			w.RegisterActivity(onboardingActivities.ValidatePANUniquenessActivity)
			// ACT-018: ValidateMandatoryFieldsActivity
			w.RegisterActivity(onboardingActivities.ValidateMandatoryFieldsActivity)
			// ACT-019: UploadKYCDocumentsActivity
			w.RegisterActivity(onboardingActivities.UploadKYCDocumentsActivity)
			// ACT-020: ValidateDocumentsActivity
			w.RegisterActivity(onboardingActivities.ValidateDocumentsActivity)
			// ACT-021: CheckApprovalRequiredActivity
			w.RegisterActivity(onboardingActivities.CheckApprovalRequiredActivity)
			// ACT-022: SendApprovalRequestActivity
			w.RegisterActivity(onboardingActivities.SendApprovalRequestActivity)
			// ACT-023: GenerateAgentCodeActivity
			w.RegisterActivity(onboardingActivities.GenerateAgentCodeActivity)
			// ACT-024: CreateAgentProfileActivity
			w.RegisterActivity(onboardingActivities.CreateAgentProfileActivity)
			// ACT-025: LinkToHierarchyActivity
			w.RegisterActivity(onboardingActivities.LinkToHierarchyActivity)
			// ACT-026: CreateLicenseRecordActivity
			w.RegisterActivity(onboardingActivities.CreateLicenseRecordActivity)
			// ACT-027: SendWelcomeEmailActivity
			w.RegisterActivity(onboardingActivities.SendWelcomeEmailActivity)
			// ACT-028: SendWelcomeSMSActivity
			w.RegisterActivity(onboardingActivities.SendWelcomeSMSActivity)
			// SendApprovalNotificationActivity (used by approval child workflow)
			w.RegisterActivity(onboardingActivities.SendApprovalNotificationActivity)

			// PHASE 8: License Deactivation Activities (WF-AGT-PRF-007)
			// ACT-DEC-001: Find Expired Licenses
			w.RegisterActivity(deactivationActivities.FindExpiredLicensesActivity)
			// ACT-DEC-002: Batch Update Agent Status
			w.RegisterActivity(deactivationActivities.BatchUpdateAgentStatusActivity)
			// ACT-DEC-003: Batch Disable Portal Access
			w.RegisterActivity(deactivationActivities.BatchDisablePortalAccessActivity)
			// ACT-DEC-004: Batch Stop Commission
			w.RegisterActivity(deactivationActivities.BatchStopCommissionActivity)
			// ACT-DEC-005: Batch Send Notification
			w.RegisterActivity(deactivationActivities.BatchSendNotificationActivity)
			// ACT-DEC-006: Create Batch Audit Log
			w.RegisterActivity(deactivationActivities.CreateBatchAuditLogActivity)

			// PHASE 8: Agent Termination Activities (WF-AGT-PRF-004)
			// ACT-TERM-001: Disable Portal Access
			w.RegisterActivity(terminationActivities.DisablePortalAccessActivity)
			// ACT-TERM-002: Stop Commission Processing
			w.RegisterActivity(terminationActivities.StopCommissionProcessingActivity)
			// ACT-TERM-003: Generate Termination Letter
			w.RegisterActivity(terminationActivities.GenerateTerminationLetterActivity)
			// ACT-TERM-004: Archive Agent Data
			w.RegisterActivity(terminationActivities.ArchiveAgentDataActivity)
			// ACT-TERM-005: Send Termination Notifications
			w.RegisterActivity(terminationActivities.SendTerminationNotificationsActivity)
			// ACT-TERM-006: Update Termination Record
			w.RegisterActivity(terminationActivities.UpdateTerminationRecordActivity)

			// PHASE 8: Agent Reinstatement Activities (WF-AGT-PRF-011)
			// ACT-REINST-001: Send Approval Request Notification
			w.RegisterActivity(reinstatementActivities.SendApprovalRequestNotificationActivity)
			// ACT-REINST-002: Approve Reinstatement
			w.RegisterActivity(reinstatementActivities.ApproveReinstatementActivity)
			// ACT-REINST-003: Reject Reinstatement
			w.RegisterActivity(reinstatementActivities.RejectReinstatementActivity)
			// ACT-REINST-004: Restore Portal Access
			w.RegisterActivity(reinstatementActivities.RestorePortalAccessActivity)
			// ACT-REINST-005: Send Approval Confirmation Notification
			w.RegisterActivity(reinstatementActivities.SendReinstatementApprovalNotificationActivity)
			// ACT-REINST-006: Send Rejection Notification
			w.RegisterActivity(reinstatementActivities.SendReinstatementRejectionNotificationActivity)

			// Start worker in lifecycle
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					return w.Start()
				},
				OnStop: func(ctx context.Context) error {
					w.Stop()
					c.Close()
					return nil
				},
			})

			return nil
		},
	),
)
