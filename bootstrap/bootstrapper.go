package bootstrap

import (
	"go.uber.org/fx"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"

	handler "pli-agent-api/handler"
	repo "pli-agent-api/repo/postgres"
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
		// Add more repository constructors here as needed
	),
)

// FxHandler module provides all HTTP handlers
// Each handler must be annotated to implement serverHandler.Handler interface
var FxHandler = fx.Module(
	"Handlermodule",
	fx.Provide(
		// Agent Profile Management Handler
		fx.Annotate(
			handler.NewAgentProfileHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),
		// Agent Lookup Handler
		fx.Annotate(
			handler.NewAgentLookupHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),
		// Agent License Handler
		fx.Annotate(
			handler.NewAgentLicenseHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),
		// Agent Status Handler
		fx.Annotate(
			handler.NewAgentStatusHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),
		// Agent Search Handler
		fx.Annotate(
			handler.NewAgentSearchHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),
		// Add more handler constructors here as needed
	),
)

// FxTemporal module provides Temporal client and worker (Optional - for workflows)
// Uncomment when implementing long-running workflows
// var FxTemporal = fx.Module(
// 	"Temporalmodule",
// 	fx.Provide(
// 		// Provide Temporal client
// 		func(cfg *config.Config) (client.Client, error) {
// 			return client.NewClient(client.Options{
// 				HostPort: cfg.GetString("temporal.hostport"),
// 			})
// 		},
//
// 		// Provide activity structs
// 		// activities.NewAgentProfileActivities,
// 	),
//
// 	fx.Invoke(
// 		// Register workflows and activities with worker
// 		func(c client.Client) error {
// 			w := worker.New(c, "agent-profile-management-queue", worker.Options{})
//
// 			// Register workflows
// 			// w.RegisterWorkflow(workflows.AgentOnboardingWorkflow)
//
// 			// Register activities
// 			// w.RegisterActivity(activities.ValidateHRMSEmployee)
//
// 			return w.Start()
// 		},
// 	),
// )
