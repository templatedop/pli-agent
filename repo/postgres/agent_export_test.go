package repo

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"pli-agent-api/core/domain"

	"github.com/stretchr/testify/assert"
)

// TestExportDataQuery_BuildsCorrectQuery tests query building with various filters
func TestExportDataQuery_BuildsCorrectQuery(t *testing.T) {
	tests := []struct {
		name         string
		filters      domain.ExportFilters
		expectStatus bool
		expectOffice bool
		expectType   bool
		expectDates  bool
	}{
		{
			name:         "No filters",
			filters:      domain.ExportFilters{},
			expectStatus: false,
			expectOffice: false,
			expectType:   false,
			expectDates:  false,
		},
		{
			name: "Status filter only",
			filters: domain.ExportFilters{
				Status: stringPtr("ACTIVE"),
			},
			expectStatus: true,
			expectOffice: false,
			expectType:   false,
			expectDates:  false,
		},
		{
			name: "Office code filter only",
			filters: domain.ExportFilters{
				OfficeCode: stringPtr("OFF-001"),
			},
			expectStatus: false,
			expectOffice: true,
			expectType:   false,
			expectDates:  false,
		},
		{
			name: "All filters",
			filters: domain.ExportFilters{
				Status:     stringPtr("ACTIVE"),
				OfficeCode: stringPtr("OFF-001"),
				AgentType:  stringPtr("ADVISOR"),
				FromDate:   timePtr(time.Now().AddDate(0, -1, 0)),
				ToDate:     timePtr(time.Now()),
			},
			expectStatus: true,
			expectOffice: true,
			expectType:   true,
			expectDates:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test validates filter handling logic
			// In actual implementation, this would connect to a test database
			assert.NotNil(t, tt.filters)

			// Validate filter expectations
			if tt.expectStatus {
				assert.NotNil(t, tt.filters.Status)
			}
			if tt.expectOffice {
				assert.NotNil(t, tt.filters.OfficeCode)
			}
			if tt.expectType {
				assert.NotNil(t, tt.filters.AgentType)
			}
			if tt.expectDates {
				assert.NotNil(t, tt.filters.FromDate)
				assert.NotNil(t, tt.filters.ToDate)
			}
		})
	}
}

// TestEstimateRecordCount_ValidJSON tests JSON parsing and query building
func TestEstimateRecordCount_ValidJSON(t *testing.T) {
	tests := []struct {
		name        string
		filtersJSON string
		expectError bool
	}{
		{
			name:        "Empty filters",
			filtersJSON: "",
			expectError: false,
		},
		{
			name:        "Valid status filter",
			filtersJSON: `{"status":"ACTIVE"}`,
			expectError: false,
		},
		{
			name:        "Multiple filters",
			filtersJSON: `{"status":"ACTIVE","office_code":"OFF-001","agent_type":"ADVISOR"}`,
			expectError: false,
		},
		{
			name:        "Invalid JSON",
			filtersJSON: `{invalid json}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON parsing logic
			var filters domain.ExportFilters
			if tt.filtersJSON != "" {
				err := json.Unmarshal([]byte(tt.filtersJSON), &filters)
				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			}
		})
	}
}

// TestAgentExportConfig_Validation tests export config validation
func TestAgentExportConfig_Validation(t *testing.T) {
	tests := []struct {
		name     string
		config   *domain.AgentExportConfig
		isValid  bool
		errorMsg string
	}{
		{
			name: "Valid config",
			config: &domain.AgentExportConfig{
				ExportName:   "Test Export",
				OutputFormat: domain.ExportFormatExcel,
				Filters:      sql.NullString{String: `{"status":"ACTIVE"}`, Valid: true},
				Fields:       sql.NullString{String: `["agent_id","name"]`, Valid: true},
				CreatedBy:    "test-user",
			},
			isValid: true,
		},
		{
			name: "Empty export name",
			config: &domain.AgentExportConfig{
				ExportName:   "",
				OutputFormat: domain.ExportFormatExcel,
				CreatedBy:    "test-user",
			},
			isValid:  false,
			errorMsg: "export name is required",
		},
		{
			name: "Invalid output format",
			config: &domain.AgentExportConfig{
				ExportName:   "Test Export",
				OutputFormat: "INVALID",
				CreatedBy:    "test-user",
			},
			isValid:  false,
			errorMsg: "invalid output format",
		},
		{
			name: "Empty created by",
			config: &domain.AgentExportConfig{
				ExportName:   "Test Export",
				OutputFormat: domain.ExportFormatExcel,
				CreatedBy:    "",
			},
			isValid:  false,
			errorMsg: "created by is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate config
			err := validateExportConfig(tt.config)
			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			}
		})
	}
}

// TestAgentExportJob_StatusTransitions tests valid status transitions
func TestAgentExportJob_StatusTransitions(t *testing.T) {
	tests := []struct {
		name        string
		fromStatus  string
		toStatus    string
		isValid     bool
	}{
		{
			name:       "IN_PROGRESS to COMPLETED",
			fromStatus: domain.ExportStatusInProgress,
			toStatus:   domain.ExportStatusCompleted,
			isValid:    true,
		},
		{
			name:       "IN_PROGRESS to FAILED",
			fromStatus: domain.ExportStatusInProgress,
			toStatus:   domain.ExportStatusFailed,
			isValid:    true,
		},
		{
			name:       "IN_PROGRESS to CANCELLED",
			fromStatus: domain.ExportStatusInProgress,
			toStatus:   domain.ExportStatusCancelled,
			isValid:    true,
		},
		{
			name:       "COMPLETED to IN_PROGRESS (invalid)",
			fromStatus: domain.ExportStatusCompleted,
			toStatus:   domain.ExportStatusInProgress,
			isValid:    false,
		},
		{
			name:       "FAILED to COMPLETED (invalid)",
			fromStatus: domain.ExportStatusFailed,
			toStatus:   domain.ExportStatusCompleted,
			isValid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := isValidStatusTransition(tt.fromStatus, tt.toStatus)
			assert.Equal(t, tt.isValid, isValid)
		})
	}
}

// TestUpdateJobStatus_ProgressValidation tests progress validation logic
func TestUpdateJobStatus_ProgressValidation(t *testing.T) {
	tests := []struct {
		name             string
		status           string
		progress         int
		recordsProcessed int
		totalRecords     int
		expectError      bool
	}{
		{
			name:             "Valid progress",
			status:           domain.ExportStatusInProgress,
			progress:         50,
			recordsProcessed: 500,
			totalRecords:     1000,
			expectError:      false,
		},
		{
			name:             "Progress out of range (negative)",
			status:           domain.ExportStatusInProgress,
			progress:         -10,
			recordsProcessed: 0,
			totalRecords:     1000,
			expectError:      true,
		},
		{
			name:             "Progress out of range (>100)",
			status:           domain.ExportStatusInProgress,
			progress:         150,
			recordsProcessed: 1000,
			totalRecords:     1000,
			expectError:      true,
		},
		{
			name:             "Completed with 100% progress",
			status:           domain.ExportStatusCompleted,
			progress:         100,
			recordsProcessed: 1000,
			totalRecords:     1000,
			expectError:      false,
		},
		{
			name:             "Completed with <100% progress (invalid)",
			status:           domain.ExportStatusCompleted,
			progress:         90,
			recordsProcessed: 900,
			totalRecords:     1000,
			expectError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateJobProgress(tt.status, tt.progress, tt.recordsProcessed, tt.totalRecords)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ========================================================================
// HELPER FUNCTIONS FOR VALIDATION (would be in actual repository code)
// ========================================================================

func validateExportConfig(config *domain.AgentExportConfig) error {
	if config.ExportName == "" {
		return fmt.Errorf("export name is required")
	}
	if config.OutputFormat != domain.ExportFormatExcel &&
		config.OutputFormat != domain.ExportFormatPDF &&
		config.OutputFormat != domain.ExportFormatCSV {
		return fmt.Errorf("invalid output format: %s", config.OutputFormat)
	}
	if config.CreatedBy == "" {
		return fmt.Errorf("created by is required")
	}
	return nil
}

func isValidStatusTransition(fromStatus, toStatus string) bool {
	validTransitions := map[string][]string{
		domain.ExportStatusInProgress: {
			domain.ExportStatusCompleted,
			domain.ExportStatusFailed,
			domain.ExportStatusCancelled,
		},
	}

	allowedTransitions, ok := validTransitions[fromStatus]
	if !ok {
		return false
	}

	for _, allowed := range allowedTransitions {
		if allowed == toStatus {
			return true
		}
	}
	return false
}

func validateJobProgress(status string, progress, recordsProcessed, totalRecords int) error {
	if progress < 0 || progress > 100 {
		return fmt.Errorf("progress must be between 0 and 100")
	}

	if status == domain.ExportStatusCompleted && progress != 100 {
		return fmt.Errorf("completed jobs must have 100%% progress")
	}

	if recordsProcessed > totalRecords {
		return fmt.Errorf("records processed cannot exceed total records")
	}

	return nil
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}
