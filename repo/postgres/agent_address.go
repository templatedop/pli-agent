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

// AgentAddressRepository handles all database operations for agent addresses
// E-02: Agent Address Entity
// BR-AGT-PRF-008: Multiple Address Types Support
// BR-AGT-PRF-009: Communication Address Same as Permanent Option
type AgentAddressRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewAgentAddressRepository creates a new agent address repository
func NewAgentAddressRepository(db *dblib.DB, cfg *config.Config) *AgentAddressRepository {
	return &AgentAddressRepository{
		db:  db,
		cfg: cfg,
	}
}

const agentAddressTable = "agent_addresses"

// Create inserts a new agent address
// FR-AGT-PRF-009: Address Management
// VR-AGT-PRF-008: Pincode Validation (6 digits)
// VR-AGT-PRF-009: Address Line1 Mandatory
// VR-AGT-PRF-010: City Mandatory
// VR-AGT-PRF-011: State Mandatory
// VR-AGT-PRF-012: Country Mandatory
func (r *AgentAddressRepository) Create(ctx context.Context, address domain.AgentAddress) (*domain.AgentAddress, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use CTE to combine INSERT + INSERT audit in single query
	// CRITICAL: Golang variables cannot be passed between batch queries - must combine at SQL level
	// BR-AGT-PRF-008: Multiple Address Types Support (OFFICIAL, PERMANENT, COMMUNICATION)
	// BR-AGT-PRF-005: Audit Logging
	batch := &pgx.Batch{}

	sql := `
		WITH inserted AS (
			INSERT INTO agent_addresses (
				agent_id, address_type, address_line1, address_line2, village,
				taluka, city, district, state, country, pincode,
				is_same_as_permanent, effective_from, metadata, created_by
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
			RETURNING *
		)
		INSERT INTO agent_audit_logs (agent_id, action_type, field_name, new_value, action_reason, performed_by, performed_at)
		SELECT agent_id, $16, $17, $18, $19, $20, $21
		FROM inserted
		RETURNING (SELECT ROW(address_id, agent_id, address_type, address_line1, address_line2, village,
			taluka, city, district, state, country, pincode, is_same_as_permanent, effective_from,
			metadata, created_at, updated_at, created_by, updated_by, deleted_at, version) FROM inserted)
	`

	args := []interface{}{
		address.AgentID, address.AddressType, address.AddressLine1, address.AddressLine2,
		address.Village, address.Taluka, address.City, address.District, address.State,
		address.Country, address.Pincode, address.IsSameAsPermanent, address.EffectiveFrom,
		address.Metadata, address.CreatedBy,
		domain.AuditActionAddressUpdate, "address_type", address.AddressType, "New address added", address.CreatedBy, time.Now(),
	}

	var result domain.AgentAddress
	err := dbutil.QueueReturnRowRaw(batch, sql, args, pgx.RowToStructByNameLax[domain.AgentAddress], &result)
	if err != nil {
		return nil, err
	}

	err = r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// FindByID retrieves an address by ID
func (r *AgentAddressRepository) FindByID(ctx context.Context, addressID string) (*domain.AgentAddress, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentAddressTable).
		Where(sq.Eq{"address_id": addressID, "deleted_at": nil})

	var address domain.AgentAddress
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentAddress], &address)
	if err != nil {
		return nil, err
	}

	return &address, nil
}

// FindByAgentID retrieves all addresses for an agent
// FR-AGT-PRF-009: Address Management
// BR-AGT-PRF-008: Multiple Address Types Support
func (r *AgentAddressRepository) FindByAgentID(ctx context.Context, agentID string) ([]domain.AgentAddress, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentAddressTable).
		Where(sq.Eq{"agent_id": agentID, "deleted_at": nil}).
		OrderBy("effective_from DESC")

	var addresses []domain.AgentAddress
	err := dblib.SelectRows(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentAddress], &addresses)
	if err != nil {
		return nil, err
	}

	return addresses, nil
}

// FindByAgentIDAndType retrieves a specific address type for an agent
// BR-AGT-PRF-008: Multiple Address Types Support
func (r *AgentAddressRepository) FindByAgentIDAndType(ctx context.Context, agentID, addressType string) (*domain.AgentAddress, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentAddressTable).
		Where(sq.Eq{"agent_id": agentID, "address_type": addressType, "deleted_at": nil}).
		OrderBy("effective_from DESC").
		Limit(1)

	var address domain.AgentAddress
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentAddress], &address)
	if err != nil {
		return nil, err
	}

	return &address, nil
}

// Update updates an agent address
// FR-AGT-PRF-009: Address Management
// VR-AGT-PRF-008 to VR-AGT-PRF-012: Address Validations
func (r *AgentAddressRepository) Update(ctx context.Context, addressID string, updates map[string]interface{}, updatedBy string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use CTE to combine UPDATE + INSERT audit logs in single query
	// CRITICAL: Golang variables cannot be passed between batch queries - must combine at SQL level
	// BR-AGT-PRF-005: Audit Logging
	batch := &pgx.Batch{}

	// Build SET clause dynamically
	setClauses := "updated_at = $2, updated_by = $3"
	args := []interface{}{addressID, time.Now(), updatedBy}
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
			UPDATE agent_addresses
			SET %s
			WHERE address_id = $1 AND deleted_at IS NULL
			RETURNING agent_id
		)
		INSERT INTO agent_audit_logs (agent_id, action_type, field_name, new_value, performed_by, performed_at)
		SELECT agent_id, $%d, unnest($%d::text[]), unnest($%d::text[]), $%d, $%d
		FROM updated
	`, setClauses, argIndex, argIndex+1, argIndex+2, argIndex+3, argIndex+4)

	args = append(args, domain.AuditActionAddressUpdate, fieldNames, newValues, updatedBy, time.Now())

	err := dbutil.QueueExecRowRaw(batch, sql, args...)
	if err != nil {
		return err
	}

	// Execute batch
	return r.db.SendBatch(cCtx, batch).Close()
}

// Delete soft deletes an agent address
func (r *AgentAddressRepository) Delete(ctx context.Context, addressID, deletedBy string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use CTE to combine UPDATE + INSERT audit in single query
	// CRITICAL: Golang variables cannot be passed between batch queries - must combine at SQL level
	batch := &pgx.Batch{}

	sql := `
		WITH updated AS (
			UPDATE agent_addresses
			SET deleted_at = $2, updated_by = $3
			WHERE address_id = $1 AND deleted_at IS NULL
			RETURNING agent_id
		)
		INSERT INTO agent_audit_logs (agent_id, action_type, field_name, action_reason, performed_by, performed_at)
		SELECT agent_id, $4, $5, $6, $7, $8
		FROM updated
	`

	args := []interface{}{
		addressID, time.Now(), deletedBy,
		domain.AuditActionDelete, "address", "Address deleted", deletedBy, time.Now(),
	}

	err := dbutil.QueueExecRowRaw(batch, sql, args...)
	if err != nil {
		return err
	}

	// Execute batch
	return r.db.SendBatch(cCtx, batch).Close()
}

// BatchCreate inserts multiple agent addresses in a single transaction
// OPTIMIZATION: Batch operation for multiple address inserts using UNNEST
// FR-AGT-PRF-009: Address Management
func (r *AgentAddressRepository) BatchCreate(ctx context.Context, addresses []domain.AgentAddress) ([]domain.AgentAddress, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	// Use UNNEST to bulk insert addresses with audit logs in single query
	// CRITICAL: Golang variables cannot be passed between batch queries - must use UNNEST pattern
	// BR-AGT-PRF-005: Audit Logging
	batch := &pgx.Batch{}

	// Prepare arrays for UNNEST
	agentIDs := make([]string, len(addresses))
	addressTypes := make([]string, len(addresses))
	addressLine1s := make([]string, len(addresses))
	addressLine2s := make([]interface{}, len(addresses))
	villages := make([]interface{}, len(addresses))
	talukas := make([]interface{}, len(addresses))
	cities := make([]string, len(addresses))
	districts := make([]interface{}, len(addresses))
	states := make([]string, len(addresses))
	countries := make([]string, len(addresses))
	pincodes := make([]string, len(addresses))
	isSameAsPermanents := make([]bool, len(addresses))
	effectiveFroms := make([]time.Time, len(addresses))
	metadatas := make([]interface{}, len(addresses))
	createdBys := make([]string, len(addresses))

	for i, address := range addresses {
		agentIDs[i] = address.AgentID
		addressTypes[i] = address.AddressType
		addressLine1s[i] = address.AddressLine1
		addressLine2s[i] = address.AddressLine2
		villages[i] = address.Village
		talukas[i] = address.Taluka
		cities[i] = address.City
		districts[i] = address.District
		states[i] = address.State
		countries[i] = address.Country
		pincodes[i] = address.Pincode
		isSameAsPermanents[i] = address.IsSameAsPermanent
		effectiveFroms[i] = address.EffectiveFrom
		metadatas[i] = address.Metadata
		createdBys[i] = address.CreatedBy
	}

	sql := `
		WITH inserted AS (
			INSERT INTO agent_addresses (
				agent_id, address_type, address_line1, address_line2, village, taluka,
				city, district, state, country, pincode, is_same_as_permanent,
				effective_from, metadata, created_by
			)
			SELECT * FROM UNNEST(
				$1::uuid[],
				$2::text[],
				$3::text[],
				$4::text[],
				$5::text[],
				$6::text[],
				$7::text[],
				$8::text[],
				$9::text[],
				$10::text[],
				$11::text[],
				$12::boolean[],
				$13::timestamp[],
				$14::jsonb[],
				$15::text[]
			)
			RETURNING *
		)
		INSERT INTO agent_audit_logs (agent_id, action_type, field_name, new_value, action_reason, performed_by, performed_at)
		SELECT agent_id, $16, $17, address_type, $18, created_by, NOW()
		FROM inserted
	`

	args := []interface{}{
		agentIDs, addressTypes, addressLine1s, addressLine2s, villages, talukas,
		cities, districts, states, countries, pincodes, isSameAsPermanents,
		effectiveFroms, metadatas, createdBys,
		domain.AuditActionAddressUpdate, "address_type", "New address added",
	}

	err := dbutil.QueueExecRowRaw(batch, sql, args...)
	if err != nil {
		return nil, err
	}

	// Execute batch
	err = r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, err
	}

	return addresses, nil
}

// CopyCommunicationFromPermanent copies permanent address to communication address
// BR-AGT-PRF-009: Communication Address Same as Permanent Option
func (r *AgentAddressRepository) CopyCommunicationFromPermanent(ctx context.Context, agentID, createdBy string) (*domain.AgentAddress, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use CTE to combine SELECT + INSERT + INSERT audit in single query
	// CRITICAL: Golang variables cannot be passed between batch queries - must combine at SQL level
	// BR-AGT-PRF-009: Communication Address Same as Permanent Option
	batch := &pgx.Batch{}

	sql := `
		WITH permanent_addr AS (
			SELECT * FROM agent_addresses
			WHERE agent_id = $1 AND address_type = $2 AND deleted_at IS NULL
			ORDER BY effective_from DESC
			LIMIT 1
		),
		inserted AS (
			INSERT INTO agent_addresses (
				agent_id, address_type, address_line1, address_line2, village, taluka,
				city, district, state, country, pincode, is_same_as_permanent,
				effective_from, metadata, created_by
			)
			SELECT agent_id, $3, address_line1, address_line2, village, taluka,
				city, district, state, country, pincode, $4,
				$5, metadata, $6
			FROM permanent_addr
			RETURNING *
		)
		INSERT INTO agent_audit_logs (agent_id, action_type, field_name, new_value, action_reason, performed_by, performed_at)
		SELECT agent_id, $7, $8, $9, $10, $11, $12
		FROM inserted
		RETURNING (SELECT ROW(address_id, agent_id, address_type, address_line1, address_line2, village, taluka,
			city, district, state, country, pincode, is_same_as_permanent, effective_from, metadata,
			created_at, created_by, updated_at, updated_by, deleted_at, version) FROM inserted)
	`

	args := []interface{}{
		agentID, domain.AddressTypePermanent, domain.AddressTypeCommunication, true, time.Now(), createdBy,
		domain.AuditActionAddressUpdate, "address_type", domain.AddressTypeCommunication,
		"Communication address copied from permanent", createdBy, time.Now(),
	}

	var result domain.AgentAddress
	err := dbutil.QueueReturnRowRaw(batch, sql, args, pgx.RowToStructByNameLax[domain.AgentAddress], &result)
	if err != nil {
		return nil, err
	}

	// Execute batch
	err = r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, err
	}

	return &result, nil
}
