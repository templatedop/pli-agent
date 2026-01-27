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
)

const (
	licenseReminderTable = "agent_license_reminders"
)

// LicenseReminderRepository handles license reminder operations
type LicenseReminderRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewLicenseReminderRepository creates a new license reminder repository
func NewLicenseReminderRepository(db *dblib.DB, cfg *config.Config) *LicenseReminderRepository {
	return &LicenseReminderRepository{
		db:  db,
		cfg: cfg,
	}
}

// CreateBatch creates multiple reminders for a license
// AGT-030: Add License (triggers reminder creation)
// FR-AGT-PRF-011: License Renewal Automation
// BR-AGT-PRF-014: License Renewal Reminders (30, 15, 7 days, expiry day)
// WF-AGT-PRF-003: License Renewal Workflow
// CRITICAL: Bulk insert of 4 reminders in single database round trip
func (r *LicenseReminderRepository) CreateBatch(ctx context.Context, licenseID string, renewalDate time.Time, createdBy string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Calculate reminder dates
	reminders := []struct {
		Type string
		Date time.Time
	}{
		{domain.ReminderType30Days, renewalDate.AddDate(0, 0, -30)},
		{domain.ReminderType15Days, renewalDate.AddDate(0, 0, -15)},
		{domain.ReminderType7Days, renewalDate.AddDate(0, 0, -7)},
		{domain.ReminderTypeExpiryDay, renewalDate},
	}

	// Build VALUES for bulk insert
	insertQuery := dblib.Psql.Insert(licenseReminderTable).
		Columns("license_id", "reminder_type", "reminder_date", "sent_status", "created_by")

	for _, reminder := range reminders {
		insertQuery = insertQuery.Values(
			licenseID,
			reminder.Type,
			reminder.Date,
			domain.ReminderStatusPending,
			createdBy,
		)
	}

	_, err := dblib.Insert(cCtx, r.db, insertQuery)
	if err != nil {
		return fmt.Errorf("failed to create reminders: %w", err)
	}

	return nil
}

// FindByLicenseID retrieves all reminders for a license
// AGT-037: Get License Reminders
// FR-AGT-PRF-011: License Renewal Automation
// BR-AGT-PRF-014: License Renewal Reminders
// CRITICAL: Single database round trip with ordered results
func (r *LicenseReminderRepository) FindByLicenseID(ctx context.Context, licenseID string) ([]domain.LicenseReminder, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(licenseReminderTable).
		Where(sq.Eq{"license_id": licenseID}).
		OrderBy("reminder_date ASC")

	var reminders []domain.LicenseReminder
	err := dblib.SelectRows(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.LicenseReminder], &reminders)
	if err != nil {
		return nil, fmt.Errorf("failed to find reminders: %w", err)
	}

	return reminders, nil
}

// UpdateSentStatus marks a reminder as sent or failed
// FR-AGT-PRF-011: License Renewal Automation
// BR-AGT-PRF-014: License Renewal Reminders
// WF-AGT-PRF-003: License Renewal Workflow
// CRITICAL: Atomic status update with retry count tracking
func (r *LicenseReminderRepository) UpdateSentStatus(
	ctx context.Context,
	reminderID string,
	status string,
	emailSent, smsSent bool,
	failureReason *string,
) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	updateQuery := dblib.Psql.Update(licenseReminderTable).
		Set("sent_status", status).
		Set("sent_date", time.Now()).
		Set("email_sent", emailSent).
		Set("sms_sent", smsSent).
		Set("retry_count", sq.Expr("retry_count + 1")).
		Where(sq.Eq{"reminder_id": reminderID})

	if failureReason != nil {
		updateQuery = updateQuery.Set("failure_reason", *failureReason)
	}

	tag, err := dblib.Update(cCtx, r.db, updateQuery)
	if err != nil {
		return fmt.Errorf("failed to update reminder status: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("reminder not found: %s", reminderID)
	}

	return nil
}

// FindPendingReminders retrieves pending reminders for today or past due
// Used by background job to send reminders
// FR-AGT-PRF-011: License Renewal Automation
// BR-AGT-PRF-014: License Renewal Reminders
// WF-AGT-PRF-003: License Renewal Workflow
// CRITICAL: Batch query for system job processing
func (r *LicenseReminderRepository) FindPendingReminders(ctx context.Context, upToDate time.Time) ([]domain.LicenseReminder, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(licenseReminderTable).
		Where(sq.And{
			sq.Eq{"sent_status": domain.ReminderStatusPending},
			sq.LtOrEq{"reminder_date": upToDate},
		}).
		OrderBy("reminder_date ASC")

	var reminders []domain.LicenseReminder
	err := dblib.SelectRows(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.LicenseReminder], &reminders)
	if err != nil {
		return nil, fmt.Errorf("failed to find pending reminders: %w", err)
	}

	return reminders, nil
}
