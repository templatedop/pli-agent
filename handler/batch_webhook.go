package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"pli-agent-api/core/domain"
	req "pli-agent-api/handler/request"
	resp "pli-agent-api/handler/response"
	repo "pli-agent-api/repo/postgres"

	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"
	"go.temporal.io/sdk/client"
)

// AgentBatchWebhookHandler handles batch operations and webhook APIs
// Phase 10: Batch & Webhook APIs
// AGT-038, AGT-064 to AGT-067, AGT-078
// FR-AGT-PRF-012: License Auto-Deactivation
// FR-AGT-PRF-025: Profile Export
// INT-AGT-001: HRMS System Integration
type AgentBatchWebhookHandler struct {
	*serverHandler.Base
	exportRepo     *repo.AgentExportRepository
	webhookRepo    *repo.HRMSWebhookRepository
	profileRepo    *repo.AgentProfileRepository
	licenseRepo    *repo.AgentLicenseRepository
	temporalClient client.Client
	webhookSecret  string
}

// NewAgentBatchWebhookHandler creates a new batch and webhook handler
func NewAgentBatchWebhookHandler(
	exportRepo *repo.AgentExportRepository,
	webhookRepo *repo.HRMSWebhookRepository,
	profileRepo *repo.AgentProfileRepository,
	licenseRepo *repo.AgentLicenseRepository,
	temporalClient client.Client,
) *AgentBatchWebhookHandler {
	return &AgentBatchWebhookHandler{
		Base:           &serverHandler.Base{},
		exportRepo:     exportRepo,
		webhookRepo:    webhookRepo,
		profileRepo:    profileRepo,
		licenseRepo:    licenseRepo,
		temporalClient: temporalClient,
		webhookSecret:  "your-webhook-secret-key", // TODO: Load from config
	}
}

// RegisterRoutes registers all batch and webhook routes
func (h *AgentBatchWebhookHandler) RegisterRoutes() []serverRoute.Route {
	return []serverRoute.Route{
		// AGT-064: Configure Export Parameters
		serverRoute.NewRoute("POST", "/agents/export/configure", h.ConfigureExport),
		// AGT-065: Execute Export Asynchronously
		serverRoute.NewRoute("POST", "/agents/export/execute", h.ExecuteExport),
		// AGT-066: Get Export Status
		serverRoute.NewRoute("GET", "/agents/export/:export_id/status", h.GetExportStatus),
		// AGT-067: Download Exported File
		serverRoute.NewRoute("GET", "/agents/export/:export_id/download", h.DownloadExport),
		// AGT-078: HRMS Webhook Receiver
		serverRoute.NewRoute("POST", "/webhooks/hrms/employee-update", h.HandleHRMSWebhook),
	}
}

// ConfigureExport creates export configuration with filters and fields
// AGT-064: Configure Export Parameters
// FR-AGT-PRF-025: Profile Export
func (h *AgentBatchWebhookHandler) ConfigureExport(
	sctx *serverRoute.Context,
	request req.ConfigureExportRequest,
) (*resp.ConfigureExportResponse, error) {
	log.Info(sctx.Ctx, "Configuring export: %s", request.ExportName)

	// Serialize filters and fields to JSON
	filtersJSON, err := json.Marshal(request.Filters)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize filters: %w", err)
	}

	fieldsJSON, err := json.Marshal(request.Fields)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize fields: %w", err)
	}

	// Estimate record count
	estimatedRecords, err := h.exportRepo.EstimateRecordCount(sctx.Ctx, string(filtersJSON))
	if err != nil {
		log.Warn(sctx.Ctx, "Failed to estimate records: %v", err)
		estimatedRecords = 0
	}

	// Calculate estimated time (rough estimate: 100 records per second)
	estimatedTimeSeconds := estimatedRecords / 100
	if estimatedTimeSeconds < 1 {
		estimatedTimeSeconds = 1
	}

	// Create export configuration
	exportConfig := &domain.AgentExportConfig{
		ExportName:       request.ExportName,
		Filters:          sql.NullString{String: string(filtersJSON), Valid: true},
		Fields:           sql.NullString{String: string(fieldsJSON), Valid: true},
		OutputFormat:     request.OutputFormat,
		EstimatedRecords: estimatedRecords,
		CreatedBy:        request.CreatedBy,
	}

	result, err := h.exportRepo.CreateConfig(sctx.Ctx, exportConfig)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to create export config: %v", err)
		return nil, fmt.Errorf("failed to create export config: %w", err)
	}

	return &resp.ConfigureExportResponse{
		ExportConfigID:       result.ExportConfigID,
		ExportName:           result.ExportName,
		EstimatedRecords:     estimatedRecords,
		EstimatedTimeSeconds: estimatedTimeSeconds,
	}, nil
}

// ExecuteExport starts asynchronous export execution
// AGT-065: Execute Export Asynchronously
// WF-AGT-PRF-012: Profile Export Workflow
func (h *AgentBatchWebhookHandler) ExecuteExport(
	sctx *serverRoute.Context,
	request req.ExecuteExportRequest,
) (*resp.ExecuteExportResponse, error) {
	log.Info(sctx.Ctx, "Executing export for config: %s", request.ExportConfigID)

	// Fetch export configuration
	config, err := h.exportRepo.GetConfigByID(sctx.Ctx, request.ExportConfigID)
	if err != nil {
		log.Error(sctx.Ctx, "Export config not found: %v", err)
		return nil, fmt.Errorf("export config not found: %w", err)
	}

	// Create export job
	workflowID := fmt.Sprintf("profile-export-%s-%d", request.ExportConfigID, time.Now().Unix())
	exportJob := &domain.AgentExportJob{
		ExportConfigID: request.ExportConfigID,
		RequestedBy:    request.RequestedBy,
		Status:         domain.ExportStatusInProgress,
		TotalRecords:   config.EstimatedRecords,
		WorkflowID:     sql.NullString{String: workflowID, Valid: true},
	}

	result, err := h.exportRepo.CreateJob(sctx.Ctx, exportJob)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to create export job: %v", err)
		return nil, fmt.Errorf("failed to create export job: %w", err)
	}

	// TODO: Start Temporal workflow WF-AGT-PRF-012
	// For now, we'll simulate the export process
	// In production, this would start: workflows.ProfileExportWorkflow
	log.Info(sctx.Ctx, "Export job created: %s (workflow would start here)", result.ExportID)

	return &resp.ExecuteExportResponse{
		ExportID: result.ExportID,
		Status:   result.Status,
		Message:  "Export started, check status via AGT-066",
	}, nil
}

// GetExportStatus retrieves export job status
// AGT-066: Get Export Status
func (h *AgentBatchWebhookHandler) GetExportStatus(
	sctx *serverRoute.Context,
	uri ExportIDUri,
) (*resp.GetExportStatusResponse, error) {
	log.Info(sctx.Ctx, "Getting export status: %s", uri.ExportID)

	job, err := h.exportRepo.GetJobByID(sctx.Ctx, uri.ExportID)
	if err != nil {
		log.Error(sctx.Ctx, "Export job not found: %v", err)
		return nil, fmt.Errorf("export job not found: %w", err)
	}

	response := &resp.GetExportStatusResponse{
		ExportID:           job.ExportID,
		Status:             job.Status,
		ProgressPercentage: job.ProgressPercentage,
		RecordsProcessed:   job.RecordsProcessed,
		TotalRecords:       job.TotalRecords,
		StartedAt:          job.StartedAt,
	}

	if job.FileURL.Valid {
		response.FileURL = &job.FileURL.String
	}
	if job.CompletedAt.Valid {
		response.CompletedAt = &job.CompletedAt.Time
	}
	if job.ErrorMessage.Valid {
		response.ErrorMessage = &job.ErrorMessage.String
	}

	return response, nil
}

// DownloadExport returns the exported file for download
// AGT-067: Download Exported File
func (h *AgentBatchWebhookHandler) DownloadExport(
	sctx *serverRoute.Context,
	uri ExportIDUri,
) (*resp.DownloadExportResponse, error) {
	log.Info(sctx.Ctx, "Downloading export: %s", uri.ExportID)

	job, err := h.exportRepo.GetJobByID(sctx.Ctx, uri.ExportID)
	if err != nil {
		log.Error(sctx.Ctx, "Export job not found: %v", err)
		return nil, fmt.Errorf("export job not found: %w", err)
	}

	if job.Status != domain.ExportStatusCompleted {
		return nil, fmt.Errorf("export not completed yet, status: %s", job.Status)
	}

	if !job.FileURL.Valid {
		return nil, fmt.Errorf("export file URL not available")
	}

	// In production, this would stream the file from storage
	// For now, return the URL for client-side download
	return &resp.DownloadExportResponse{
		FileURL:       job.FileURL.String,
		FileName:      fmt.Sprintf("agent-export-%s.xlsx", uri.ExportID),
		FileSizeBytes: job.FileSizeBytes.Int64,
	}, nil
}

// HandleHRMSWebhook processes HRMS employee update webhooks
// AGT-078: HRMS Webhook Receiver
// INT-AGT-001: HRMS System Integration
func (h *AgentBatchWebhookHandler) HandleHRMSWebhook(
	sctx *serverRoute.Context,
	request req.HRMSWebhookRequest,
) (*resp.HRMSWebhookResponse, error) {
	log.Info(sctx.Ctx, "Received HRMS webhook: event_id=%s, type=%s", request.EventID, request.EventType)

	// Validate webhook signature
	signatureValid := h.validateWebhookSignature(request, sctx.Ctx)
	if !signatureValid {
		log.Error(sctx.Ctx, "Invalid webhook signature for event %s", request.EventID)
		return nil, fmt.Errorf("invalid webhook signature")
	}

	// Serialize employee data to JSON
	employeeDataJSON, err := json.Marshal(request.EmployeeData)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize employee data: %w", err)
	}

	// Store webhook event
	webhookEvent := &domain.HRMSWebhookEvent{
		EventID:        request.EventID,
		EventType:      request.EventType,
		EmployeeID:     request.EmployeeData.EmployeeID,
		EmployeeData:   sql.NullString{String: string(employeeDataJSON), Valid: true},
		Signature:      request.Signature,
		SignatureValid: signatureValid,
		Status:         domain.WebhookStatusReceived,
	}

	_, err = h.webhookRepo.CreateEvent(sctx.Ctx, webhookEvent)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to store webhook event: %v", err)
		return nil, fmt.Errorf("failed to store webhook event: %w", err)
	}

	// Process webhook event
	err = h.processWebhookEvent(sctx, request)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to process webhook event: %v", err)
		// Update status to failed
		h.webhookRepo.UpdateEventStatus(sctx.Ctx, request.EventID, domain.WebhookStatusFailed, nil, &err.Error())
		return nil, fmt.Errorf("failed to process webhook: %w", err)
	}

	// Update status to processed
	successResult := `{"message":"Webhook processed successfully"}`
	h.webhookRepo.UpdateEventStatus(sctx.Ctx, request.EventID, domain.WebhookStatusProcessed, &successResult, nil)

	return &resp.HRMSWebhookResponse{
		Status:  "PROCESSED",
		EventID: request.EventID,
		Message: "Webhook processed successfully",
	}, nil
}

// Helper: Validate webhook signature
func (h *AgentBatchWebhookHandler) validateWebhookSignature(request req.HRMSWebhookRequest, ctx interface{}) bool {
	// Serialize request for signature validation
	payload, err := json.Marshal(request.EmployeeData)
	if err != nil {
		return false
	}

	// Calculate expected signature using HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(h.webhookSecret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expectedSignature), []byte(request.Signature))
}

// Helper: Process webhook event based on type
func (h *AgentBatchWebhookHandler) processWebhookEvent(sctx *serverRoute.Context, request req.HRMSWebhookRequest) error {
	switch request.EventType {
	case domain.WebhookEventEmployeeUpdated:
		return h.handleEmployeeUpdated(sctx, request.EmployeeData)
	case domain.WebhookEventEmployeeTerminated:
		return h.handleEmployeeTerminated(sctx, request.EmployeeData)
	case domain.WebhookEventEmployeeTransferred:
		return h.handleEmployeeTransferred(sctx, request.EmployeeData)
	case domain.WebhookEventEmployeeCreated:
		log.Info(sctx.Ctx, "Employee created event - manual profile creation required")
		return nil
	default:
		return fmt.Errorf("unknown event type: %s", request.EventType)
	}
}

func (h *AgentBatchWebhookHandler) handleEmployeeUpdated(sctx *serverRoute.Context, data domain.HRMSEmployeeData) error {
	// Find agent by employee ID
	// TODO: Implement FindByEmployeeID in profile repository
	log.Info(sctx.Ctx, "Processing employee update for: %s", data.EmployeeID)
	return nil
}

func (h *AgentBatchWebhookHandler) handleEmployeeTerminated(sctx *serverRoute.Context, data domain.HRMSEmployeeData) error {
	log.Info(sctx.Ctx, "Processing employee termination for: %s", data.EmployeeID)
	// TODO: Trigger agent termination workflow
	return nil
}

func (h *AgentBatchWebhookHandler) handleEmployeeTransferred(sctx *serverRoute.Context, data domain.HRMSEmployeeData) error {
	log.Info(sctx.Ctx, "Processing employee transfer for: %s", data.EmployeeID)
	// TODO: Update office code for agent
	return nil
}

// URI parameter structs
type ExportIDUri struct {
	ExportID string `uri:"export_id" validate:"required,uuid4"`
}
