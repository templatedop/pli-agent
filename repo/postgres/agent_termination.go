package repo

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"

	"pli-agent-api/core/domain"
	dbutil "pli-agent-api/db"
)

// AgentTerminationRepository handles termination and reinstatement operations
// AGT-039 to AGT-041: Status Management APIs
// BR-AGT-PRF-016, BR-AGT-PRF-017: Termination workflow
type AgentTerminationRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewAgentTerminationRepository creates a new termination repository
func NewAgentTerminationRepository(db *dblib.DB, cfg *config.Config) *AgentTerminationRepository {
	return &AgentTerminationRepository{
		db:  db,
		cfg: cfg,
	}
}

const terminationRecordsTable = "agent_termination_records"
const reinstatementRequestsTable = "agent_reinstatement_requests"
const dataArchivesTable = "agent_data_archives"

// TerminateAgent terminates an agent and creates termination record in SINGLE database hit
// AGT-039: Terminate Agent
// BR-AGT-PRF-017: Agent Termination Workflow
// OPTIMIZED: Uses CTE to update agent_profiles + create termination_record + audit log atomically
func (r *AgentTerminationRepository) TerminateAgent(
	ctx context.Context,
	agentID string,
	terminationReason string,
	terminationReasonCode string,
	effectiveDate time.Time,
	terminatedBy string,
	workflowID string,
) (*domain.AgentTerminationRecord, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutHigh"))
	defer cancel()

	// SINGLE database hit: Update agent_profiles + INSERT termination_record + audit log
	// Uses CTE pattern for atomicity
	batch := &pgx.Batch{}

	sql := `
		WITH profile_update AS (
			UPDATE agent_profiles
			SET
				status = $2,
				status_date = $3,
				status_reason = $4,
				termination_date = $5,
				termination_reason = $6,
				termination_reason_code = $7,
				terminated_by = $8,
				commission_enabled = false,
				updated_at = $9,
				updated_by = $10,
				version = version + 1
			WHERE agent_id = $1 AND deleted_at IS NULL
			RETURNING agent_id
		),
		term_insert AS (
			INSERT INTO agent_termination_records (
				agent_id, termination_date, effective_date, termination_reason,
				termination_reason_code, terminated_by, workflow_id, workflow_status,
				status_updated
			)
			SELECT $1, $11, $5, $6, $7, $8, $12, $13, true
			WHERE EXISTS (SELECT 1 FROM profile_update)
			RETURNING *
		),
		audit_insert AS (
			INSERT INTO agent_audit_logs (
				agent_id, action_type, field_name, old_value, new_value, action_reason, performed_by, performed_at
			)
			SELECT agent_id, $14, $15, $16, $2, $6, $8, $9
			FROM profile_update
		)
		SELECT * FROM term_insert
	`

	args := []interface{}{
		agentID,                                // $1
		domain.AgentStatusTerminated,           // $2
		effectiveDate,                          // $3
		terminationReason,                      // $4
		effectiveDate,                          // $5
		terminationReason,                      // $6
		terminationReasonCode,                  // $7
		terminatedBy,                           // $8
		time.Now(),                             // $9
		terminatedBy,                           // $10
		time.Now(),                             // $11 - termination_date
		workflowID,                             // $12
		domain.WorkflowStatusInProgress,        // $13
		domain.AuditActionStatusChange,         // $14
		"status",                               // $15
		"", // old_value placeholder            // $16
	}

	var termRecord domain.AgentTerminationRecord
	err := dbutil.QueueReturnRowRaw(batch, sql, args, pgx.RowToStructByNameLax[domain.AgentTerminationRecord], &termRecord)
	if err != nil {
		return nil, err
	}

	err = r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, err
	}

	return &termRecord, nil
}

// UpdateTerminationRecord updates termination record actions
// Used by workflow activities to track progress
func (r *AgentTerminationRepository) UpdateTerminationRecord(
	ctx context.Context,
	terminationID string,
	updates map[string]interface{},
) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Build dynamic UPDATE
	updateQuery := dblib.Psql.Update(terminationRecordsTable).
		Set("updated_at", time.Now()).
		Set("version", sq.Expr("version + 1")).
		Where(sq.Eq{"termination_id": terminationID})

	for field, value := range updates {
		updateQuery = updateQuery.Set(field, value)
	}

	sql, args, err := updateQuery.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build update query: %w", err)
	}

	_, err = r.db.Exec(cCtx, sql, args...)
	return err
}

// FindTerminationRecord retrieves termination record by agent ID
// AGT-040: Get Termination Letter
func (r *AgentTerminationRepository) FindTerminationRecord(ctx context.Context, agentID string) (*domain.AgentTerminationRecord, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(terminationRecordsTable).
		Where(sq.Eq{"agent_id": agentID}).
		OrderBy("termination_date DESC").
		Limit(1)

	var record domain.AgentTerminationRecord
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentTerminationRecord], &record)
	if err != nil {
		return nil, err
	}

	return &record, nil
}

// CreateReinstatementRequest creates a reinstatement request
// AGT-041: Reinstate Agent
// OPTIMIZED: Single INSERT with audit log via CTE
func (r *AgentTerminationRepository) CreateReinstatementRequest(
	ctx context.Context,
	agentID string,
	reinstatementReason string,
	requestedBy string,
	workflowID string,
) (*domain.AgentReinstatementRequest, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// SINGLE database hit with audit log
	batch := &pgx.Batch{}

	sql := `
		WITH request_insert AS (
			INSERT INTO agent_reinstatement_requests (
				agent_id, reinstatement_reason, requested_by, workflow_id, status
			)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING *
		),
		audit_insert AS (
			INSERT INTO agent_audit_logs (
				agent_id, action_type, field_name, new_value, action_reason, performed_by, performed_at
			)
			SELECT agent_id, $6, $7, $5, $2, $3, $8
			FROM request_insert
		)
		SELECT * FROM request_insert
	`

	args := []interface{}{
		agentID,
		reinstatementReason,
		requestedBy,
		workflowID,
		domain.ReinstatementStatusPending,
		domain.AuditActionStatusChange,
		"reinstatement_status",
		time.Now(),
	}

	var request domain.AgentReinstatementRequest
	err := dbutil.QueueReturnRowRaw(batch, sql, args, pgx.RowToStructByNameLax[domain.AgentReinstatementRequest], &request)
	if err != nil {
		return nil, err
	}

	err = r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, err
	}

	return &request, nil
}

// ApproveReinstatement approves reinstatement and updates agent status in SINGLE database hit
// OPTIMIZED: Uses CTE to update reinstatement_request + agent_profiles + audit log atomically
func (r *AgentTerminationRepository) ApproveReinstatement(
	ctx context.Context,
	reinstatementID string,
	approvedBy string,
	conditions string,
	probationDays int,
) (*domain.AgentReinstatementRequest, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutHigh"))
	defer cancel()

	// SINGLE database hit: Update reinstatement_request + agent_profiles + audit log
	batch := &pgx.Batch{}

	sql := `
		WITH reinst_update AS (
			UPDATE agent_reinstatement_requests
			SET
				status = $2,
				approved_by = $3,
				approved_at = $4,
				reinstatement_conditions = $5,
				probation_period_days = $6,
				updated_at = $4,
				version = version + 1
			WHERE reinstatement_id = $1
			RETURNING *
		),
		profile_update AS (
			UPDATE agent_profiles
			SET
				status = $7,
				status_date = $4,
				status_reason = $8,
				reinstatement_date = $4,
				reinstated_by = $3,
				reinstatement_reason = (SELECT reinstatement_reason FROM reinst_update),
				commission_enabled = true,
				termination_date = NULL,
				termination_reason = NULL,
				updated_at = $4,
				updated_by = $3,
				version = version + 1
			WHERE agent_id = (SELECT agent_id FROM reinst_update)
		),
		audit_insert AS (
			INSERT INTO agent_audit_logs (
				agent_id, action_type, field_name, old_value, new_value, action_reason, performed_by, performed_at
			)
			SELECT agent_id, $9, $10, $11, $7, $8, $3, $4
			FROM reinst_update
		)
		SELECT * FROM reinst_update
	`

	args := []interface{}{
		reinstatementID,
		domain.ReinstatementStatusApproved,
		approvedBy,
		time.Now(),
		conditions,
		probationDays,
		domain.AgentStatusActive,
		"Reinstated after approval",
		domain.AuditActionStatusChange,
		"status",
		domain.AgentStatusTerminated,
	}

	var request domain.AgentReinstatementRequest
	err := dbutil.QueueReturnRowRaw(batch, sql, args, pgx.RowToStructByNameLax[domain.AgentReinstatementRequest], &request)
	if err != nil {
		return nil, err
	}

	err = r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, err
	}

	return &request, nil
}

// FindReinstatementRequest retrieves reinstatement request
func (r *AgentTerminationRepository) FindReinstatementRequest(ctx context.Context, reinstatementID string) (*domain.AgentReinstatementRequest, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(reinstatementRequestsTable).
		Where(sq.Eq{"reinstatement_id": reinstatementID, "deleted_at": nil})

	var request domain.AgentReinstatementRequest
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentReinstatementRequest], &request)
	if err != nil {
		return nil, err
	}

	return &request, nil
}

// CreateDataArchive creates data archive record
// BR-AGT-PRF-017: 7-year retention
func (r *AgentTerminationRepository) CreateDataArchive(
	ctx context.Context,
	agentID string,
	archiveType string,
	dataSnapshot string,
	archivedBy string,
) (*domain.AgentDataArchive, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// 7-year retention
	retentionUntil := time.Now().AddDate(7, 0, 0)

	query := dblib.Psql.Insert(dataArchivesTable).
		Columns(
			"agent_id", "archive_type", "retention_until",
			"data_snapshot", "archived_by",
		).
		Values(agentID, archiveType, retentionUntil, dataSnapshot, archivedBy).
		Suffix("RETURNING *")

	var archive domain.AgentDataArchive
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentDataArchive], &archive)
	if err != nil {
		return nil, err
	}

	return &archive, nil
}

// FindPendingReinstatementRequests retrieves all pending reinstatement requests
func (r *AgentTerminationRepository) FindPendingReinstatementRequests(ctx context.Context) ([]domain.AgentReinstatementRequest, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(reinstatementRequestsTable).
		Where(sq.Eq{"status": domain.ReinstatementStatusPending, "deleted_at": nil}).
		OrderBy("request_date ASC")

	var requests []domain.AgentReinstatementRequest
	err := dblib.SelectRows(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentReinstatementRequest], &requests)
	if err != nil {
		return nil, err
	}

	return requests, nil
}
