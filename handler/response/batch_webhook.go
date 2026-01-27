package response

import (
	"time"
)

// ========================================================================
// PHASE 10: BATCH & WEBHOOK API RESPONSES
// ========================================================================

// ConfigureExportResponse for AGT-064
// FR-AGT-PRF-025: Profile Export Configuration
type ConfigureExportResponse struct {
	ExportConfigID       string `json:"export_config_id"`
	ExportName           string `json:"export_name"`
	EstimatedRecords     int    `json:"estimated_records"`
	EstimatedTimeSeconds int    `json:"estimated_time_seconds"`
}

// ExecuteExportResponse for AGT-065
// WF-AGT-PRF-012: Profile Export Workflow
type ExecuteExportResponse struct {
	ExportID string `json:"export_id"`
	Status   string `json:"status"`
	Message  string `json:"message"`
}

// GetExportStatusResponse for AGT-066
// Returns current export job status
type GetExportStatusResponse struct {
	ExportID           string     `json:"export_id"`
	Status             string     `json:"status"`
	ProgressPercentage int        `json:"progress_percentage"`
	RecordsProcessed   int        `json:"records_processed"`
	TotalRecords       int        `json:"total_records"`
	FileURL            *string    `json:"file_url,omitempty"`
	StartedAt          time.Time  `json:"started_at"`
	CompletedAt        *time.Time `json:"completed_at,omitempty"`
	ErrorMessage       *string    `json:"error_message,omitempty"`
}

// DownloadExportResponse for AGT-067
// Returns file information for download
type DownloadExportResponse struct {
	FileURL       string `json:"file_url"`
	FileName      string `json:"file_name"`
	FileSizeBytes int64  `json:"file_size_bytes"`
}

// HRMSWebhookResponse for AGT-078
// Acknowledges webhook receipt and processing
type HRMSWebhookResponse struct {
	Status  string `json:"status"`
	EventID string `json:"event_id"`
	Message string `json:"message"`
}
