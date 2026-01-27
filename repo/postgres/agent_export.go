package repo

import (
	"context"
	"encoding/json"

	sq "github.com/Masterminds/squirrel"
	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"

	"pli-agent-api/core/domain"
)

// AgentExportRepository handles all database operations for agent exports
// Phase 10: Batch & Webhook APIs
// AGT-064 to AGT-067: Export Configuration and Execution
// WF-AGT-PRF-012: Profile Export Workflow
type AgentExportRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewAgentExportRepository creates a new agent export repository
func NewAgentExportRepository(db *dblib.DB, cfg *config.Config) *AgentExportRepository {
	return &AgentExportRepository{
		db:  db,
		cfg: cfg,
	}
}

const agentExportConfigTable = "agent_export_configs"
const agentExportJobTable = "agent_export_jobs"

// CreateConfig creates a new export configuration
// AGT-064: Configure Export Parameters
func (r *AgentExportRepository) CreateConfig(ctx context.Context, config *domain.AgentExportConfig) (*domain.AgentExportConfig, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Insert(agentExportConfigTable).
		Columns(
			"export_name", "filters", "fields", "output_format",
			"estimated_records", "created_by",
		).
		Values(
			config.ExportName, config.Filters, config.Fields,
			config.OutputFormat, config.EstimatedRecords, config.CreatedBy,
		).
		Suffix("RETURNING *")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var result domain.AgentExportConfig
	err = r.db.Get(cCtx, &result, sql, args...)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// CreateJob creates a new export job
// AGT-065: Execute Export Asynchronously
func (r *AgentExportRepository) CreateJob(ctx context.Context, job *domain.AgentExportJob) (*domain.AgentExportJob, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Insert(agentExportJobTable).
		Columns(
			"export_config_id", "requested_by", "status", "workflow_id",
			"total_records", "metadata",
		).
		Values(
			job.ExportConfigID, job.RequestedBy, job.Status,
			job.WorkflowID, job.TotalRecords, job.Metadata,
		).
		Suffix("RETURNING *")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var result domain.AgentExportJob
	err = r.db.Get(cCtx, &result, sql, args...)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateJobStatus updates export job status with progress
// Used by Profile Export Workflow activities
func (r *AgentExportRepository) UpdateJobStatus(
	ctx context.Context,
	exportID string,
	status string,
	progress int,
	recordsProcessed int,
	fileURL, errorMsg *string,
) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Update(agentExportJobTable).
		Set("status", status).
		Set("progress_percentage", progress).
		Set("records_processed", recordsProcessed)

	if fileURL != nil {
		query = query.Set("file_url", *fileURL)
	}
	if errorMsg != nil {
		query = query.Set("error_message", *errorMsg)
	}
	if status == domain.ExportStatusCompleted || status == domain.ExportStatusFailed {
		query = query.Set("completed_at", sq.Expr("NOW()"))
	}

	query = query.Where(sq.Eq{"export_id": exportID})

	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(cCtx, sql, args...)
	return err
}

// GetJobByID retrieves an export job by ID
// AGT-066: Get Export Status
func (r *AgentExportRepository) GetJobByID(ctx context.Context, exportID string) (*domain.AgentExportJob, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentExportJobTable).
		Where(sq.Eq{"export_id": exportID})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var job domain.AgentExportJob
	err = r.db.Get(cCtx, &job, sql, args...)
	if err != nil {
		return nil, err
	}

	return &job, nil
}

// GetConfigByID retrieves an export configuration by ID
func (r *AgentExportRepository) GetConfigByID(ctx context.Context, configID string) (*domain.AgentExportConfig, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentExportConfigTable).
		Where(sq.Eq{"export_config_id": configID})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var config domain.AgentExportConfig
	err = r.db.Get(cCtx, &config, sql, args...)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// GetJobsByUser retrieves export jobs for a user
func (r *AgentExportRepository) GetJobsByUser(ctx context.Context, requestedBy string, limit int) ([]domain.AgentExportJob, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentExportJobTable).
		Where(sq.Eq{"requested_by": requestedBy}).
		OrderBy("started_at DESC").
		Limit(uint64(limit))

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var jobs []domain.AgentExportJob
	err = r.db.Select(cCtx, &jobs, sql, args...)
	if err != nil {
		return nil, err
	}

	return jobs, nil
}

// ExportDataQuery builds and executes query for export data
// Used by Profile Export Workflow to fetch agent data
func (r *AgentExportRepository) ExportDataQuery(
	ctx context.Context,
	filters domain.ExportFilters,
) ([]domain.AgentProfile, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutHigh"))
	defer cancel()

	// Build query with filters
	query := dblib.Psql.Select("*").
		From("agent_profiles").
		Where(sq.Eq{"deleted_at": nil})

	if filters.Status != nil {
		query = query.Where(sq.Eq{"status": *filters.Status})
	}
	if filters.OfficeCode != nil {
		query = query.Where(sq.Eq{"office_code": *filters.OfficeCode})
	}
	if filters.AgentType != nil {
		query = query.Where(sq.Eq{"agent_type": *filters.AgentType})
	}
	if filters.FromDate != nil {
		query = query.Where(sq.GtOrEq{"created_at": *filters.FromDate})
	}
	if filters.ToDate != nil {
		query = query.Where(sq.LtOrEq{"created_at": *filters.ToDate})
	}

	query = query.OrderBy("created_at DESC")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var profiles []domain.AgentProfile
	err = r.db.Select(cCtx, &profiles, sql, args...)
	if err != nil {
		return nil, err
	}

	return profiles, nil
}

// EstimateRecordCount estimates number of records for export
// Used during export configuration
func (r *AgentExportRepository) EstimateRecordCount(ctx context.Context, filtersJSON string) (int, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Parse filters
	var filters domain.ExportFilters
	if filtersJSON != "" {
		err := json.Unmarshal([]byte(filtersJSON), &filters)
		if err != nil {
			return 0, err
		}
	}

	// Build count query
	query := dblib.Psql.Select("COUNT(*)").
		From("agent_profiles").
		Where(sq.Eq{"deleted_at": nil})

	if filters.Status != nil {
		query = query.Where(sq.Eq{"status": *filters.Status})
	}
	if filters.OfficeCode != nil {
		query = query.Where(sq.Eq{"office_code": *filters.OfficeCode})
	}
	if filters.AgentType != nil {
		query = query.Where(sq.Eq{"agent_type": *filters.AgentType})
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return 0, err
	}

	var count int
	err = r.db.Get(cCtx, &count, sql, args...)
	if err != nil {
		return 0, err
	}

	return count, nil
}
