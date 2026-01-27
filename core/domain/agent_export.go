package domain

import (
	"database/sql"
	"time"
)

// AgentExportConfig represents an export configuration template
// Phase 10: Batch & Webhook APIs
// AGT-064: Configure Export Parameters
// FR-AGT-PRF-025: Profile Export
type AgentExportConfig struct {
	ExportConfigID   string         `json:"export_config_id" db:"export_config_id"`
	ExportName       string         `json:"export_name" db:"export_name"`
	Filters          sql.NullString `json:"filters" db:"filters"`           // JSONB - export filters
	Fields           sql.NullString `json:"fields" db:"fields"`             // JSONB - fields to export
	OutputFormat     string         `json:"output_format" db:"output_format"` // EXCEL, PDF, CSV
	EstimatedRecords int            `json:"estimated_records" db:"estimated_records"`
	CreatedBy        string         `json:"created_by" db:"created_by"`
	CreatedAt        time.Time      `json:"created_at" db:"created_at"`
}

// AgentExportJob represents an asynchronous export job execution
// Phase 10: Batch & Webhook APIs
// AGT-065: Execute Export Asynchronously
// AGT-066: Get Export Status
// AGT-067: Download Exported File
// WF-AGT-PRF-012: Profile Export Workflow
type AgentExportJob struct {
	ExportID          string         `json:"export_id" db:"export_id"`
	ExportConfigID    string         `json:"export_config_id" db:"export_config_id"`
	RequestedBy       string         `json:"requested_by" db:"requested_by"`
	Status            string         `json:"status" db:"status"` // IN_PROGRESS, COMPLETED, FAILED, CANCELLED
	ProgressPercentage int           `json:"progress_percentage" db:"progress_percentage"`
	RecordsProcessed  int            `json:"records_processed" db:"records_processed"`
	TotalRecords      int            `json:"total_records" db:"total_records"`
	FileURL           sql.NullString `json:"file_url" db:"file_url"`
	FileSizeBytes     sql.NullInt64  `json:"file_size_bytes" db:"file_size_bytes"`
	WorkflowID        sql.NullString `json:"workflow_id" db:"workflow_id"`
	StartedAt         time.Time      `json:"started_at" db:"started_at"`
	CompletedAt       sql.NullTime   `json:"completed_at" db:"completed_at"`
	ErrorMessage      sql.NullString `json:"error_message" db:"error_message"`
	Metadata          sql.NullString `json:"metadata" db:"metadata"` // JSONB
}

// Export job status constants
const (
	ExportStatusInProgress = "IN_PROGRESS"
	ExportStatusCompleted  = "COMPLETED"
	ExportStatusFailed     = "FAILED"
	ExportStatusCancelled  = "CANCELLED"
)

// Export output format constants
const (
	ExportFormatExcel = "EXCEL"
	ExportFormatPDF   = "PDF"
	ExportFormatCSV   = "CSV"
)

// HRMSWebhookEvent represents an incoming HRMS webhook event
// Phase 10: Batch & Webhook APIs
// AGT-078: HRMS Webhook Receiver
// INT-AGT-001: HRMS System Integration
type HRMSWebhookEvent struct {
	EventID         string         `json:"event_id" db:"event_id"`
	EventType       string         `json:"event_type" db:"event_type"` // EMPLOYEE_CREATED, EMPLOYEE_UPDATED, etc.
	EmployeeID      string         `json:"employee_id" db:"employee_id"`
	EmployeeData    sql.NullString `json:"employee_data" db:"employee_data"` // JSONB
	Signature       string         `json:"signature" db:"signature"`
	SignatureValid  bool           `json:"signature_valid" db:"signature_valid"`
	ReceivedAt      time.Time      `json:"received_at" db:"received_at"`
	ProcessedAt     sql.NullTime   `json:"processed_at" db:"processed_at"`
	Status          string         `json:"status" db:"status"` // RECEIVED, PROCESSING, PROCESSED, FAILED, RETRYING
	ProcessingResult sql.NullString `json:"processing_result" db:"processing_result"` // JSONB
	ErrorMessage    sql.NullString `json:"error_message" db:"error_message"`
	RetryCount      int            `json:"retry_count" db:"retry_count"`
	NextRetryAt     sql.NullTime   `json:"next_retry_at" db:"next_retry_at"`
}

// Webhook event type constants
const (
	WebhookEventEmployeeCreated     = "EMPLOYEE_CREATED"
	WebhookEventEmployeeUpdated     = "EMPLOYEE_UPDATED"
	WebhookEventEmployeeTransferred = "EMPLOYEE_TRANSFERRED"
	WebhookEventEmployeeTerminated  = "EMPLOYEE_TERMINATED"
)

// Webhook event status constants
const (
	WebhookStatusReceived   = "RECEIVED"
	WebhookStatusProcessing = "PROCESSING"
	WebhookStatusProcessed  = "PROCESSED"
	WebhookStatusFailed     = "FAILED"
	WebhookStatusRetrying   = "RETRYING"
)

// AgentBatchOperationLog represents a batch operation log
// Phase 10: Batch & Webhook APIs
// AGT-038: Batch Deactivate Expired Licenses
// WF-AGT-PRF-007: License Deactivation Workflow
type AgentBatchOperationLog struct {
	BatchID         string         `json:"batch_id" db:"batch_id"`
	OperationType   string         `json:"operation_type" db:"operation_type"` // LICENSE_DEACTIVATION, STATUS_UPDATE, etc.
	BatchDate       time.Time      `json:"batch_date" db:"batch_date"`
	WorkflowID      sql.NullString `json:"workflow_id" db:"workflow_id"`
	TotalAgents     int            `json:"total_agents" db:"total_agents"`
	AgentsProcessed int            `json:"agents_processed" db:"agents_processed"`
	AgentsSucceeded int            `json:"agents_succeeded" db:"agents_succeeded"`
	AgentsFailed    int            `json:"agents_failed" db:"agents_failed"`
	DryRun          bool           `json:"dry_run" db:"dry_run"`
	AgentIDs        []string       `json:"agent_ids" db:"agent_ids"` // PostgreSQL array
	StartedAt       time.Time      `json:"started_at" db:"started_at"`
	CompletedAt     sql.NullTime   `json:"completed_at" db:"completed_at"`
	Status          string         `json:"status" db:"status"` // IN_PROGRESS, COMPLETED, FAILED
	ErrorSummary    sql.NullString `json:"error_summary" db:"error_summary"`
}

// Batch operation type constants
const (
	BatchOperationLicenseDeactivation = "LICENSE_DEACTIVATION"
	BatchOperationStatusUpdate        = "STATUS_UPDATE"
	BatchOperationBulkNotification    = "BULK_NOTIFICATION"
)

// Batch operation status constants
const (
	BatchStatusInProgress = "IN_PROGRESS"
	BatchStatusCompleted  = "COMPLETED"
	BatchStatusFailed     = "FAILED"
)

// ExportFilters represents filters for agent export
// Used in export configuration
type ExportFilters struct {
	Status     *string    `json:"status,omitempty"`
	OfficeCode *string    `json:"office_code,omitempty"`
	AgentType  *string    `json:"agent_type,omitempty"`
	FromDate   *time.Time `json:"from_date,omitempty"`
	ToDate     *time.Time `json:"to_date,omitempty"`
}

// ExportFields represents fields to include in export
// Used in export configuration
type ExportFields []string

// HRMSEmployeeData represents employee data from HRMS webhook
// Parsed from employee_data JSONB
type HRMSEmployeeData struct {
	EmployeeID      string     `json:"employee_id"`
	EmployeeName    string     `json:"employee_name"`
	Name            string     `json:"name,omitempty"` // Alias for EmployeeName
	Department      string     `json:"department,omitempty"`
	Designation     string     `json:"designation,omitempty"`
	OfficeCode      string     `json:"office_code,omitempty"`
	Status          string     `json:"status,omitempty"`
	EmailAddress    string     `json:"email_address,omitempty"`
	PhoneNumber     string     `json:"phone_number,omitempty"`
	TerminationDate *time.Time `json:"termination_date,omitempty"`
	TransferDate    *time.Time `json:"transfer_date,omitempty"`
}
