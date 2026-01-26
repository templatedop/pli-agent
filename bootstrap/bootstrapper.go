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
		repo.NewAgentProfileSessionRepository, // Phase 5: Session management
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
		// TODO: PHASE 6 - Add profile update handlers
		// TODO: PHASE 7 - Add license management handlers
		// TODO: PHASE 8 - Add status management handlers
		// TODO: PHASE 9 - Add search and dashboard handlers
		// TODO: PHASE 10 - Add batch and webhook handlers
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
		activities.NewAgentOnboardingActivities,
	),

	fx.Invoke(
		// Register workflows and activities with worker
		func(lc fx.Lifecycle, c client.Client, cfg *config.Config, activities *activities.AgentOnboardingActivities) error {
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

			// Register workflows
			// WF-002: Agent Onboarding Workflow
			w.RegisterWorkflow(workflows.AgentOnboardingWorkflow)
			// Agent Approval Child Workflow (human-in-the-loop pattern)
			w.RegisterWorkflow(workflows.AgentApprovalWorkflow)

			// Register all activities for WF-002
			// RecordWorkflowStartActivity (FIRST activity - makes workflow self-recording)
			w.RegisterActivity(activities.RecordWorkflowStartActivity)
			// ACT-011: ValidateAgentTypeActivity
			w.RegisterActivity(activities.ValidateAgentTypeActivity)
			// ACT-012: ValidateProfileDataActivity
			w.RegisterActivity(activities.ValidateProfileDataActivity)
			// ACT-013: ValidateEmployeeIDActivity
			w.RegisterActivity(activities.ValidateEmployeeIDActivity)
			// ACT-014: FetchHRMSDataActivity
			w.RegisterActivity(activities.FetchHRMSDataActivity)
			// ACT-015: AutoPopulateProfileActivity
			w.RegisterActivity(activities.AutoPopulateProfileActivity)
			// ACT-016: ValidateAdvisorCoordinatorActivity
			w.RegisterActivity(activities.ValidateAdvisorCoordinatorActivity)
			// ACT-017: ValidatePANUniquenessActivity
			w.RegisterActivity(activities.ValidatePANUniquenessActivity)
			// ACT-018: ValidateMandatoryFieldsActivity
			w.RegisterActivity(activities.ValidateMandatoryFieldsActivity)
			// ACT-019: UploadKYCDocumentsActivity
			w.RegisterActivity(activities.UploadKYCDocumentsActivity)
			// ACT-020: ValidateDocumentsActivity
			w.RegisterActivity(activities.ValidateDocumentsActivity)
			// ACT-021: CheckApprovalRequiredActivity
			w.RegisterActivity(activities.CheckApprovalRequiredActivity)
			// ACT-022: SendApprovalRequestActivity
			w.RegisterActivity(activities.SendApprovalRequestActivity)
			// ACT-023: GenerateAgentCodeActivity
			w.RegisterActivity(activities.GenerateAgentCodeActivity)
			// ACT-024: CreateAgentProfileActivity
			w.RegisterActivity(activities.CreateAgentProfileActivity)
			// ACT-025: LinkToHierarchyActivity
			w.RegisterActivity(activities.LinkToHierarchyActivity)
			// ACT-026: CreateLicenseRecordActivity
			w.RegisterActivity(activities.CreateLicenseRecordActivity)
			// ACT-027: SendWelcomeEmailActivity
			w.RegisterActivity(activities.SendWelcomeEmailActivity)
			// ACT-028: SendWelcomeSMSActivity
			w.RegisterActivity(activities.SendWelcomeSMSActivity)
			// SendApprovalNotificationActivity (used by approval child workflow)
			w.RegisterActivity(activities.SendApprovalNotificationActivity)

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
