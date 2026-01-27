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

// AgentLicenseRepository handles all database operations for agent licenses
// E-06: Agent License Entity
// BR-AGT-PRF-012: License Renewal Period Rules
// BR-AGT-PRF-013: Auto-Deactivation on License Expiry
// BR-AGT-PRF-014: License Renewal Reminder Schedule
type AgentLicenseRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewAgentLicenseRepository creates a new agent license repository
func NewAgentLicenseRepository(db *dblib.DB, cfg *config.Config) *AgentLicenseRepository {
	return &AgentLicenseRepository{
		db:  db,
		cfg: cfg,
	}
}

const agentLicenseTable = "agent_licenses"

// Create inserts a new agent license
// FR-AGT-PRF-014: License Management
// BR-AGT-PRF-012: License Renewal Period Rules
// BR-AGT-PRF-030: License Date Tracking
// VR-AGT-PRF-018: License Line Validation
// VR-AGT-PRF-019: License Type Validation
// VR-AGT-PRF-020: License Number Uniqueness
// VR-AGT-PRF-021: Resident Status Validation
// VR-AGT-PRF-022: License Date Validation
// VR-AGT-PRF-023: Renewal Date Validation
func (r *AgentLicenseRepository) Create(ctx context.Context, license domain.AgentLicense) (*domain.AgentLicense, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use CTE to combine INSERT + INSERT audit in single query
	// CRITICAL: Golang variables cannot be passed between batch queries - must combine at SQL level
	// BR-AGT-PRF-012: Provisional license valid for 1 year, renewable max 2 times
	// BR-AGT-PRF-030: Track license_date, renewal_date, authority_date
	// BR-AGT-PRF-005: Audit Logging
	batch := &pgx.Batch{}

	sql := `
		WITH inserted AS (
			INSERT INTO agent_licenses (
				agent_id, license_line, license_type, license_number, resident_status,
				license_date, renewal_date, authority_date, renewal_count, license_status,
				licentiate_exam_passed, licentiate_exam_date, licentiate_certificate_number,
				is_primary, metadata, created_by
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
			RETURNING *
		)
		INSERT INTO agent_audit_logs (agent_id, action_type, field_name, new_value, action_reason, performed_by, performed_at)
		SELECT agent_id, $17, $18, $19, $20, $21, $22
		FROM inserted
		RETURNING (SELECT ROW(license_id, agent_id, license_line, license_type, license_number, resident_status,
			license_date, renewal_date, authority_date, renewal_count, license_status, licentiate_exam_passed,
			licentiate_exam_date, licentiate_certificate_number, is_primary, metadata,
			created_at, updated_at, created_by, updated_by, deleted_at, version) FROM inserted)
	`

	args := []interface{}{
		license.AgentID, license.LicenseLine, license.LicenseType, license.LicenseNumber,
		license.ResidentStatus, license.LicenseDate, license.RenewalDate, license.AuthorityDate,
		license.RenewalCount, license.LicenseStatus, license.LicentiateExamPassed,
		license.LicentiateExamDate, license.LicentiateCertificateNumber, license.IsPrimary,
		license.Metadata, license.CreatedBy,
		domain.AuditActionLicenseAdd, "license_number", license.LicenseNumber, "New license added", license.CreatedBy, time.Now(),
	}

	var result domain.AgentLicense
	err := dbutil.QueueReturnRowRaw(batch, sql, args, pgx.RowToStructByNameLax[domain.AgentLicense], &result)
	if err != nil {
		return nil, err
	}

	err = r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// FindByID retrieves a license by ID
func (r *AgentLicenseRepository) FindByID(ctx context.Context, licenseID string) (*domain.AgentLicense, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentLicenseTable).
		Where(sq.Eq{"license_id": licenseID, "deleted_at": nil})

	var license domain.AgentLicense
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentLicense], &license)
	if err != nil {
		return nil, err
	}

	return &license, nil
}

// FindByAgentID retrieves all licenses for an agent
// FR-AGT-PRF-014: License Management
func (r *AgentLicenseRepository) FindByAgentID(ctx context.Context, agentID string) ([]domain.AgentLicense, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentLicenseTable).
		Where(sq.Eq{"agent_id": agentID, "deleted_at": nil}).
		OrderBy("is_primary DESC, license_date DESC")

	var licenses []domain.AgentLicense
	err := dblib.SelectRows(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentLicense], &licenses)
	if err != nil {
		return nil, err
	}

	return licenses, nil
}

// FindPrimaryLicense retrieves the primary license for an agent
// BR-AGT-PRF-012: License Renewal Period Rules
func (r *AgentLicenseRepository) FindPrimaryLicense(ctx context.Context, agentID string) (*domain.AgentLicense, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentLicenseTable).
		Where(sq.Eq{"agent_id": agentID, "is_primary": true, "deleted_at": nil}).
		OrderBy("license_date DESC").
		Limit(1)

	var license domain.AgentLicense
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentLicense], &license)
	if err != nil {
		return nil, err
	}

	return &license, nil
}

// FindByLicenseNumber retrieves a license by license number
// VR-AGT-PRF-020: License Number Uniqueness
func (r *AgentLicenseRepository) FindByLicenseNumber(ctx context.Context, licenseNumber string) (*domain.AgentLicense, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentLicenseTable).
		Where(sq.Eq{"license_number": licenseNumber, "deleted_at": nil}).
		Limit(1)

	var license domain.AgentLicense
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentLicense], &license)
	if err != nil {
		return nil, err
	}

	return &license, nil
}

// Update updates an agent license and returns the updated license
// FR-AGT-PRF-014: License Management
// BR-AGT-PRF-012: License Renewal Period Rules
// OPTIMIZED: Returns updated license using RETURNING to eliminate extra SELECT
func (r *AgentLicenseRepository) Update(ctx context.Context, licenseID string, updates map[string]interface{}, updatedBy string) (*domain.AgentLicense, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use CTE to combine UPDATE + INSERT audit logs in single query
	// CRITICAL: Golang variables cannot be passed between batch queries - must combine at SQL level
	// BR-AGT-PRF-005: Audit Logging
	// RETURNING clause eliminates need for separate SELECT after update
	batch := &pgx.Batch{}

	// Build SET clause dynamically
	setClauses := "updated_at = $2, updated_by = $3, version = version + 1"
	args := []interface{}{licenseID, time.Now(), updatedBy}
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
			UPDATE agent_licenses
			SET %s
			WHERE license_id = $1 AND deleted_at IS NULL
			RETURNING *
		),
		audit_insert AS (
			INSERT INTO agent_audit_logs (agent_id, action_type, field_name, new_value, performed_by, performed_at)
			SELECT agent_id, $%d, unnest($%d::text[]), unnest($%d::text[]), $%d, $%d
			FROM updated
		)
		SELECT * FROM updated
	`, setClauses, argIndex, argIndex+1, argIndex+2, argIndex+3, argIndex+4)

	args = append(args, domain.AuditActionLicenseUpdate, fieldNames, newValues, updatedBy, time.Now())

	var updatedLicense domain.AgentLicense
	err := dbutil.QueueReturnRowRaw(batch, sql, args, pgx.RowToStructByNameLax[domain.AgentLicense], &updatedLicense)
	if err != nil {
		return nil, err
	}

	err = r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, err
	}

	return &updatedLicense, nil
}

// RenewLicense renews a license and returns the renewed license
// BR-AGT-PRF-012: License Renewal Period Rules
// Provisional: Max 2 renewals, 1 year each
// Permanent: Renewable every 1 year after 5-year validity
// OPTIMIZED: Returns renewed license using RETURNING to eliminate extra SELECT
func (r *AgentLicenseRepository) RenewLicense(ctx context.Context, licenseID, updatedBy string, newRenewalDate time.Time) (*domain.AgentLicense, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use CTE to combine UPDATE + INSERT audit in single query
	// CRITICAL: Golang variables cannot be passed between batch queries - must combine at SQL level
	// RETURNING clause eliminates need for separate SELECT after renewal
	batch := &pgx.Batch{}

	sql := `
		WITH updated AS (
			UPDATE agent_licenses
			SET renewal_count = renewal_count + 1, renewal_date = $2, license_status = $3,
				updated_at = $4, updated_by = $5, version = version + 1
			WHERE license_id = $1 AND deleted_at IS NULL
			RETURNING *
		),
		audit_insert AS (
			INSERT INTO agent_audit_logs (agent_id, action_type, field_name, new_value, action_reason, performed_by, performed_at)
			SELECT agent_id, $6, $7, $8, $9, $10, $11
			FROM updated
		)
		SELECT * FROM updated
	`

	args := []interface{}{
		licenseID, newRenewalDate, domain.LicenseStatusRenewed, time.Now(), updatedBy,
		domain.AuditActionLicenseUpdate, "renewal_date", newRenewalDate, "License renewed", updatedBy, time.Now(),
	}

	var renewedLicense domain.AgentLicense
	err := dbutil.QueueReturnRowRaw(batch, sql, args, pgx.RowToStructByNameLax[domain.AgentLicense], &renewedLicense)
	if err != nil {
		return nil, err
	}

	err = r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, err
	}

	return &renewedLicense, nil
}

// ConvertToPermanent converts a provisional license to permanent and returns the converted license
// BR-AGT-PRF-012: License Renewal Period Rules
// After passing exam within 3 years: Permanent license with 5-year validity, renewable every 1 year
// OPTIMIZED: Returns converted license using RETURNING to eliminate extra SELECT
func (r *AgentLicenseRepository) ConvertToPermanent(ctx context.Context, licenseID, updatedBy string, examDate time.Time, certificateNumber string) (*domain.AgentLicense, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use CTE to combine UPDATE + INSERT audit in single query
	// CRITICAL: Golang variables cannot be passed between batch queries - must combine at SQL level
	// Calculate new renewal date: 5 years from conversion, then renewable every 1 year
	// RETURNING clause eliminates need for separate SELECT after conversion
	batch := &pgx.Batch{}

	permanentValidityDate := examDate.AddDate(5, 0, 0)

	sql := `
		WITH updated AS (
			UPDATE agent_licenses
			SET license_type = $2, licentiate_exam_passed = $3, licentiate_exam_date = $4,
				licentiate_certificate_number = $5, renewal_date = $6, license_status = $7,
				updated_at = $8, updated_by = $9, version = version + 1
			WHERE license_id = $1 AND deleted_at IS NULL
			RETURNING *
		),
		audit_insert AS (
			INSERT INTO agent_audit_logs (agent_id, action_type, field_name, new_value, action_reason, performed_by, performed_at)
			SELECT agent_id, $10, $11, $12, $13, $14, $15
			FROM updated
		)
		SELECT * FROM updated
	`

	args := []interface{}{
		licenseID, domain.LicenseTypePermanent, true, examDate, certificateNumber,
		permanentValidityDate, domain.LicenseStatusActive, time.Now(), updatedBy,
		domain.AuditActionLicenseUpdate, "license_type", domain.LicenseTypePermanent,
		"Converted to permanent after passing licentiate exam", updatedBy, time.Now(),
	}

	var convertedLicense domain.AgentLicense
	err := dbutil.QueueReturnRowRaw(batch, sql, args, pgx.RowToStructByNameLax[domain.AgentLicense], &convertedLicense)
	if err != nil {
		return nil, err
	}

	err = r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, err
	}

	return &convertedLicense, nil
}

// FindExpiringLicenses retrieves licenses expiring within specified days
// BR-AGT-PRF-014: License Renewal Reminder Schedule
// Used for sending reminders at 30, 15, 7 days before expiry and on expiry day
func (r *AgentLicenseRepository) FindExpiringLicenses(ctx context.Context, daysUntilExpiry int) ([]domain.AgentLicense, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	// Calculate the target date for expiry check
	targetDate := time.Now().AddDate(0, 0, daysUntilExpiry)

	query := dblib.Psql.Select("*").
		From(agentLicenseTable).
		Where(sq.Eq{"license_status": domain.LicenseStatusActive, "deleted_at": nil}).
		Where(sq.Expr("DATE(renewal_date) = DATE(?)", targetDate))

	var licenses []domain.AgentLicense
	err := dblib.SelectRows(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentLicense], &licenses)
	if err != nil {
		return nil, err
	}

	return licenses, nil
}

// FindExpiringLicensesWithAgentDetails retrieves licenses expiring within specified days WITH agent profile data
// AGT-036: Get Expiring Licenses - Optimized with JOIN to eliminate N+1 query problem
// BR-AGT-PRF-014: License Renewal Reminder Schedule
// SINGLE database hit: JOINs agent_licenses with agent_profiles
func (r *AgentLicenseRepository) FindExpiringLicensesWithAgentDetails(ctx context.Context, daysUntilExpiry int) ([]domain.AgentLicenseWithProfile, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	// Calculate date range for expiry check
	now := time.Now()
	futureDate := now.AddDate(0, 0, daysUntilExpiry)

	// JOIN agent_licenses with agent_profiles to get all data in single query
	// Eliminates N+1 query problem (1 query instead of 1 + N queries)
	query := dblib.Psql.
		Select(
			"al.license_id",
			"al.agent_id",
			"al.license_line",
			"al.license_type",
			"al.license_number",
			"al.license_date",
			"al.renewal_date",
			"al.renewal_count",
			"al.license_status",
			"ap.agent_code",
			"ap.first_name",
			"ap.middle_name",
			"ap.last_name",
			"ap.office_code",
		).
		From("agent_licenses al").
		Join("agent_profiles ap ON al.agent_id = ap.agent_id").
		Where(sq.Eq{"al.license_status": domain.LicenseStatusActive}).
		Where(sq.Eq{"al.deleted_at": nil}).
		Where(sq.Eq{"ap.deleted_at": nil}).
		Where(sq.LtOrEq{"al.renewal_date": futureDate}).
		Where(sq.GtOrEq{"al.renewal_date": now}).
		OrderBy("al.renewal_date ASC")

	var licenses []domain.AgentLicenseWithProfile
	err := dblib.SelectRows(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentLicenseWithProfile], &licenses)
	if err != nil {
		return nil, err
	}

	return licenses, nil
}

// FindExpiredLicenses retrieves all expired licenses that need auto-deactivation
// BR-AGT-PRF-013: Auto-Deactivation on License Expiry
func (r *AgentLicenseRepository) FindExpiredLicenses(ctx context.Context) ([]domain.AgentLicense, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentLicenseTable).
		Where(sq.Eq{"license_status": domain.LicenseStatusActive, "deleted_at": nil}).
		Where(sq.Lt{"renewal_date": time.Now()})

	var licenses []domain.AgentLicense
	err := dblib.SelectRows(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentLicense], &licenses)
	if err != nil {
		return nil, err
	}

	return licenses, nil
}

// MarkAsExpired marks a license as expired
// BR-AGT-PRF-013: Auto-Deactivation on License Expiry
func (r *AgentLicenseRepository) MarkAsExpired(ctx context.Context, licenseID, updatedBy string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use CTE to combine UPDATE + INSERT audit in single query
	// CRITICAL: Golang variables cannot be passed between batch queries - must combine at SQL level
	batch := &pgx.Batch{}

	sql := `
		WITH updated AS (
			UPDATE agent_licenses
			SET license_status = $2, updated_at = $3, updated_by = $4
			WHERE license_id = $1 AND deleted_at IS NULL
			RETURNING agent_id
		)
		INSERT INTO agent_audit_logs (agent_id, action_type, field_name, new_value, action_reason, performed_by, performed_at)
		SELECT agent_id, $5, $6, $7, $8, $9, $10
		FROM updated
	`

	args := []interface{}{
		licenseID, domain.LicenseStatusExpired, time.Now(), updatedBy,
		domain.AuditActionLicenseUpdate, "license_status", domain.LicenseStatusExpired, "License expired", updatedBy, time.Now(),
	}

	err := dbutil.QueueExecRowRaw(batch, sql, args...)
	if err != nil {
		return err
	}

	// Execute batch
	return r.db.SendBatch(cCtx, batch).Close()
}

// BatchMarkAsExpired marks multiple licenses as expired
// OPTIMIZATION: UNNEST pattern for bulk expiry processing with CTE
// BR-AGT-PRF-013: Auto-Deactivation on License Expiry
func (r *AgentLicenseRepository) BatchMarkAsExpired(ctx context.Context, licenseIDs []string, updatedBy string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	// Use UNNEST + CTE to bulk update licenses and insert audit logs in single query
	// CRITICAL: Golang variables cannot be passed between batch queries - must use UNNEST pattern
	batch := &pgx.Batch{}

	sql := `
		WITH updated AS (
			UPDATE agent_licenses
			SET license_status = $2, updated_at = $3, updated_by = $4
			WHERE license_id = ANY($1::uuid[]) AND deleted_at IS NULL
			RETURNING agent_id, license_id
		)
		INSERT INTO agent_audit_logs (agent_id, action_type, field_name, new_value, action_reason, performed_by, performed_at)
		SELECT agent_id, $5, $6, $7, $8, $9, $10
		FROM updated
	`

	args := []interface{}{
		licenseIDs, domain.LicenseStatusExpired, time.Now(), updatedBy,
		domain.AuditActionLicenseUpdate, "license_status", domain.LicenseStatusExpired,
		"License expired - batch processing", updatedBy, time.Now(),
	}

	err := dbutil.QueueExecRowRaw(batch, sql, args...)
	if err != nil {
		return err
	}

	// Execute batch
	return r.db.SendBatch(cCtx, batch).Close()
}

// Delete soft deletes an agent license
func (r *AgentLicenseRepository) Delete(ctx context.Context, licenseID, deletedBy string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use CTE to combine UPDATE + INSERT audit in single query
	// CRITICAL: Golang variables cannot be passed between batch queries - must combine at SQL level
	batch := &pgx.Batch{}

	sql := `
		WITH updated AS (
			UPDATE agent_licenses
			SET deleted_at = $2, updated_by = $3
			WHERE license_id = $1 AND deleted_at IS NULL
			RETURNING agent_id
		)
		INSERT INTO agent_audit_logs (agent_id, action_type, field_name, action_reason, performed_by, performed_at)
		SELECT agent_id, $4, $5, $6, $7, $8
		FROM updated
	`

	args := []interface{}{
		licenseID, time.Now(), deletedBy,
		domain.AuditActionDelete, "license", "License deleted", deletedBy, time.Now(),
	}

	err := dbutil.QueueExecRowRaw(batch, sql, args...)
	if err != nil {
		return err
	}

	// Execute batch
	return r.db.SendBatch(cCtx, batch).Close()
}

// ValidateLicenseNumberUniqueness checks if license number is unique
// VR-AGT-PRF-020: License Number Uniqueness
func (r *AgentLicenseRepository) ValidateLicenseNumberUniqueness(ctx context.Context, licenseNumber, excludeLicenseID string) (bool, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("COUNT(*)").
		From(agentLicenseTable).
		Where(sq.Eq{"license_number": licenseNumber, "deleted_at": nil}).
		Where(sq.NotEq{"license_id": excludeLicenseID})

	var count int64
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowTo[int64], &count)
	if err != nil {
		return false, err
	}

	return count == 0, nil
}
