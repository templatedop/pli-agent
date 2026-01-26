package main

import (
	"context"
	"pli-agent-api/bootstrap"

	bootstrapper "gitlab.cept.gov.in/it-2.0-common/n-api-bootstrapper"
)

// main is the entry point for the Agent Profile Management API
// This application handles complete lifecycle management of insurance agents
// including onboarding, profile maintenance, licensing, and status management
func main() {
	app := bootstrapper.New().Options(
		bootstrap.FxHandler,  // Register all HTTP handlers
		bootstrap.FxRepo,     // Register all repositories
		bootstrap.FxTemporal, // Register Temporal workflows (Phase 5: WF-002 Agent Onboarding)
	)
	app.WithContext(context.Background()).Run()
}
