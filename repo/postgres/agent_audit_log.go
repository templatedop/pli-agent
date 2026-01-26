package repo

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"

	"pli-agent-api/core/domain"
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
	AgentID      string
	ActionType   string
	FieldName    string
	PerformedBy  string
	FromDate     *time.Time
	ToDate       *time.Time
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
// OPTIMIZATION: Batch operation for multiple audit log inserts
// BR-AGT-PRF-005: Audit Logging - Bulk audit logging
func (r *AgentAuditLogRepository) BatchCreate(ctx context.Context, auditLogs []domain.AgentAuditLog) ([]domain.AgentAuditLog, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	// Use batch for multiple audit log inserts
	// OPTIMIZATION: Batch operation combines multiple INSERTs in single round-trip
	batch := &pgx.Batch{}
	results := make([]domain.AgentAuditLog, len(auditLogs))

	for i, auditLog := range auditLogs {
		// Insert audit log
		insertQuery := dblib.Psql.Insert(agentAuditLogTable).
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

		err := dblib.QueueReturnRow(batch, insertQuery, pgx.RowToStructByNameLax[domain.AgentAuditLog], &results[i])
		if err != nil {
			return nil, err
		}
	}

	// Execute batch
	err := r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, err
	}

	// Copy input data to results
	for i := range auditLogs {
		results[i] = auditLogs[i]
	}

	return results, nil
}

// GetAuditSummary retrieves audit summary statistics for an agent
// BR-AGT-PRF-005: Audit Logging - Summary statistics
type AuditSummary struct {
	AgentID            string
	TotalActions       int64
	LastActionDate     time.Time
	ActionTypeCounts   map[string]int64
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
