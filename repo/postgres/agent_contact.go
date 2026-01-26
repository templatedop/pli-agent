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

// AgentContactRepository handles all database operations for agent contacts
// E-03: Agent Contact Entity
// BR-AGT-PRF-010: Phone Number Categories
type AgentContactRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewAgentContactRepository creates a new agent contact repository
func NewAgentContactRepository(db *dblib.DB, cfg *config.Config) *AgentContactRepository {
	return &AgentContactRepository{
		db:  db,
		cfg: cfg,
	}
}

const agentContactTable = "agent_contacts"

// Create inserts a new agent contact
// FR-AGT-PRF-010: Contact Management
// VR-AGT-PRF-013: Mobile Format Validation (10 digits)
func (r *AgentContactRepository) Create(ctx context.Context, contact domain.AgentContact) (*domain.AgentContact, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use CTE to combine INSERT + INSERT audit in single query
	// CRITICAL: Golang variables cannot be passed between batch queries - must combine at SQL level
	// BR-AGT-PRF-010: Phone Number Categories (MOBILE, OFFICIAL_LANDLINE, RESIDENT_LANDLINE)
	// BR-AGT-PRF-005: Audit Logging
	batch := &pgx.Batch{}

	sql := `
		WITH inserted AS (
			INSERT INTO agent_contacts (
				agent_id, contact_type, contact_number, is_primary,
				effective_from, metadata, created_by
			) VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING *
		)
		INSERT INTO agent_audit_logs (agent_id, action_type, field_name, new_value, action_reason, performed_by, performed_at)
		SELECT agent_id, $8, $9, $10, $11, $12, $13
		FROM inserted
		RETURNING (SELECT ROW(contact_id, agent_id, contact_type, contact_number, is_primary,
			effective_from, metadata, created_at, updated_at, created_by, updated_by, deleted_at, version) FROM inserted)
	`

	args := []interface{}{
		contact.AgentID, contact.ContactType, contact.ContactNumber, contact.IsPrimary,
		contact.EffectiveFrom, contact.Metadata, contact.CreatedBy,
		domain.AuditActionContactUpdate, "contact_type", contact.ContactType, "New contact added", contact.CreatedBy, time.Now(),
	}

	var result domain.AgentContact
	err := dbutil.QueueReturnRowRaw(batch, sql, args, pgx.RowToStructByNameLax[domain.AgentContact], &result)
	if err != nil {
		return nil, err
	}

	err = r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// FindByID retrieves a contact by ID
func (r *AgentContactRepository) FindByID(ctx context.Context, contactID string) (*domain.AgentContact, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentContactTable).
		Where(sq.Eq{"contact_id": contactID, "deleted_at": nil})

	var contact domain.AgentContact
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentContact], &contact)
	if err != nil {
		return nil, err
	}

	return &contact, nil
}

// FindByAgentID retrieves all contacts for an agent
// FR-AGT-PRF-010: Contact Management
// BR-AGT-PRF-010: Phone Number Categories
func (r *AgentContactRepository) FindByAgentID(ctx context.Context, agentID string) ([]domain.AgentContact, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentContactTable).
		Where(sq.Eq{"agent_id": agentID, "deleted_at": nil}).
		OrderBy("is_primary DESC, effective_from DESC")

	var contacts []domain.AgentContact
	err := dblib.SelectRows(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentContact], &contacts)
	if err != nil {
		return nil, err
	}

	return contacts, nil
}

// FindByAgentIDAndType retrieves a specific contact type for an agent
// BR-AGT-PRF-010: Phone Number Categories
func (r *AgentContactRepository) FindByAgentIDAndType(ctx context.Context, agentID, contactType string) (*domain.AgentContact, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentContactTable).
		Where(sq.Eq{"agent_id": agentID, "contact_type": contactType, "deleted_at": nil}).
		OrderBy("is_primary DESC, effective_from DESC").
		Limit(1)

	var contact domain.AgentContact
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentContact], &contact)
	if err != nil {
		return nil, err
	}

	return &contact, nil
}

// FindPrimaryContact retrieves the primary contact for an agent
func (r *AgentContactRepository) FindPrimaryContact(ctx context.Context, agentID string) (*domain.AgentContact, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentContactTable).
		Where(sq.Eq{"agent_id": agentID, "is_primary": true, "deleted_at": nil}).
		OrderBy("effective_from DESC").
		Limit(1)

	var contact domain.AgentContact
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentContact], &contact)
	if err != nil {
		return nil, err
	}

	return &contact, nil
}

// Update updates an agent contact
// FR-AGT-PRF-010: Contact Management
// VR-AGT-PRF-013: Mobile Format Validation
func (r *AgentContactRepository) Update(ctx context.Context, contactID string, updates map[string]interface{}, updatedBy string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use CTE to combine UPDATE + INSERT audit logs in single query
	// CRITICAL: Golang variables cannot be passed between batch queries - must combine at SQL level
	// BR-AGT-PRF-005: Audit Logging
	batch := &pgx.Batch{}

	// Build SET clause dynamically
	setClauses := "updated_at = $2, updated_by = $3"
	args := []interface{}{contactID, time.Now(), updatedBy}
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
			UPDATE agent_contacts
			SET %s
			WHERE contact_id = $1 AND deleted_at IS NULL
			RETURNING agent_id
		)
		INSERT INTO agent_audit_logs (agent_id, action_type, field_name, new_value, performed_by, performed_at)
		SELECT agent_id, $%d, unnest($%d::text[]), unnest($%d::text[]), $%d, $%d
		FROM updated
	`, setClauses, argIndex, argIndex+1, argIndex+2, argIndex+3, argIndex+4)

	args = append(args, domain.AuditActionContactUpdate, fieldNames, newValues, updatedBy, time.Now())

	err := dbutil.QueueExecRowRaw(batch, sql, args...)
	if err != nil {
		return err
	}

	// Execute batch
	return r.db.SendBatch(cCtx, batch).Close()
}

// Delete soft deletes an agent contact
func (r *AgentContactRepository) Delete(ctx context.Context, contactID, deletedBy string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use CTE to combine UPDATE + INSERT audit in single query
	// CRITICAL: Golang variables cannot be passed between batch queries - must combine at SQL level
	batch := &pgx.Batch{}

	sql := `
		WITH updated AS (
			UPDATE agent_contacts
			SET deleted_at = $2, updated_by = $3
			WHERE contact_id = $1 AND deleted_at IS NULL
			RETURNING agent_id
		)
		INSERT INTO agent_audit_logs (agent_id, action_type, field_name, action_reason, performed_by, performed_at)
		SELECT agent_id, $4, $5, $6, $7, $8
		FROM updated
	`

	args := []interface{}{
		contactID, time.Now(), deletedBy,
		domain.AuditActionDelete, "contact", "Contact deleted", deletedBy, time.Now(),
	}

	err := dbutil.QueueExecRowRaw(batch, sql, args...)
	if err != nil {
		return err
	}

	// Execute batch
	return r.db.SendBatch(cCtx, batch).Close()
}

// BatchCreate inserts multiple agent contacts in a single transaction
// OPTIMIZATION: Batch operation for multiple contact inserts using UNNEST
// FR-AGT-PRF-010: Contact Management
func (r *AgentContactRepository) BatchCreate(ctx context.Context, contacts []domain.AgentContact) ([]domain.AgentContact, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	// Use UNNEST to bulk insert contacts with audit logs in single query
	// CRITICAL: Golang variables cannot be passed between batch queries - must use UNNEST pattern
	// BR-AGT-PRF-005: Audit Logging
	batch := &pgx.Batch{}

	// Prepare arrays for UNNEST
	agentIDs := make([]string, len(contacts))
	contactTypes := make([]string, len(contacts))
	contactNumbers := make([]string, len(contacts))
	isPrimaries := make([]bool, len(contacts))
	effectiveFroms := make([]time.Time, len(contacts))
	metadatas := make([]interface{}, len(contacts))
	createdBys := make([]string, len(contacts))

	for i, contact := range contacts {
		agentIDs[i] = contact.AgentID
		contactTypes[i] = contact.ContactType
		contactNumbers[i] = contact.ContactNumber
		isPrimaries[i] = contact.IsPrimary
		effectiveFroms[i] = contact.EffectiveFrom
		metadatas[i] = contact.Metadata
		createdBys[i] = contact.CreatedBy
	}

	sql := `
		WITH inserted AS (
			INSERT INTO agent_contacts (agent_id, contact_type, contact_number, is_primary, effective_from, metadata, created_by)
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
		SELECT agent_id, $8, $9, contact_type, $10, created_by, NOW()
		FROM inserted
		RETURNING (SELECT array_agg(ROW(contact_id, agent_id, contact_type, contact_number, is_primary,
			effective_from, metadata, created_at, created_by, updated_at, updated_by, deleted_at, version)::agent_contacts) FROM inserted)
	`

	args := []interface{}{
		agentIDs, contactTypes, contactNumbers, isPrimaries, effectiveFroms, metadatas, createdBys,
		domain.AuditActionContactUpdate, "contact_type", "New contact added",
	}

	var results []domain.AgentContact
	err := dbutil.QueueReturnRaw(batch, sql, args, pgx.RowToStructByNameLax[domain.AgentContact], &results)
	if err != nil {
		return nil, err
	}

	// Execute batch
	err = r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, err
	}

	// Copy input data to results if needed
	for i := range contacts {
		if i < len(results) {
			results[i].ContactType = contacts[i].ContactType
			results[i].ContactNumber = contacts[i].ContactNumber
			results[i].IsPrimary = contacts[i].IsPrimary
		}
	}

	return results, nil
}

// SetPrimaryContact sets a contact as primary and unsets others
// OPTIMIZATION: CTE pattern to update multiple contacts and insert audit in single query
func (r *AgentContactRepository) SetPrimaryContact(ctx context.Context, contactID, agentID, updatedBy string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use CTE to combine multiple UPDATEs + INSERT audit in single query
	// CRITICAL: Golang variables cannot be passed between batch queries - must combine at SQL level
	batch := &pgx.Batch{}

	sql := `
		WITH unset_primary AS (
			UPDATE agent_contacts
			SET is_primary = false, updated_at = $3, updated_by = $4
			WHERE agent_id = $1 AND deleted_at IS NULL
		),
		set_primary AS (
			UPDATE agent_contacts
			SET is_primary = true, updated_at = $3, updated_by = $4
			WHERE contact_id = $2 AND deleted_at IS NULL
			RETURNING agent_id
		)
		INSERT INTO agent_audit_logs (agent_id, action_type, field_name, new_value, action_reason, performed_by, performed_at)
		SELECT agent_id, $5, $6, $7, $8, $9, $10
		FROM set_primary
	`

	args := []interface{}{
		agentID, contactID, time.Now(), updatedBy,
		domain.AuditActionContactUpdate, "is_primary", "true", "Primary contact changed", updatedBy, time.Now(),
	}

	err := dbutil.QueueExecRowRaw(batch, sql, args...)
	if err != nil {
		return err
	}

	return r.db.SendBatch(cCtx, batch).Close()
}
