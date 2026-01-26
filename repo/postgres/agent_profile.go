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

// AgentProfileRepository handles all database operations for agent profiles
// E-01: Agent Profile Entity
type AgentProfileRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewAgentProfileRepository creates a new agent profile repository
func NewAgentProfileRepository(db *dblib.DB, cfg *config.Config) *AgentProfileRepository {
	return &AgentProfileRepository{
		db:  db,
		cfg: cfg,
	}
}

const agentProfileTable = "agent_profiles"

// Create inserts a new agent profile
// FR-AGT-PRF-001: New Profile Creation
// BR-AGT-PRF-001: Advisor Coordinator Linkage Requirement
// BR-AGT-PRF-006: PAN Update with Format and Uniqueness Validation
func (r *AgentProfileRepository) Create(ctx context.Context, profile domain.AgentProfile) (*domain.AgentProfile, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use CTE to combine INSERT profile + INSERT audit in single query
	// CRITICAL: Golang variables cannot be passed between batch queries - must combine at SQL level
	// VR-AGT-PRF-001 to VR-AGT-PRF-007: Personal Information Validation
	// VR-AGT-PRF-003: PAN Format Validation
	// VR-AGT-PRF-004: Aadhar Format Validation
	// BR-AGT-PRF-005: Name Update with Audit Logging
	batch := &pgx.Batch{}

	sql := `
		WITH inserted AS (
			INSERT INTO agent_profiles (
				agent_type, employee_id, office_code, circle_id, division_id,
				advisor_coordinator_id, title, first_name, middle_name, last_name,
				gender, date_of_birth, category, marital_status, aadhar_number,
				pan_number, designation_rank, service_number, professional_title,
				status, status_date, distribution_channel, product_class,
				external_identification_number, workflow_state, created_by
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26)
			RETURNING *
		)
		INSERT INTO agent_audit_logs (agent_id, action_type, action_reason, performed_by, performed_at)
		SELECT agent_id, $27, $28, $29, $30
		FROM inserted
		RETURNING (SELECT ROW(agent_id, agent_code, agent_type, employee_id, office_code, circle_id, division_id,
			advisor_coordinator_id, title, first_name, middle_name, last_name, gender, date_of_birth,
			category, marital_status, aadhar_number, pan_number, designation_rank, service_number,
			professional_title, status, status_date, status_reason, distribution_channel, product_class,
			external_identification_number, workflow_state, created_at, created_by, updated_at, updated_by,
			deleted_at, version) FROM inserted)
	`

	args := []interface{}{
		profile.AgentType, profile.EmployeeID, profile.OfficeCode, profile.CircleID,
		profile.DivisionID, profile.AdvisorCoordinatorID, profile.Title, profile.FirstName,
		profile.MiddleName, profile.LastName, profile.Gender, profile.DateOfBirth,
		profile.Category, profile.MaritalStatus, profile.AadharNumber, profile.PANNumber,
		profile.DesignationRank, profile.ServiceNumber, profile.ProfessionalTitle,
		profile.Status, profile.StatusDate, profile.DistributionChannel, profile.ProductClass,
		profile.ExternalIdentificationNumber, profile.WorkflowState, profile.CreatedBy,
		domain.AuditActionCreate, "Agent profile created", profile.CreatedBy, time.Now(),
	}

	var result domain.AgentProfile
	err := dbutil.QueueReturnRowRaw(batch, sql, args, pgx.RowToStructByNameLax[domain.AgentProfile], &result)
	if err != nil {
		return nil, err
	}

	// Execute batch
	err = r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, err
	}

	// Copy input data to result
	result = profile
	return &result, nil
}

// FindByID retrieves an agent profile by ID
// FR-AGT-PRF-021: Multi-Criteria Agent Search
func (r *AgentProfileRepository) FindByID(ctx context.Context, agentID string) (*domain.AgentProfile, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentProfileTable).
		Where(sq.Eq{"agent_id": agentID, "deleted_at": nil})

	var profile domain.AgentProfile
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentProfile], &profile)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

// FindByAgentCode retrieves an agent profile by agent code
// FR-AGT-PRF-021: Multi-Criteria Agent Search
func (r *AgentProfileRepository) FindByAgentCode(ctx context.Context, agentCode string) (*domain.AgentProfile, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentProfileTable).
		Where(sq.Eq{"agent_code": agentCode, "deleted_at": nil})

	var profile domain.AgentProfile
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentProfile], &profile)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

// FindByPAN retrieves an agent profile by PAN number
// BR-AGT-PRF-006: PAN Update with Format and Uniqueness Validation
// VR-AGT-PRF-003: PAN Format Validation
func (r *AgentProfileRepository) FindByPAN(ctx context.Context, panNumber string) (*domain.AgentProfile, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentProfileTable).
		Where(sq.Eq{"pan_number": panNumber, "deleted_at": nil})

	var profile domain.AgentProfile
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentProfile], &profile)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

// List retrieves agent profiles with pagination and filtering
// FR-AGT-PRF-021: Multi-Criteria Agent Search
// BR-AGT-PRF-022: Multi-Criteria Agent Search
type AgentSearchFilters struct {
	Status             string
	AgentType          string
	CircleID           string
	DivisionID         string
	CoordinatorID      string
	OfficeCode         string
	Name               string
	MobileNumber       string
	Email              string
}

func (r *AgentProfileRepository) List(ctx context.Context, filters AgentSearchFilters, skip, limit uint64, orderBy, sortType string) ([]domain.AgentProfile, int64, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	// Use batch to get count and data in single round trip
	// OPTIMIZATION: Batch combines COUNT + SELECT queries
	batch := &pgx.Batch{}

	// Base query with filters
	baseQuery := sq.Select().From(agentProfileTable).Where(sq.Eq{"deleted_at": nil})

	// Apply filters
	if filters.Status != "" {
		baseQuery = baseQuery.Where(sq.Eq{"status": filters.Status})
	}
	if filters.AgentType != "" {
		baseQuery = baseQuery.Where(sq.Eq{"agent_type": filters.AgentType})
	}
	if filters.CircleID != "" {
		baseQuery = baseQuery.Where(sq.Eq{"circle_id": filters.CircleID})
	}
	if filters.DivisionID != "" {
		baseQuery = baseQuery.Where(sq.Eq{"division_id": filters.DivisionID})
	}
	if filters.CoordinatorID != "" {
		baseQuery = baseQuery.Where(sq.Eq{"advisor_coordinator_id": filters.CoordinatorID})
	}
	if filters.OfficeCode != "" {
		baseQuery = baseQuery.Where(sq.Eq{"office_code": filters.OfficeCode})
	}
	if filters.Name != "" {
		// Search in first_name, middle_name, last_name
		namePattern := "%" + filters.Name + "%"
		baseQuery = baseQuery.Where(
			sq.Or{
				sq.Like{"first_name": namePattern},
				sq.Like{"middle_name": namePattern},
				sq.Like{"last_name": namePattern},
			},
		)
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

	var profiles []domain.AgentProfile
	err = dblib.QueueReturn(batch, dataQuery, pgx.RowToStructByNameLax[domain.AgentProfile], &profiles)
	if err != nil {
		return nil, 0, err
	}

	// Execute batch
	err = r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, 0, err
	}

	return profiles, totalCount, nil
}

// UpdatePersonalInfo updates agent personal information
// FR-AGT-PRF-006: Profile Update
// BR-AGT-PRF-005: Name Update with Audit Logging
// BR-AGT-PRF-007: Personal Information Update Rules
func (r *AgentProfileRepository) UpdatePersonalInfo(ctx context.Context, agentID string, updates map[string]interface{}, updatedBy string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use CTE to combine UPDATE + INSERT audit logs in single query
	// CRITICAL: Golang variables cannot be passed between batch queries - must combine at SQL level
	// BR-AGT-PRF-005: Audit Logging
	batch := &pgx.Batch{}

	// Build SET clause dynamically
	setClauses := "updated_at = $2, updated_by = $3"
	args := []interface{}{agentID, time.Now(), updatedBy}
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
			UPDATE agent_profiles
			SET %s
			WHERE agent_id = $1 AND deleted_at IS NULL
			RETURNING agent_id
		)
		INSERT INTO agent_audit_logs (agent_id, action_type, field_name, new_value, performed_by, performed_at)
		SELECT agent_id, $%d, unnest($%d::text[]), unnest($%d::text[]), $%d, $%d
		FROM updated
	`, setClauses, argIndex, argIndex+1, argIndex+2, argIndex+3, argIndex+4)

	args = append(args, domain.AuditActionUpdate, fieldNames, newValues, updatedBy, time.Now())

	err := dbutil.QueueExecRowRaw(batch, sql, args...)
	if err != nil {
		return err
	}

	// Execute batch
	return r.db.SendBatch(cCtx, batch).Close()
}

// UpdateStatus updates agent status
// FR-AGT-PRF-017: Status Management
// BR-AGT-PRF-016: Status Update with Mandatory Reason
func (r *AgentProfileRepository) UpdateStatus(ctx context.Context, agentID, status, reason, updatedBy string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use CTE to combine UPDATE + INSERT audit in single query
	// CRITICAL: Golang variables cannot be passed between batch queries - must combine at SQL level
	// BR-AGT-PRF-016: Status Update with Mandatory Reason
	batch := &pgx.Batch{}

	sql := `
		WITH updated AS (
			UPDATE agent_profiles
			SET status = $2, status_date = $3, status_reason = $4, updated_at = $5, updated_by = $6
			WHERE agent_id = $1 AND deleted_at IS NULL
			RETURNING agent_id
		)
		INSERT INTO agent_audit_logs (agent_id, action_type, field_name, new_value, action_reason, performed_by, performed_at)
		SELECT agent_id, $7, $8, $9, $10, $11, $12
		FROM updated
	`

	args := []interface{}{
		agentID, status, time.Now(), reason, time.Now(), updatedBy,
		domain.AuditActionStatusChange, "status", status, reason, updatedBy, time.Now(),
	}

	err := dbutil.QueueExecRowRaw(batch, sql, args...)
	if err != nil {
		return err
	}

	// Execute batch
	return r.db.SendBatch(cCtx, batch).Close()
}

// Terminate terminates an agent
// FR-AGT-PRF-018: Agent Termination
// BR-AGT-PRF-017: Agent Termination Workflow
func (r *AgentProfileRepository) Terminate(ctx context.Context, agentID, reason, terminatedBy string, effectiveDate time.Time) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use CTE to combine UPDATE + INSERT audit in single query
	// CRITICAL: Golang variables cannot be passed between batch queries - must combine at SQL level
	// BR-AGT-PRF-017: Agent Termination Workflow
	batch := &pgx.Batch{}

	sql := `
		WITH updated AS (
			UPDATE agent_profiles
			SET status = $2, status_date = $3, status_reason = $4, updated_at = $5, updated_by = $6
			WHERE agent_id = $1 AND deleted_at IS NULL
			RETURNING agent_id
		)
		INSERT INTO agent_audit_logs (agent_id, action_type, field_name, new_value, action_reason, performed_by, performed_at)
		SELECT agent_id, $7, $8, $9, $10, $11, $12
		FROM updated
	`

	args := []interface{}{
		agentID, domain.AgentStatusTerminated, effectiveDate, reason, time.Now(), terminatedBy,
		domain.AuditActionTerminate, "status", domain.AgentStatusTerminated, reason, terminatedBy, time.Now(),
	}

	err := dbutil.QueueExecRowRaw(batch, sql, args...)
	if err != nil {
		return err
	}

	// Execute batch
	return r.db.SendBatch(cCtx, batch).Close()
}

// GetActiveAdvisorCoordinators retrieves active advisor coordinators for dropdown
// FR-AGT-PRF-003: Advisor Coordinator Selection
// BR-AGT-PRF-001: Advisor Coordinator Linkage Requirement
// BR-AGT-PRF-002: Advisor Coordinator Geographic Assignment
func (r *AgentProfileRepository) GetActiveAdvisorCoordinators(ctx context.Context, circleID, divisionID string) ([]domain.AgentProfile, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("agent_id", "agent_code", "first_name", "middle_name", "last_name", "circle_id", "division_id").
		From(agentProfileTable).
		Where(sq.Eq{
			"agent_type":  domain.AgentTypeAdvisorCoordinator,
			"status":      domain.AgentStatusActive,
			"deleted_at":  nil,
		})

	// Apply geographic filters if provided
	// BR-AGT-PRF-002: Advisor Coordinator Geographic Assignment
	if circleID != "" {
		query = query.Where(sq.Eq{"circle_id": circleID})
	}
	if divisionID != "" {
		query = query.Where(sq.Eq{"division_id": divisionID})
	}

	query = query.OrderBy("first_name ASC")

	var coordinators []domain.AgentProfile
	err := dblib.SelectRows(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentProfile], &coordinators)
	if err != nil {
		return nil, err
	}

	return coordinators, nil
}

// ValidatePANUniqueness checks if PAN is unique (excluding current agent)
// BR-AGT-PRF-006: PAN Update with Format and Uniqueness Validation
func (r *AgentProfileRepository) ValidatePANUniqueness(ctx context.Context, panNumber, excludeAgentID string) (bool, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("COUNT(*)").
		From(agentProfileTable).
		Where(sq.Eq{"pan_number": panNumber, "deleted_at": nil}).
		Where(sq.NotEq{"agent_id": excludeAgentID})

	var count int64
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowTo[int64], &count)
	if err != nil {
		return false, err
	}

	return count == 0, nil
}
