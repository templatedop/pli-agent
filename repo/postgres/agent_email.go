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

// AgentEmailRepository handles all database operations for agent emails
// E-04: Agent Email Entity
// BR-AGT-PRF-011: Email Address Management
type AgentEmailRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewAgentEmailRepository creates a new agent email repository
func NewAgentEmailRepository(db *dblib.DB, cfg *config.Config) *AgentEmailRepository {
	return &AgentEmailRepository{
		db:  db,
		cfg: cfg,
	}
}

const agentEmailTable = "agent_emails"

// Create inserts a new agent email
// FR-AGT-PRF-011: Email Management
// VR-AGT-PRF-014: Email Format Validation
func (r *AgentEmailRepository) Create(ctx context.Context, email domain.AgentEmail) (*domain.AgentEmail, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use CTE to combine INSERT + INSERT audit in single query
	// CRITICAL: Golang variables cannot be passed between batch queries - must combine at SQL level
	// BR-AGT-PRF-011: Email Address Categories (OFFICIAL, PERMANENT, COMMUNICATION)
	// BR-AGT-PRF-005: Audit Logging
	batch := &pgx.Batch{}

	sql := `
		WITH inserted AS (
			INSERT INTO agent_emails (
				agent_id, email_type, email_address, is_primary,
				effective_from, metadata, created_by
			) VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING *
		)
		INSERT INTO agent_audit_logs (agent_id, action_type, field_name, new_value, action_reason, performed_by, performed_at)
		SELECT agent_id, $8, $9, $10, $11, $12, $13
		FROM inserted
		RETURNING (SELECT ROW(email_id, agent_id, email_type, email_address, is_primary,
			effective_from, metadata, created_at, updated_at, created_by, updated_by, deleted_at, version) FROM inserted)
	`

	args := []interface{}{
		email.AgentID, email.EmailType, email.EmailAddress, email.IsPrimary,
		email.EffectiveFrom, email.Metadata, email.CreatedBy,
		domain.AuditActionEmailUpdate, "email_type", email.EmailType, "New email added", email.CreatedBy, time.Now(),
	}

	var result domain.AgentEmail
	err := dbutil.QueueReturnRowRaw(batch, sql, args, pgx.RowToStructByNameLax[domain.AgentEmail], &result)
	if err != nil {
		return nil, err
	}

	err = r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// FindByID retrieves an email by ID
func (r *AgentEmailRepository) FindByID(ctx context.Context, emailID string) (*domain.AgentEmail, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentEmailTable).
		Where(sq.Eq{"email_id": emailID, "deleted_at": nil})

	var email domain.AgentEmail
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentEmail], &email)
	if err != nil {
		return nil, err
	}

	return &email, nil
}

// FindByAgentID retrieves all emails for an agent
// FR-AGT-PRF-011: Email Management
// BR-AGT-PRF-011: Email Address Categories
func (r *AgentEmailRepository) FindByAgentID(ctx context.Context, agentID string) ([]domain.AgentEmail, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentEmailTable).
		Where(sq.Eq{"agent_id": agentID, "deleted_at": nil}).
		OrderBy("is_primary DESC, effective_from DESC")

	var emails []domain.AgentEmail
	err := dblib.SelectRows(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentEmail], &emails)
	if err != nil {
		return nil, err
	}

	return emails, nil
}

// FindByAgentIDAndType retrieves a specific email type for an agent
// BR-AGT-PRF-011: Email Address Categories
func (r *AgentEmailRepository) FindByAgentIDAndType(ctx context.Context, agentID, emailType string) (*domain.AgentEmail, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentEmailTable).
		Where(sq.Eq{"agent_id": agentID, "email_type": emailType, "deleted_at": nil}).
		OrderBy("is_primary DESC, effective_from DESC").
		Limit(1)

	var email domain.AgentEmail
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentEmail], &email)
	if err != nil {
		return nil, err
	}

	return &email, nil
}

// FindPrimaryEmail retrieves the primary email for an agent
func (r *AgentEmailRepository) FindPrimaryEmail(ctx context.Context, agentID string) (*domain.AgentEmail, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentEmailTable).
		Where(sq.Eq{"agent_id": agentID, "is_primary": true, "deleted_at": nil}).
		OrderBy("effective_from DESC").
		Limit(1)

	var email domain.AgentEmail
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentEmail], &email)
	if err != nil {
		return nil, err
	}

	return &email, nil
}

// FindByEmailAddress retrieves an email by email address
// VR-AGT-PRF-014: Email Format Validation - Check uniqueness
func (r *AgentEmailRepository) FindByEmailAddress(ctx context.Context, emailAddress string) (*domain.AgentEmail, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentEmailTable).
		Where(sq.Eq{"email_address": emailAddress, "deleted_at": nil}).
		Limit(1)

	var email domain.AgentEmail
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentEmail], &email)
	if err != nil {
		return nil, err
	}

	return &email, nil
}

// Update updates an agent email
// FR-AGT-PRF-011: Email Management
// VR-AGT-PRF-014: Email Format Validation
func (r *AgentEmailRepository) Update(ctx context.Context, emailID string, updates map[string]interface{}, updatedBy string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use CTE to combine UPDATE + INSERT audit logs in single query
	// CRITICAL: Golang variables cannot be passed between batch queries - must combine at SQL level
	// BR-AGT-PRF-005: Audit Logging
	batch := &pgx.Batch{}

	// Build SET clause dynamically
	setClauses := "updated_at = $2, updated_by = $3"
	args := []interface{}{emailID, time.Now(), updatedBy}
	argIndex := 4

	for field, value := range updates {
		setClauses += fmt.Sprintf(", %s = $%d", field, argIndex)
		args = append(args, value)
		argIndex++
	}

	// Build audit log values for UNNEST
	fieldNames := []string{}
	newValues := []interface{}{}
	for field, value := range updates {
		fieldNames = append(fieldNames, field)
		newValues = append(newValues, value)
	}

	sql := fmt.Sprintf(`
		WITH updated AS (
			UPDATE agent_emails
			SET %s
			WHERE email_id = $1 AND deleted_at IS NULL
			RETURNING agent_id
		)
		INSERT INTO agent_audit_logs (agent_id, action_type, field_name, new_value, performed_by, performed_at)
		SELECT agent_id, $%d, unnest($%d::text[]), unnest($%d::text[]), $%d, $%d
		FROM updated
	`, setClauses, argIndex, argIndex+1, argIndex+2, argIndex+3, argIndex+4)

	args = append(args, domain.AuditActionEmailUpdate, fieldNames, newValues, updatedBy, time.Now())

	err := dbutil.QueueExecRowRaw(batch, sql, args...)
	if err != nil {
		return err
	}

	// Execute batch
	return r.db.SendBatch(cCtx, batch).Close()
}

// Delete soft deletes an agent email
func (r *AgentEmailRepository) Delete(ctx context.Context, emailID, deletedBy string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use CTE to combine UPDATE + INSERT audit in single query
	// CRITICAL: Golang variables cannot be passed between batch queries - must combine at SQL level
	batch := &pgx.Batch{}

	sql := `
		WITH updated AS (
			UPDATE agent_emails
			SET deleted_at = $2, updated_by = $3
			WHERE email_id = $1 AND deleted_at IS NULL
			RETURNING agent_id
		)
		INSERT INTO agent_audit_logs (agent_id, action_type, field_name, action_reason, performed_by, performed_at)
		SELECT agent_id, $4, $5, $6, $7, $8
		FROM updated
	`

	args := []interface{}{
		emailID, time.Now(), deletedBy,
		domain.AuditActionDelete, "email", "Email deleted", deletedBy, time.Now(),
	}

	err := dbutil.QueueExecRowRaw(batch, sql, args...)
	if err != nil {
		return err
	}

	// Execute batch
	return r.db.SendBatch(cCtx, batch).Close()
}

// BatchCreate inserts multiple agent emails in a single transaction
// OPTIMIZATION: Batch operation for multiple email inserts using UNNEST
// FR-AGT-PRF-011: Email Management
func (r *AgentEmailRepository) BatchCreate(ctx context.Context, emails []domain.AgentEmail) ([]domain.AgentEmail, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	// Use UNNEST to bulk insert emails with audit logs in single query
	// CRITICAL: Golang variables cannot be passed between batch queries - must use UNNEST pattern
	// BR-AGT-PRF-005: Audit Logging
	batch := &pgx.Batch{}

	// Prepare arrays for UNNEST
	agentIDs := make([]string, len(emails))
	emailTypes := make([]string, len(emails))
	emailAddresses := make([]string, len(emails))
	isPrimaries := make([]bool, len(emails))
	effectiveFroms := make([]time.Time, len(emails))
	metadatas := make([]interface{}, len(emails))
	createdBys := make([]string, len(emails))

	for i, email := range emails {
		agentIDs[i] = email.AgentID
		emailTypes[i] = email.EmailType
		emailAddresses[i] = email.EmailAddress
		isPrimaries[i] = email.IsPrimary
		effectiveFroms[i] = email.EffectiveFrom
		metadatas[i] = email.Metadata
		createdBys[i] = email.CreatedBy
	}

	sql := `
		WITH inserted AS (
			INSERT INTO agent_emails (agent_id, email_type, email_address, is_primary, effective_from, metadata, created_by)
			SELECT * FROM UNNEST(
				$1::uuid[],
				$2::text[],
				$3::text[],
				$4::boolean[],
				$5::timestamp[],
				$6::jsonb[],
				$7::text[]
			)
			RETURNING *
		)
		INSERT INTO agent_audit_logs (agent_id, action_type, field_name, new_value, action_reason, performed_by, performed_at)
		SELECT agent_id, $8, $9, email_type, $10, created_by, NOW()
		FROM inserted
		RETURNING (SELECT array_agg(ROW(email_id, agent_id, email_type, email_address, is_primary,
			effective_from, metadata, created_at, created_by, updated_at, updated_by, deleted_at, version)::agent_emails) FROM inserted)
	`

	args := []interface{}{
		agentIDs, emailTypes, emailAddresses, isPrimaries, effectiveFroms, metadatas, createdBys,
		domain.AuditActionEmailUpdate, "email_type", "New email added",
	}

	var results []domain.AgentEmail
	err := dbutil.QueueReturnRaw(batch, sql, args, pgx.RowToStructByNameLax[domain.AgentEmail], &results)
	if err != nil {
		return nil, err
	}

	// Execute batch
	err = r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, err
	}

	// Copy input data to results if needed
	for i := range emails {
		if i < len(results) {
			results[i].EmailType = emails[i].EmailType
			results[i].EmailAddress = emails[i].EmailAddress
			results[i].IsPrimary = emails[i].IsPrimary
		}
	}

	return results, nil
}

// SetPrimaryEmail sets an email as primary and unsets others
// OPTIMIZATION: CTE pattern to update multiple emails and insert audit in single query
func (r *AgentEmailRepository) SetPrimaryEmail(ctx context.Context, emailID, agentID, updatedBy string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use CTE to combine multiple UPDATEs + INSERT audit in single query
	// CRITICAL: Golang variables cannot be passed between batch queries - must combine at SQL level
	batch := &pgx.Batch{}

	sql := `
		WITH unset_primary AS (
			UPDATE agent_emails
			SET is_primary = false, updated_at = $3, updated_by = $4
			WHERE agent_id = $1 AND deleted_at IS NULL
		),
		set_primary AS (
			UPDATE agent_emails
			SET is_primary = true, updated_at = $3, updated_by = $4
			WHERE email_id = $2 AND deleted_at IS NULL
			RETURNING agent_id
		)
		INSERT INTO agent_audit_logs (agent_id, action_type, field_name, new_value, action_reason, performed_by, performed_at)
		SELECT agent_id, $5, $6, $7, $8, $9, $10
		FROM set_primary
	`

	args := []interface{}{
		agentID, emailID, time.Now(), updatedBy,
		domain.AuditActionEmailUpdate, "is_primary", "true", "Primary email changed", updatedBy, time.Now(),
	}

	err := dbutil.QueueExecRowRaw(batch, sql, args...)
	if err != nil {
		return err
	}

	return r.db.SendBatch(cCtx, batch).Close()
}

// ValidateEmailUniqueness checks if email address is unique (excluding current email)
// VR-AGT-PRF-014: Email Format Validation
func (r *AgentEmailRepository) ValidateEmailUniqueness(ctx context.Context, emailAddress, excludeEmailID string) (bool, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("COUNT(*)").
		From(agentEmailTable).
		Where(sq.Eq{"email_address": emailAddress, "deleted_at": nil}).
		Where(sq.NotEq{"email_id": excludeEmailID})

	var count int64
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowTo[int64], &count)
	if err != nil {
		return false, err
	}

	return count == 0, nil
}
