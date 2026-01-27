package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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
	Status        string
	AgentType     string
	CircleID      string
	DivisionID    string
	CoordinatorID string
	OfficeCode    string
	Name          string
	MobileNumber  string
	Email         string
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
			"agent_type": domain.AgentTypeAdvisorCoordinator,
			"status":     domain.AgentStatusActive,
			"deleted_at": nil,
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

// ========================================================================
// ATOMIC BATCH OPERATIONS (Single Round-Trip to Database)
// ========================================================================

// CreateWithRelatedEntitiesInput holds all data for atomic profile creation
type CreateWithRelatedEntitiesInput struct {
	Profile   domain.AgentProfile
	Addresses []domain.AgentAddress
	Contacts  []domain.AgentContact
	Emails    []domain.AgentEmail
}

// CreateWithRelatedEntities atomically creates agent profile with all related entities
// Single database round trip using CTE pattern with UNNEST for bulk inserts
// ACT-024: CreateAgentProfileActivity
// FR-AGT-PRF-001: New Profile Creation
// Ensures atomicity: Either all entities are created, or none are (transaction)
func (r *AgentProfileRepository) CreateWithRelatedEntities(ctx context.Context, input CreateWithRelatedEntitiesInput) (*domain.AgentProfile, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutHigh"))
	defer cancel()

	profile := input.Profile

	// Build arrays for UNNEST bulk insert
	var (
		// Address arrays
		addrAgentIDs    []string
		addrTypes       []string
		addrLine1s      []string
		addrLine2s      []string
		addrLine3s      []string
		addrCities      []string
		addrDistricts   []string
		addrStates      []string
		addrCountries   []string
		addrPincodes    []string
		addrIsPrimaries []bool
		addrValidFroms  []time.Time
		addrCreatedBys  []string

		// Contact arrays
		contactAgentIDs    []string
		contactTypes       []string
		contactNumbers     []string
		contactIsPrimaries []bool
		contactIsVerifieds []bool
		contactCreatedBys  []string

		// Email arrays
		emailAgentIDs    []string
		emailAddresses   []string
		emailIsPrimaries []bool
		emailIsVerifieds []bool
		emailCreatedBys  []string
	)

	// We'll use a placeholder for agent_id - it will be filled from the CTE
	agentIDPlaceholder := "<<AGENT_ID>>"

	// Prepare address arrays
	for _, addr := range input.Addresses {
		addrAgentIDs = append(addrAgentIDs, agentIDPlaceholder)
		addrTypes = append(addrTypes, addr.AddressType)
		addrLine1s = append(addrLine1s, addr.Line1)
		addrLine2s = append(addrLine2s, addr.Line2.String)
		addrLine3s = append(addrLine3s, addr.Line3.String)
		addrCities = append(addrCities, addr.City)
		addrDistricts = append(addrDistricts, addr.District.String)
		addrStates = append(addrStates, addr.State)
		addrCountries = append(addrCountries, addr.Country)
		addrPincodes = append(addrPincodes, addr.Pincode)
		addrIsPrimaries = append(addrIsPrimaries, addr.IsPrimary)
		addrValidFroms = append(addrValidFroms, addr.ValidFrom)
		addrCreatedBys = append(addrCreatedBys, profile.CreatedBy)
	}

	// Prepare contact arrays
	for _, contact := range input.Contacts {
		contactAgentIDs = append(contactAgentIDs, agentIDPlaceholder)
		contactTypes = append(contactTypes, contact.ContactType)
		contactNumbers = append(contactNumbers, contact.ContactNumber)
		contactIsPrimaries = append(contactIsPrimaries, contact.IsPrimary)
		contactIsVerifieds = append(contactIsVerifieds, contact.IsVerified)
		contactCreatedBys = append(contactCreatedBys, profile.CreatedBy)
	}

	// Prepare email arrays
	for _, email := range input.Emails {
		emailAgentIDs = append(emailAgentIDs, agentIDPlaceholder)
		emailAddresses = append(emailAddresses, email.EmailAddress)
		emailIsPrimaries = append(emailIsPrimaries, email.IsPrimary)
		emailIsVerifieds = append(emailIsVerifieds, email.IsVerified)
		emailCreatedBys = append(emailCreatedBys, profile.CreatedBy)
	}

	// Build the complex CTE query
	// Uses CTEs to:
	// 1. INSERT profile and get agent_id
	// 2. INSERT addresses using UNNEST with agent_id from step 1
	// 3. INSERT contacts using UNNEST with agent_id from step 1
	// 4. INSERT emails using UNNEST with agent_id from step 1
	// 5. INSERT audit log with agent_id from step 1
	// All in single atomic transaction
	sql := `
		WITH inserted_profile AS (
			INSERT INTO agent_profiles (
				agent_type, employee_id, office_code, circle_id, division_id,
				advisor_coordinator_id, title, first_name, middle_name, last_name,
				gender, date_of_birth, category, marital_status, aadhar_number,
				pan_number, designation_rank, service_number, professional_title,
				status, status_date, distribution_channel, product_class,
				external_identification_number, workflow_state, created_by
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26)
			RETURNING *
		),
		inserted_addresses AS (
			INSERT INTO agent_addresses (agent_id, address_type, line1, line2, line3, city, district, state, country, pincode, is_primary, valid_from, created_by)
			SELECT ip.agent_id, addr_type, addr_line1, addr_line2, addr_line3, addr_city, addr_district, addr_state, addr_country, addr_pincode, addr_is_primary, addr_valid_from, addr_created_by
			FROM inserted_profile ip
			CROSS JOIN UNNEST(
				$27::text[], $28::text[], $29::text[], $30::text[], $31::text[], $32::text[], $33::text[], $34::text[], $35::text[],
				$36::boolean[], $37::timestamp[], $38::text[]
			) AS t(addr_type, addr_line1, addr_line2, addr_line3, addr_city, addr_district, addr_state, addr_country, addr_pincode, addr_is_primary, addr_valid_from, addr_created_by)
			WHERE ARRAY_LENGTH($27::text[], 1) > 0
			RETURNING *
		),
		inserted_contacts AS (
			INSERT INTO agent_contacts (agent_id, contact_type, contact_number, is_primary, is_verified, created_by)
			SELECT ip.agent_id, contact_type, contact_number, contact_is_primary, contact_is_verified, contact_created_by
			FROM inserted_profile ip
			CROSS JOIN UNNEST(
				$39::text[], $40::text[], $41::boolean[], $42::boolean[], $43::text[]
			) AS t(contact_type, contact_number, contact_is_primary, contact_is_verified, contact_created_by)
			WHERE ARRAY_LENGTH($39::text[], 1) > 0
			RETURNING *
		),
		inserted_emails AS (
			INSERT INTO agent_emails (agent_id, email_address, is_primary, is_verified, created_by)
			SELECT ip.agent_id, email_address, email_is_primary, email_is_verified, email_created_by
			FROM inserted_profile ip
			CROSS JOIN UNNEST(
				$44::text[], $45::boolean[], $46::boolean[], $47::text[]
			) AS t(email_address, email_is_primary, email_is_verified, email_created_by)
			WHERE ARRAY_LENGTH($44::text[], 1) > 0
			RETURNING *
		),
		inserted_audit AS (
			INSERT INTO agent_audit_logs (agent_id, action_type, action_reason, performed_by, performed_at)
			SELECT agent_id, $48, $49, $50, $51
			FROM inserted_profile
			RETURNING *
		)
		SELECT * FROM inserted_profile
	`

	args := []interface{}{
		// Profile fields ($1 to $26)
		profile.AgentType, profile.EmployeeID, profile.OfficeCode, profile.CircleID,
		profile.DivisionID, profile.AdvisorCoordinatorID, profile.Title, profile.FirstName,
		profile.MiddleName, profile.LastName, profile.Gender, profile.DateOfBirth,
		profile.Category, profile.MaritalStatus, profile.AadharNumber, profile.PANNumber,
		profile.DesignationRank, profile.ServiceNumber, profile.ProfessionalTitle,
		profile.Status, profile.StatusDate, profile.DistributionChannel, profile.ProductClass,
		profile.ExternalIdentificationNumber, profile.WorkflowState, profile.CreatedBy,
		// Address arrays ($27 to $38)
		addrTypes, addrLine1s, addrLine2s, addrLine3s, addrCities, addrDistricts,
		addrStates, addrCountries, addrPincodes, addrIsPrimaries, addrValidFroms, addrCreatedBys,
		// Contact arrays ($39 to $43)
		contactTypes, contactNumbers, contactIsPrimaries, contactIsVerifieds, contactCreatedBys,
		// Email arrays ($44 to $47)
		emailAddresses, emailIsPrimaries, emailIsVerifieds, emailCreatedBys,
		// Audit fields ($48 to $51)
		domain.AuditActionCreate, "Agent profile created", profile.CreatedBy, time.Now(),
	}

	batch := &pgx.Batch{}
	var result domain.AgentProfile
	err := dbutil.QueueReturnRowRaw(batch, sql, args, pgx.RowToStructByNameLax[domain.AgentProfile], &result)
	if err != nil {
		return nil, fmt.Errorf("failed to queue profile creation: %w", err)
	}

	// Execute batch
	err = r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, fmt.Errorf("failed to execute profile creation: %w", err)
	}

	return &result, nil
}

// Search performs multi-criteria agent search with pagination
// AGT-022: Search Agents
// FR-AGT-PRF-004: Multi-criteria agent search
// BR-AGT-PRF-022: Multi-Criteria Agent Search
// CRITICAL: Single database round trip with pagination
func (r *AgentProfileRepository) Search(
	ctx context.Context,
	agentID *string,
	name *string,
	panNumber *string,
	mobileNumber *string,
	email *string,
	status *string,
	officeCode *string,
	page, limit int,
) ([]domain.AgentProfile, int, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	offset := (page - 1) * limit

	// Build base query
	baseQuery := dblib.Psql.Select("p.*").
		From("agent_profiles p").
		LeftJoin("agent_contacts c ON p.agent_id = c.agent_id AND c.contact_type = 'PRIMARY' AND c.deleted_at IS NULL").
		LeftJoin("agent_emails e ON p.agent_id = e.agent_id AND e.email_type = 'PRIMARY' AND e.deleted_at IS NULL").
		Where(sq.Eq{"p.deleted_at": nil})

	// Apply filters dynamically
	if agentID != nil && *agentID != "" {
		baseQuery = baseQuery.Where(sq.Eq{"p.agent_id": *agentID})
	}
	if name != nil && *name != "" {
		baseQuery = baseQuery.Where(sq.Or{
			sq.ILike{"p.first_name": "%" + *name + "%"},
			sq.ILike{"p.last_name": "%" + *name + "%"},
			sq.Expr("CONCAT(p.first_name, ' ', p.last_name) ILIKE ?", "%"+*name+"%"),
		})
	}
	if panNumber != nil && *panNumber != "" {
		baseQuery = baseQuery.Where(sq.Eq{"p.pan_number": *panNumber})
	}
	if mobileNumber != nil && *mobileNumber != "" {
		baseQuery = baseQuery.Where(sq.Eq{"c.mobile_number": *mobileNumber})
	}
	if email != nil && *email != "" {
		baseQuery = baseQuery.Where(sq.Eq{"e.email_address": *email})
	}
	if status != nil && *status != "" {
		baseQuery = baseQuery.Where(sq.Eq{"p.status": *status})
	}
	if officeCode != nil && *officeCode != "" {
		baseQuery = baseQuery.Where(sq.Eq{"p.office_code": *officeCode})
	}

	// Use batch for single database round trip
	batch := &pgx.Batch{}

	// Query 1: Count total records
	countQuery := baseQuery
	countSQL, countArgs, _ := countQuery.ToSql()
	countSQL = "SELECT COUNT(DISTINCT p.agent_id) FROM (" + countSQL + ") AS subquery"

	var totalCount int
	err := dblib.QueueReturnRowRaw(batch, countSQL, countArgs, pgx.RowTo[int], &totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to queue count query: %w", err)
	}

	// Query 2: Get paginated data
	dataQuery := baseQuery.
		Distinct().
		OrderBy("p.created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	var profiles []domain.AgentProfile
	err = dblib.QueueReturn(batch, dataQuery, pgx.RowToStructByNameLax[domain.AgentProfile], &profiles)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to queue data query: %w", err)
	}

	// Execute batch in single round trip
	err = r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute search batch: %w", err)
	}

	return profiles, totalCount, nil
}

// GetProfileWithRelatedEntities retrieves complete agent profile with all related entities
// AGT-023: Get Agent Profile Details
// FR-AGT-PRF-005: Profile Dashboard View
// CRITICAL: Uses JSON aggregation for single query (no N+1 problem)
func (r *AgentProfileRepository) GetProfileWithRelatedEntities(
	ctx context.Context,
	agentID string,
) (*domain.AgentProfile, []domain.AgentAddress, []domain.AgentContact, []domain.AgentEmail, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use batch for single database round trip
	batch := &pgx.Batch{}

	// Query 1: Get profile
	profileQuery := dblib.Psql.Select("*").
		From(agentProfileTable).
		Where(sq.And{
			sq.Eq{"agent_id": agentID},
			sq.Eq{"deleted_at": nil},
		})

	var profile domain.AgentProfile
	err := dblib.QueueReturnRow(batch, profileQuery, pgx.RowToStructByNameLax[domain.AgentProfile], &profile)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to queue profile query: %w", err)
	}

	// Query 2: Get addresses
	addressQuery := dblib.Psql.Select("*").
		From("agent_addresses").
		Where(sq.And{
			sq.Eq{"agent_id": agentID},
			sq.Eq{"deleted_at": nil},
		}).
		OrderBy("is_primary DESC, created_at DESC")

	var addresses []domain.AgentAddress
	err = dblib.QueueReturn(batch, addressQuery, pgx.RowToStructByNameLax[domain.AgentAddress], &addresses)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to queue addresses query: %w", err)
	}

	// Query 3: Get contacts
	contactQuery := dblib.Psql.Select("*").
		From("agent_contacts").
		Where(sq.And{
			sq.Eq{"agent_id": agentID},
			sq.Eq{"deleted_at": nil},
		}).
		OrderBy("is_primary DESC, created_at DESC")

	var contacts []domain.AgentContact
	err = dblib.QueueReturn(batch, contactQuery, pgx.RowToStructByNameLax[domain.AgentContact], &contacts)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to queue contacts query: %w", err)
	}

	// Query 4: Get emails
	emailQuery := dblib.Psql.Select("*").
		From("agent_emails").
		Where(sq.And{
			sq.Eq{"agent_id": agentID},
			sq.Eq{"deleted_at": nil},
		}).
		OrderBy("is_primary DESC, created_at DESC")

	var emails []domain.AgentEmail
	err = dblib.QueueReturn(batch, emailQuery, pgx.RowToStructByNameLax[domain.AgentEmail], &emails)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to queue emails query: %w", err)
	}

	// Execute batch in single round trip
	err = r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to execute profile batch: %w", err)
	}

	return &profile, addresses, contacts, emails, nil
}

// UpdateSectionReturning updates profile section fields and creates audit logs
// AGT-025: Update Profile Section
// FR-AGT-PRF-006: Personal Information Update
// BR-AGT-PRF-005: Name Update with Audit Logging
// CRITICAL: Single SQL statement using CTE (capture old + UPDATE + bulk audit INSERT + return updated)
func (r *AgentProfileRepository) UpdateSectionReturning(
	ctx context.Context,
	agentID string,
	updates map[string]interface{},
	updatedBy string,
) (*domain.AgentProfile, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	now := time.Now()

	// Build dynamic SET clauses with parameterized values
	args := []interface{}{agentID, now, updatedBy}
	argIndex := 4
	var setClauses []string
	var fieldNames []string
	var oldValueSelects []string

	for field, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, argIndex))
		args = append(args, value)
		fieldNames = append(fieldNames, field)
		oldValueSelects = append(oldValueSelects, fmt.Sprintf("old.%s::text", field))
		argIndex++
	}

	// Build arrays for UNNEST: new values as text
	var newValuesArray []string
	for _, value := range updates {
		newValuesArray = append(newValuesArray, fmt.Sprintf("%v", value))
	}
	args = append(args, newValuesArray) // $argIndex

	// Build CTE: capture old -> update -> insert audits for changed fields -> return updated profile as JSON
	sql := fmt.Sprintf(`
		WITH old_profile AS (
			SELECT * FROM agent_profiles
			WHERE agent_id = $1 AND deleted_at IS NULL
		),
		updated_profile AS (
			UPDATE agent_profiles
			SET
				updated_at = $2,
				updated_by = $3,
				version = version + 1,
				%s
			WHERE agent_id = $1 AND deleted_at IS NULL
			RETURNING *
		),
		audit_changes AS (
			SELECT
				unnest(ARRAY['%s']::text[]) as field_name,
				unnest(ARRAY[%s]) as old_value,
				unnest($%d::text[]) as new_value
			FROM old_profile
		),
		inserted_audits AS (
			INSERT INTO agent_audit_logs (
				agent_id, action_type, field_name, old_value, new_value,
				performed_by, performed_at
			)
			SELECT
				$1,
				'%s',
				field_name,
				old_value,
				new_value,
				$3,
				$2
			FROM audit_changes
			WHERE old_value IS DISTINCT FROM new_value
			RETURNING audit_log_id
		)
		SELECT row_to_json(t.*) FROM updated_profile t`,
		strings.Join(setClauses, ",\n\t\t\t\t"),
		strings.Join(fieldNames, "', '"),
		strings.Join(oldValueSelects, ", "),
		argIndex,
		domain.AuditActionUpdate,
	)

	// Execute CTE and get updated profile as JSON
	var profileJSON []byte
	err := r.db.QueryRow(cCtx, sql, args...).Scan(&profileJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to update profile with audit: %w", err)
	}

	// Parse JSON to struct
	var result domain.AgentProfile
	err = json.Unmarshal(profileJSON, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse updated profile: %w", err)
	}

	return &result, nil
}

// ApproveRequestAndUpdateProfile approves update request AND applies profile changes in SINGLE database round trip
// AGT-026: Approve Profile Update - Ultimate optimization
// Uses PostgreSQL stored function: approve_request_and_update_profile()
// Single function call replaces 2 separate operations (approve + update)
func (r *AgentProfileRepository) ApproveRequestAndUpdateProfile(
	ctx context.Context,
	requestID string,
	approvedBy string,
	comments string,
) (*domain.AgentProfile, *domain.AgentProfileUpdateRequest, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutHigh"))
	defer cancel()

	// Call stored function - single database hit
	sql := `SELECT * FROM approve_request_and_update_profile($1, $2, $3)`

	var profileJSON, requestJSON []byte
	err := r.db.QueryRow(cCtx, sql, requestID, approvedBy, comments).Scan(&profileJSON, &requestJSON)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to approve and update in single call: %w", err)
	}

	// Parse profile JSON
	var profile domain.AgentProfile
	err = json.Unmarshal(profileJSON, &profile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse updated profile: %w", err)
	}

	// Parse request JSON
	var request domain.AgentProfileUpdateRequest
	err = json.Unmarshal(requestJSON, &request)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse approved request: %w", err)
	}

	return &profile, &request, nil
}

// GetHierarchy retrieves agent hierarchy chain using recursive CTE
// AGT-073: Get Agent Hierarchy
// Phase 9: Search & Dashboard APIs
// OPTIMIZED: Single recursive query traverses entire hierarchy chain
func (r *AgentProfileRepository) GetHierarchy(ctx context.Context, agentID string) ([]domain.HierarchyNode, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Recursive CTE to build hierarchy chain
	// Starts with given agent, follows advisor_coordinator_id links upward
	sql := `
		WITH RECURSIVE hierarchy AS (
			-- Base case: Start with the given agent
			SELECT
				agent_id,
				agent_code,
				CONCAT(first_name, ' ', COALESCE(middle_name || ' ', ''), last_name) AS name,
				agent_type,
				advisor_coordinator_id,
				1 AS level
			FROM agent_profiles
			WHERE agent_id = $1 AND deleted_at IS NULL

			UNION ALL

			-- Recursive case: Get coordinator/manager above
			SELECT
				p.agent_id,
				p.agent_code,
				CONCAT(p.first_name, ' ', COALESCE(p.middle_name || ' ', ''), p.last_name) AS name,
				p.agent_type,
				p.advisor_coordinator_id,
				h.level + 1 AS level
			FROM agent_profiles p
			INNER JOIN hierarchy h ON p.agent_id = h.advisor_coordinator_id
			WHERE p.deleted_at IS NULL
		)
		SELECT agent_id, agent_code, name, agent_type, level
		FROM hierarchy
		ORDER BY level ASC
	`

	var hierarchy []domain.HierarchyNode
	err := r.db.Select(cCtx, &hierarchy, sql, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get hierarchy: %w", err)
	}

	return hierarchy, nil
}

// Helper function to get field value from profile
func getFieldValue(profile *domain.AgentProfile, fieldName string) interface{} {
	switch fieldName {
	case "first_name":
		return profile.FirstName
	case "middle_name":
		return profile.MiddleName
	case "last_name":
		return profile.LastName
	case "pan_number":
		return profile.PANNumber
	case "aadhar_number":
		return profile.AadharNumber
	case "date_of_birth":
		return profile.DateOfBirth
	case "gender":
		return profile.Gender
	case "marital_status":
		return profile.MaritalStatus
	case "category":
		return profile.Category
	case "title":
		return profile.Title
	case "professional_title":
		return profile.ProfessionalTitle
	default:
		return ""
	}
}
