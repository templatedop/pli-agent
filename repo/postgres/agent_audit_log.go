package repo

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"

	"pli-agent-api/core/domain"
	dbutil "pli-agent-api/db"
)

// AgentAuditLogRepository handles all database operations for agent audit logs
// E-08: Agent Audit Log Entity
// BR-AGT-PRF-005: Audit Logging
// BR-AGT-PRF-006: Name Update with Audit Logging
type AgentAuditLogRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewAgentAuditLogRepository creates a new agent audit log repository
func NewAgentAuditLogRepository(db *dblib.DB, cfg *config.Config) *AgentAuditLogRepository {
	return &AgentAuditLogRepository{
		db:  db,
		cfg: cfg,
	}
}

const agentAuditLogTable = "agent_audit_logs"

// Create inserts a new audit log entry
// BR-AGT-PRF-005: Audit Logging
// BR-AGT-PRF-006: Name Update with Audit Logging
// Note: Audit logs are INSERT-ONLY, no updates or deletes allowed
func (r *AgentAuditLogRepository) Create(ctx context.Context, auditLog domain.AgentAuditLog) (*domain.AgentAuditLog, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Insert audit log
	// BR-AGT-PRF-005: All agent actions must be logged with timestamp and user
	query := dblib.Psql.Insert(agentAuditLogTable).
		Columns(
			"agent_id", "action_type", "field_name", "old_value", "new_value",
			"action_reason", "performed_by", "performed_at", "ip_address", "metadata",
		).
		Values(
			auditLog.AgentID, auditLog.ActionType, auditLog.FieldName, auditLog.OldValue,
			auditLog.NewValue, auditLog.ActionReason, auditLog.PerformedBy,
			auditLog.PerformedAt, auditLog.IPAddress, auditLog.Metadata,
		).
		Suffix("RETURNING audit_id, created_at")

	var result domain.AgentAuditLog
	err := dblib.InsertReturning(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentAuditLog])
	if err != nil {
		return nil, err
	}

	// Copy input data to result
	result = auditLog
	return &result, nil
}

// FindByID retrieves an audit log by ID
func (r *AgentAuditLogRepository) FindByID(ctx context.Context, auditID string) (*domain.AgentAuditLog, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentAuditLogTable).
		Where(sq.Eq{"audit_id": auditID})

	var auditLog domain.AgentAuditLog
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentAuditLog], &auditLog)
	if err != nil {
		return nil, err
	}

	return &auditLog, nil
}

// FindByAgentID retrieves all audit logs for an agent
// BR-AGT-PRF-005: Audit Logging - Complete audit trail
func (r *AgentAuditLogRepository) FindByAgentID(ctx context.Context, agentID string) ([]domain.AgentAuditLog, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentAuditLogTable).
		Where(sq.Eq{"agent_id": agentID}).
		OrderBy("performed_at DESC")

	var auditLogs []domain.AgentAuditLog
	err := dblib.SelectRows(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentAuditLog], &auditLogs)
	if err != nil {
		return nil, err
	}

	return auditLogs, nil
}

// AuditLogFilters defines filters for audit log queries
type AuditLogFilters struct {
	AgentID     string
	ActionType  string
	FieldName   string
	PerformedBy string
	FromDate    *time.Time
	ToDate      *time.Time
}

// FindWithFilters retrieves audit logs with filters and pagination
// BR-AGT-PRF-005: Audit Logging - Queryable audit trail
// BR-AGT-PRF-006: Name Update with Audit Logging
func (r *AgentAuditLogRepository) FindWithFilters(ctx context.Context, filters AuditLogFilters, skip, limit uint64, orderBy, sortType string) ([]domain.AgentAuditLog, int64, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	// Use batch to get count and data in single round trip
	// OPTIMIZATION: Batch combines COUNT + SELECT queries
	batch := &pgx.Batch{}

	// Base query with filters
	baseQuery := sq.Select().From(agentAuditLogTable)

	// Apply filters
	if filters.AgentID != "" {
		baseQuery = baseQuery.Where(sq.Eq{"agent_id": filters.AgentID})
	}
	if filters.ActionType != "" {
		baseQuery = baseQuery.Where(sq.Eq{"action_type": filters.ActionType})
	}
	if filters.FieldName != "" {
		baseQuery = baseQuery.Where(sq.Eq{"field_name": filters.FieldName})
	}
	if filters.PerformedBy != "" {
		baseQuery = baseQuery.Where(sq.Eq{"performed_by": filters.PerformedBy})
	}
	if filters.FromDate != nil {
		baseQuery = baseQuery.Where(sq.GtOrEq{"performed_at": *filters.FromDate})
	}
	if filters.ToDate != nil {
		baseQuery = baseQuery.Where(sq.LtOrEq{"performed_at": *filters.ToDate})
	}

	// Query 1: Count total records
	countQuery := baseQuery.Columns("COUNT(*)")
	var totalCount int64
	err := dblib.QueueReturnRow(batch, countQuery, pgx.RowTo[int64], &totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Query 2: Get paginated data
	dataQuery := baseQuery.
		Columns("*").
		OrderBy(orderBy + " " + sortType).
		Limit(limit).
		Offset(skip)

	var auditLogs []domain.AgentAuditLog
	err = dblib.QueueReturn(batch, dataQuery, pgx.RowToStructByNameLax[domain.AgentAuditLog], &auditLogs)
	if err != nil {
		return nil, 0, err
	}

	// Execute batch
	err = r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, 0, err
	}

	return auditLogs, totalCount, nil
}

// FindByActionType retrieves audit logs by action type
// BR-AGT-PRF-005: Audit Logging - Filter by action type
func (r *AgentAuditLogRepository) FindByActionType(ctx context.Context, actionType string) ([]domain.AgentAuditLog, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentAuditLogTable).
		Where(sq.Eq{"action_type": actionType}).
		OrderBy("performed_at DESC")

	var auditLogs []domain.AgentAuditLog
	err := dblib.SelectRows(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentAuditLog], &auditLogs)
	if err != nil {
		return nil, err
	}

	return auditLogs, nil
}

// FindByDateRange retrieves audit logs within a date range
// BR-AGT-PRF-005: Audit Logging - Date range queries
func (r *AgentAuditLogRepository) FindByDateRange(ctx context.Context, fromDate, toDate time.Time) ([]domain.AgentAuditLog, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentAuditLogTable).
		Where(sq.GtOrEq{"performed_at": fromDate}).
		Where(sq.LtOrEq{"performed_at": toDate}).
		OrderBy("performed_at DESC")

	var auditLogs []domain.AgentAuditLog
	err := dblib.SelectRows(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentAuditLog], &auditLogs)
	if err != nil {
		return nil, err
	}

	return auditLogs, nil
}

// FindRecentByAgentID retrieves recent audit logs for an agent (last N entries)
// BR-AGT-PRF-005: Audit Logging - Recent activity view
func (r *AgentAuditLogRepository) FindRecentByAgentID(ctx context.Context, agentID string, limit uint64) ([]domain.AgentAuditLog, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentAuditLogTable).
		Where(sq.Eq{"agent_id": agentID}).
		OrderBy("performed_at DESC").
		Limit(limit)

	var auditLogs []domain.AgentAuditLog
	err := dblib.SelectRows(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentAuditLog], &auditLogs)
	if err != nil {
		return nil, err
	}

	return auditLogs, nil
}

// BatchCreate inserts multiple audit log entries in a single transaction
// OPTIMIZATION: UNNEST pattern for multiple audit log inserts
// BR-AGT-PRF-005: Audit Logging - Bulk audit logging
func (r *AgentAuditLogRepository) BatchCreate(ctx context.Context, auditLogs []domain.AgentAuditLog) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	// Use UNNEST to bulk insert audit logs in single query
	// CRITICAL: Golang variables cannot be passed between batch queries - must use UNNEST pattern
	batch := &pgx.Batch{}

	// Extract arrays for UNNEST
	agentIDs := make([]string, len(auditLogs))
	actionTypes := make([]string, len(auditLogs))
	fieldNames := make([]sql.NullString, len(auditLogs))
	oldValues := make([]sql.NullString, len(auditLogs))
	newValues := make([]sql.NullString, len(auditLogs))
	actionReasons := make([]sql.NullString, len(auditLogs))
	performedBys := make([]string, len(auditLogs))
	performedAts := make([]time.Time, len(auditLogs))
	ipAddresses := make([]sql.NullString, len(auditLogs))
	metadatas := make([]interface{}, len(auditLogs))

	for i, log := range auditLogs {
		agentIDs[i] = log.AgentID
		actionTypes[i] = log.ActionType
		fieldNames[i] = log.FieldName
		oldValues[i] = log.OldValue
		newValues[i] = log.NewValue
		actionReasons[i] = log.ActionReason
		performedBys[i] = log.PerformedBy
		performedAts[i] = log.PerformedAt
		ipAddresses[i] = log.IPAddress
		metadatas[i] = log.Metadata
	}

	sql := `
		INSERT INTO agent_audit_logs (
			agent_id, action_type, field_name, old_value, new_value,
			action_reason, performed_by, performed_at, ip_address, metadata
		)
		SELECT * FROM UNNEST(
			$1::uuid[],
			$2::text[],
			$3::text[],
			$4::text[],
			$5::text[],
			$6::text[],
			$7::text[],
			$8::timestamp[],
			$9::text[],
			$10::jsonb[]
		)
	`

	args := []interface{}{
		agentIDs, actionTypes, fieldNames, oldValues, newValues,
		actionReasons, performedBys, performedAts, ipAddresses, metadatas,
	}

	err := dbutil.QueueExecRowRaw(batch, sql, args...)
	if err != nil {
		return err
	}

	return r.db.SendBatch(cCtx, batch).Close()
}

// GetAuditSummary retrieves audit summary statistics for an agent
// BR-AGT-PRF-005: Audit Logging - Summary statistics
type AuditSummary struct {
	AgentID          string
	TotalActions     int64
	LastActionDate   time.Time
	ActionTypeCounts map[string]int64
}

func (r *AgentAuditLogRepository) GetAuditSummary(ctx context.Context, agentID string) (*AuditSummary, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	// Use batch to get multiple statistics in single round trip
	// OPTIMIZATION: Batch combines multiple aggregation queries
	batch := &pgx.Batch{}

	// Query 1: Total actions count
	var totalActions int64
	countQuery := dblib.Psql.Select("COUNT(*)").
		From(agentAuditLogTable).
		Where(sq.Eq{"agent_id": agentID})

	err := dblib.QueueReturnRow(batch, countQuery, pgx.RowTo[int64], &totalActions)
	if err != nil {
		return nil, err
	}

	// Query 2: Last action date
	var lastActionDate time.Time
	lastActionQuery := dblib.Psql.Select("MAX(performed_at)").
		From(agentAuditLogTable).
		Where(sq.Eq{"agent_id": agentID})

	err = dblib.QueueReturnRow(batch, lastActionQuery, pgx.RowTo[time.Time], &lastActionDate)
	if err != nil {
		return nil, err
	}

	// Execute batch
	err = r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, err
	}

	summary := &AuditSummary{
		AgentID:        agentID,
		TotalActions:   totalActions,
		LastActionDate: lastActionDate,
	}

	return summary, nil
}

// GetActionTypeCount retrieves count of actions grouped by action type
// BR-AGT-PRF-005: Audit Logging - Action type statistics
func (r *AgentAuditLogRepository) GetActionTypeCount(ctx context.Context, agentID string) (map[string]int64, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	query := dblib.Psql.Select("action_type", "COUNT(*) as count").
		From(agentAuditLogTable).
		Where(sq.Eq{"agent_id": agentID}).
		GroupBy("action_type")

	type ActionCount struct {
		ActionType string `db:"action_type"`
		Count      int64  `db:"count"`
	}

	var actionCounts []ActionCount
	err := dblib.SelectRows(cCtx, r.db, query, pgx.RowToStructByNameLax[ActionCount], &actionCounts)
	if err != nil {
		return nil, err
	}

	// Convert to map
	result := make(map[string]int64)
	for _, ac := range actionCounts {
		result[ac.ActionType] = ac.Count
	}

	return result, nil
}

// FindFieldHistory retrieves the complete history of changes for a specific field
// BR-AGT-PRF-005: Audit Logging - Field change history
// BR-AGT-PRF-006: Name Update with Audit Logging - Track name changes
func (r *AgentAuditLogRepository) FindFieldHistory(ctx context.Context, agentID, fieldName string) ([]domain.AgentAuditLog, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentAuditLogTable).
		Where(sq.Eq{"agent_id": agentID, "field_name": fieldName}).
		OrderBy("performed_at DESC")

	var auditLogs []domain.AgentAuditLog
	err := dblib.SelectRows(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentAuditLog], &auditLogs)
	if err != nil {
		return nil, err
	}

	return auditLogs, nil
}

// FindByPerformedBy retrieves audit logs by the user who performed the action
// BR-AGT-PRF-005: Audit Logging - User activity tracking
func (r *AgentAuditLogRepository) FindByPerformedBy(ctx context.Context, performedBy string, limit uint64) ([]domain.AgentAuditLog, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentAuditLogTable).
		Where(sq.Eq{"performed_by": performedBy}).
		OrderBy("performed_at DESC").
		Limit(limit)

	var auditLogs []domain.AgentAuditLog
	err := dblib.SelectRows(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentAuditLog], &auditLogs)
	if err != nil {
		return nil, err
	}

	return auditLogs, nil
}

// GetHistory retrieves audit history for an agent with pagination and date filters
// AGT-028: Get Audit History
// FR-AGT-PRF-022: Profile Change History and Audit Trail
// CRITICAL: Single query with pagination
func (r *AgentAuditLogRepository) GetHistory(
	ctx context.Context,
	agentID string,
	fromDate, toDate *time.Time,
	page, limit int,
) ([]domain.AgentAuditLog, int, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	offset := (page - 1) * limit

	// Build base query
	baseQuery := dblib.Psql.Select("*").
		From(agentAuditLogTable).
		Where(sq.Eq{"agent_id": agentID})

	// Apply date filters
	if fromDate != nil {
		baseQuery = baseQuery.Where(sq.GtOrEq{"performed_at": *fromDate})
	}
	if toDate != nil {
		baseQuery = baseQuery.Where(sq.LtOrEq{"performed_at": *toDate})
	}

	// Use batch for single database round trip
	batch := &pgx.Batch{}

	// Query 1: Count total records
	countQuery := baseQuery
	countSQL, countArgs, _ := countQuery.ToSql()
	countSQL = "SELECT COUNT(*) FROM (" + countSQL + ") AS subquery"

	var totalCount int
	err := dblib.QueueReturnRowRaw(batch, countSQL, countArgs, pgx.RowTo[int], &totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to queue count query: %w", err)
	}

	// Query 2: Get paginated data
	dataQuery := baseQuery.
		OrderBy("performed_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	var auditLogs []domain.AgentAuditLog
	err = dblib.QueueReturn(batch, dataQuery, pgx.RowToStructByNameLax[domain.AgentAuditLog], &auditLogs)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to queue data query: %w", err)
	}

	// Execute batch in single round trip
	err = r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute audit history batch: %w", err)
	}

	return auditLogs, totalCount, nil
}

// GetTimeline retrieves agent activity timeline combining audit logs, license changes, and status changes
// AGT-076: Agent Activity Timeline
// Phase 9: Search & Dashboard APIs
// CRITICAL: Single query using UNION to combine different event sources
func (r *AgentAuditLogRepository) GetTimeline(
	ctx context.Context,
	agentID string,
	activityType *string,
	fromDate, toDate *time.Time,
	page, limit int,
) ([]domain.TimelineEvent, int, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	offset := (page - 1) * limit

	// Build timeline query combining multiple event sources
	// UNION ALL combines audit logs + license changes + status changes
	timelineSQL := `
		WITH timeline_events AS (
			-- Audit log events
			SELECT
				performed_at AS timestamp,
				'PROFILE_CHANGE' AS event_type,
				CASE
					WHEN field_name IS NOT NULL THEN
						CONCAT('Updated ', field_name, ' from "', COALESCE(old_value, 'NULL'), '" to "', COALESCE(new_value, 'NULL'), '"')
					ELSE
						action_type
				END AS description,
				performed_by,
				field_name,
				old_value,
				new_value,
				action_reason
			FROM agent_audit_logs
			WHERE agent_id = $1
				AND ($4::text IS NULL OR 'PROFILE_CHANGE' = $4)
				AND ($5::timestamptz IS NULL OR performed_at >= $5)
				AND ($6::timestamptz IS NULL OR performed_at <= $6)

			UNION ALL

			-- License events
			SELECT
				updated_at AS timestamp,
				'LICENSE_UPDATE' AS event_type,
				CONCAT('License ', license_type, ' updated - Status: ', status) AS description,
				NULL AS performed_by,
				NULL AS field_name,
				NULL AS old_value,
				NULL AS new_value,
				NULL AS action_reason
			FROM agent_licenses
			WHERE agent_id = $1
				AND ($4::text IS NULL OR 'LICENSE_UPDATE' = $4)
				AND ($5::timestamptz IS NULL OR updated_at >= $5)
				AND ($6::timestamptz IS NULL OR updated_at <= $6)

			UNION ALL

			-- Status change events (from audit logs)
			SELECT
				performed_at AS timestamp,
				'STATUS_CHANGE' AS event_type,
				CONCAT('Status changed from ', COALESCE(old_value, 'NULL'), ' to ', COALESCE(new_value, 'NULL')) AS description,
				performed_by,
				NULL AS field_name,
				old_value,
				new_value,
				action_reason
			FROM agent_audit_logs
			WHERE agent_id = $1
				AND action_type = 'STATUS_CHANGE'
				AND ($4::text IS NULL OR 'STATUS_CHANGE' = $4)
				AND ($5::timestamptz IS NULL OR performed_at >= $5)
				AND ($6::timestamptz IS NULL OR performed_at <= $6)
		)
		SELECT * FROM timeline_events
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3
	`

	// Count query for total results
	countSQL := `
		WITH timeline_events AS (
			-- Same UNION logic for counting
			SELECT performed_at AS timestamp
			FROM agent_audit_logs
			WHERE agent_id = $1
				AND ($2::text IS NULL OR 'PROFILE_CHANGE' = $2)
				AND ($3::timestamptz IS NULL OR performed_at >= $3)
				AND ($4::timestamptz IS NULL OR performed_at <= $4)

			UNION ALL

			SELECT updated_at AS timestamp
			FROM agent_licenses
			WHERE agent_id = $1
				AND ($2::text IS NULL OR 'LICENSE_UPDATE' = $2)
				AND ($3::timestamptz IS NULL OR updated_at >= $3)
				AND ($4::timestamptz IS NULL OR updated_at <= $4)

			UNION ALL

			SELECT performed_at AS timestamp
			FROM agent_audit_logs
			WHERE agent_id = $1
				AND action_type = 'STATUS_CHANGE'
				AND ($2::text IS NULL OR 'STATUS_CHANGE' = $2)
				AND ($3::timestamptz IS NULL OR performed_at >= $3)
				AND ($4::timestamptz IS NULL OR performed_at <= $4)
		)
		SELECT COUNT(*) FROM timeline_events
	`

	// Use batch for single database round trip
	batch := &pgx.Batch{}

	// Query 1: Count total events
	var totalCount int
	err := dbutil.QueueReturnRowRaw(batch, countSQL, []interface{}{agentID, activityType, fromDate, toDate}, pgx.RowTo[int], &totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to queue count query: %w", err)
	}

	// Query 2: Get paginated timeline events
	var events []domain.TimelineEvent
	err = dbutil.QueueReturnRowsRaw(batch, timelineSQL, []interface{}{agentID, limit, offset, activityType, fromDate, toDate}, pgx.RowToStructByNameLax[domain.TimelineEvent], &events)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to queue timeline query: %w", err)
	}

	// Execute batch in single round trip
	err = r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute timeline batch: %w", err)
	}

	return events, totalCount, nil
}
