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

	// Use batch for license creation with audit log in single transaction
	// OPTIMIZATION: Batch operation combines INSERT license + INSERT audit log
	batch := &pgx.Batch{}

	// Query 1: Insert agent license
	// BR-AGT-PRF-012: Provisional license valid for 1 year, renewable max 2 times
	// BR-AGT-PRF-030: Track license_date, renewal_date, authority_date
	query1 := dblib.Psql.Insert(agentLicenseTable).
		Columns(
			"agent_id", "license_line", "license_type", "license_number", "resident_status",
			"license_date", "renewal_date", "authority_date", "renewal_count", "license_status",
			"licentiate_exam_passed", "licentiate_exam_date", "licentiate_certificate_number",
			"is_primary", "metadata", "created_by",
		).
		Values(
			license.AgentID, license.LicenseLine, license.LicenseType, license.LicenseNumber,
			license.ResidentStatus, license.LicenseDate, license.RenewalDate, license.AuthorityDate,
			license.RenewalCount, license.LicenseStatus, license.LicentiateExamPassed,
			license.LicentiateExamDate, license.LicentiateCertificateNumber, license.IsPrimary,
			license.Metadata, license.CreatedBy,
		).
		Suffix("RETURNING license_id, created_at, version")

	var result domain.AgentLicense
	err := dblib.QueueReturnRow(batch, query1, pgx.RowToStructByNameLax[domain.AgentLicense], &result)
	if err != nil {
		return nil, err
	}

	// Query 2: Insert audit log for license creation
	// BR-AGT-PRF-005: Audit Logging
	query2 := dblib.Psql.Insert("agent_audit_logs").
		Columns("agent_id", "action_type", "field_name", "new_value", "action_reason", "performed_by", "performed_at").
		Values(license.AgentID, domain.AuditActionLicenseAdd, "license_number", license.LicenseNumber, "New license added", license.CreatedBy, time.Now())

	err = dblib.QueueExecRow(batch, query2)
	if err != nil {
		return nil, err
	}

	// Execute batch
	err = r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, err
	}

	// Copy input data to result
	result = license
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

// Update updates an agent license
// FR-AGT-PRF-014: License Management
// BR-AGT-PRF-012: License Renewal Period Rules
func (r *AgentLicenseRepository) Update(ctx context.Context, licenseID string, updates map[string]interface{}, updatedBy string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use batch for update + audit log
	// OPTIMIZATION: Batch combines UPDATE + INSERT audit
	batch := &pgx.Batch{}

	// Query 1: Update license
	updateQuery := dblib.Psql.Update(agentLicenseTable).
		Set("updated_at", time.Now()).
		Set("updated_by", updatedBy).
		Where(sq.Eq{"license_id": licenseID, "deleted_at": nil})

	// Apply updates
	for field, value := range updates {
		updateQuery = updateQuery.Set(field, value)
	}

	err := dblib.QueueExecRow(batch, updateQuery)
	if err != nil {
		return err
	}

	// Query 2: Get agent_id for audit log
	var agentID string
	selectQuery := dblib.Psql.Select("agent_id").
		From(agentLicenseTable).
		Where(sq.Eq{"license_id": licenseID})

	err = dblib.QueueReturnRow(batch, selectQuery, pgx.RowTo[string], &agentID)
	if err != nil {
		return err
	}

	// Query 3: Insert audit logs for each field update
	// BR-AGT-PRF-005: Audit Logging
	for field, newValue := range updates {
		auditQuery := dblib.Psql.Insert("agent_audit_logs").
			Columns("agent_id", "action_type", "field_name", "new_value", "performed_by", "performed_at").
			Values(agentID, domain.AuditActionLicenseUpdate, field, newValue, updatedBy, time.Now())

		err = dblib.QueueExecRow(batch, auditQuery)
		if err != nil {
			return err
		}
	}

	// Execute batch
	return r.db.SendBatch(cCtx, batch).Close()
}

// RenewLicense renews a license and updates renewal count
// BR-AGT-PRF-012: License Renewal Period Rules
// Provisional: Max 2 renewals, 1 year each
// Permanent: Renewable every 1 year after 5-year validity
func (r *AgentLicenseRepository) RenewLicense(ctx context.Context, licenseID, updatedBy string, newRenewalDate time.Time) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use batch for renewal update + audit log
	// OPTIMIZATION: Batch combines UPDATE + INSERT audit
	batch := &pgx.Batch{}

	// Query 1: Increment renewal count and update renewal date
	updateQuery := dblib.Psql.Update(agentLicenseTable).
		Set("renewal_count", sq.Expr("renewal_count + 1")).
		Set("renewal_date", newRenewalDate).
		Set("license_status", domain.LicenseStatusRenewed).
		Set("updated_at", time.Now()).
		Set("updated_by", updatedBy).
		Where(sq.Eq{"license_id": licenseID, "deleted_at": nil})

	err := dblib.QueueExecRow(batch, updateQuery)
	if err != nil {
		return err
	}

	// Query 2: Get agent_id for audit log
	var agentID string
	selectQuery := dblib.Psql.Select("agent_id").
		From(agentLicenseTable).
		Where(sq.Eq{"license_id": licenseID})

	err = dblib.QueueReturnRow(batch, selectQuery, pgx.RowTo[string], &agentID)
	if err != nil {
		return err
	}

	// Query 3: Insert audit log
	auditQuery := dblib.Psql.Insert("agent_audit_logs").
		Columns("agent_id", "action_type", "field_name", "new_value", "action_reason", "performed_by", "performed_at").
		Values(agentID, domain.AuditActionLicenseUpdate, "renewal_date", newRenewalDate, "License renewed", updatedBy, time.Now())

	err = dblib.QueueExecRow(batch, auditQuery)
	if err != nil {
		return err
	}

	// Execute batch
	return r.db.SendBatch(cCtx, batch).Close()
}

// ConvertToPermanent converts a provisional license to permanent after passing licentiate exam
// BR-AGT-PRF-012: License Renewal Period Rules
// After passing exam within 3 years: Permanent license with 5-year validity, renewable every 1 year
func (r *AgentLicenseRepository) ConvertToPermanent(ctx context.Context, licenseID, updatedBy string, examDate time.Time, certificateNumber string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use batch for conversion update + audit log
	// OPTIMIZATION: Batch combines UPDATE + INSERT audit
	batch := &pgx.Batch{}

	// Calculate new renewal date: 5 years from conversion, then renewable every 1 year
	permanentValidityDate := examDate.AddDate(5, 0, 0)

	// Query 1: Convert to permanent license
	updateQuery := dblib.Psql.Update(agentLicenseTable).
		Set("license_type", domain.LicenseTypePermanent).
		Set("licentiate_exam_passed", true).
		Set("licentiate_exam_date", examDate).
		Set("licentiate_certificate_number", certificateNumber).
		Set("renewal_date", permanentValidityDate).
		Set("license_status", domain.LicenseStatusActive).
		Set("updated_at", time.Now()).
		Set("updated_by", updatedBy).
		Where(sq.Eq{"license_id": licenseID, "deleted_at": nil})

	err := dblib.QueueExecRow(batch, updateQuery)
	if err != nil {
		return err
	}

	// Query 2: Get agent_id for audit log
	var agentID string
	selectQuery := dblib.Psql.Select("agent_id").
		From(agentLicenseTable).
		Where(sq.Eq{"license_id": licenseID})

	err = dblib.QueueReturnRow(batch, selectQuery, pgx.RowTo[string], &agentID)
	if err != nil {
		return err
	}

	// Query 3: Insert audit log
	auditQuery := dblib.Psql.Insert("agent_audit_logs").
		Columns("agent_id", "action_type", "field_name", "new_value", "action_reason", "performed_by", "performed_at").
		Values(agentID, domain.AuditActionLicenseUpdate, "license_type", domain.LicenseTypePermanent, "Converted to permanent after passing licentiate exam", updatedBy, time.Now())

	err = dblib.QueueExecRow(batch, auditQuery)
	if err != nil {
		return err
	}

	// Execute batch
	return r.db.SendBatch(cCtx, batch).Close()
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

	// Use batch for status update + audit log
	// OPTIMIZATION: Batch combines UPDATE + INSERT audit
	batch := &pgx.Batch{}

	// Query 1: Mark license as expired
	updateQuery := dblib.Psql.Update(agentLicenseTable).
		Set("license_status", domain.LicenseStatusExpired).
		Set("updated_at", time.Now()).
		Set("updated_by", updatedBy).
		Where(sq.Eq{"license_id": licenseID, "deleted_at": nil})

	err := dblib.QueueExecRow(batch, updateQuery)
	if err != nil {
		return err
	}

	// Query 2: Get agent_id for audit log
	var agentID string
	selectQuery := dblib.Psql.Select("agent_id").
		From(agentLicenseTable).
		Where(sq.Eq{"license_id": licenseID})

	err = dblib.QueueReturnRow(batch, selectQuery, pgx.RowTo[string], &agentID)
	if err != nil {
		return err
	}

	// Query 3: Insert audit log
	auditQuery := dblib.Psql.Insert("agent_audit_logs").
		Columns("agent_id", "action_type", "field_name", "new_value", "action_reason", "performed_by", "performed_at").
		Values(agentID, domain.AuditActionLicenseUpdate, "license_status", domain.LicenseStatusExpired, "License expired", updatedBy, time.Now())

	err = dblib.QueueExecRow(batch, auditQuery)
	if err != nil {
		return err
	}

	// Execute batch
	return r.db.SendBatch(cCtx, batch).Close()
}

// BatchMarkAsExpired marks multiple licenses as expired
// OPTIMIZATION: Batch operation for bulk expiry processing
// BR-AGT-PRF-013: Auto-Deactivation on License Expiry
func (r *AgentLicenseRepository) BatchMarkAsExpired(ctx context.Context, licenseIDs []string, updatedBy string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	// Use batch for multiple license expiry updates
	// OPTIMIZATION: Batch operation combines multiple UPDATEs in single round-trip
	batch := &pgx.Batch{}

	for _, licenseID := range licenseIDs {
		// Update license status
		updateQuery := dblib.Psql.Update(agentLicenseTable).
			Set("license_status", domain.LicenseStatusExpired).
			Set("updated_at", time.Now()).
			Set("updated_by", updatedBy).
			Where(sq.Eq{"license_id": licenseID, "deleted_at": nil})

		err := dblib.QueueExecRow(batch, updateQuery)
		if err != nil {
			return err
		}

		// Get agent_id for audit log
		var agentID string
		selectQuery := dblib.Psql.Select("agent_id").
			From(agentLicenseTable).
			Where(sq.Eq{"license_id": licenseID})

		err = dblib.QueueReturnRow(batch, selectQuery, pgx.RowTo[string], &agentID)
		if err != nil {
			return err
		}

		// Insert audit log
		auditQuery := dblib.Psql.Insert("agent_audit_logs").
			Columns("agent_id", "action_type", "field_name", "new_value", "action_reason", "performed_by", "performed_at").
			Values(agentID, domain.AuditActionLicenseUpdate, "license_status", domain.LicenseStatusExpired, "License expired - batch processing", updatedBy, time.Now())

		err = dblib.QueueExecRow(batch, auditQuery)
		if err != nil {
			return err
		}
	}

	// Execute batch
	return r.db.SendBatch(cCtx, batch).Close()
}

// Delete soft deletes an agent license
func (r *AgentLicenseRepository) Delete(ctx context.Context, licenseID, deletedBy string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use batch for delete + audit log
	// OPTIMIZATION: Batch combines UPDATE + INSERT audit
	batch := &pgx.Batch{}

	// Query 1: Soft delete license
	updateQuery := dblib.Psql.Update(agentLicenseTable).
		Set("deleted_at", time.Now()).
		Set("updated_by", deletedBy).
		Where(sq.Eq{"license_id": licenseID, "deleted_at": nil})

	err := dblib.QueueExecRow(batch, updateQuery)
	if err != nil {
		return err
	}

	// Query 2: Get agent_id for audit log
	var agentID string
	selectQuery := dblib.Psql.Select("agent_id").
		From(agentLicenseTable).
		Where(sq.Eq{"license_id": licenseID})

	err = dblib.QueueReturnRow(batch, selectQuery, pgx.RowTo[string], &agentID)
	if err != nil {
		return err
	}

	// Query 3: Insert audit log
	auditQuery := dblib.Psql.Insert("agent_audit_logs").
		Columns("agent_id", "action_type", "field_name", "action_reason", "performed_by", "performed_at").
		Values(agentID, domain.AuditActionDelete, "license", "License deleted", deletedBy, time.Now())

	err = dblib.QueueExecRow(batch, auditQuery)
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
