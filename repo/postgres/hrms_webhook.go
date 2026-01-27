package repo

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"

	"pli-agent-api/core/domain"
)

// HRMSWebhookRepository handles all database operations for HRMS webhook events
// Phase 10: Batch & Webhook APIs
// AGT-078: HRMS Webhook Receiver
// INT-AGT-001: HRMS System Integration
type HRMSWebhookRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewHRMSWebhookRepository creates a new HRMS webhook repository
func NewHRMSWebhookRepository(db *dblib.DB, cfg *config.Config) *HRMSWebhookRepository {
	return &HRMSWebhookRepository{
		db:  db,
		cfg: cfg,
	}
}

const hrmsWebhookEventTable = "hrms_webhook_events"

// CreateEvent creates a new webhook event record
// AGT-078: HRMS Webhook Receiver
func (r *HRMSWebhookRepository) CreateEvent(ctx context.Context, event *domain.HRMSWebhookEvent) (*domain.HRMSWebhookEvent, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Insert(hrmsWebhookEventTable).
		Columns(
			"event_id", "event_type", "employee_id", "employee_data",
			"signature", "signature_valid", "status",
		).
		Values(
			event.EventID, event.EventType, event.EmployeeID,
			event.EmployeeData, event.Signature, event.SignatureValid,
			event.Status,
		).
		Suffix("RETURNING *")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var result domain.HRMSWebhookEvent
	err = r.db.Get(cCtx, &result, sql, args...)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateEventStatus updates webhook event processing status
// Used after webhook processing completes
func (r *HRMSWebhookRepository) UpdateEventStatus(
	ctx context.Context,
	eventID string,
	status string,
	processingResult, errorMessage *string,
) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Update(hrmsWebhookEventTable).
		Set("status", status)

	if status == domain.WebhookStatusProcessed || status == domain.WebhookStatusFailed {
		query = query.Set("processed_at", sq.Expr("NOW()"))
	}
	if processingResult != nil {
		query = query.Set("processing_result", *processingResult)
	}
	if errorMessage != nil {
		query = query.Set("error_message", *errorMessage)
	}

	query = query.Where(sq.Eq{"event_id": eventID})

	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(cCtx, sql, args...)
	return err
}

// GetEventByID retrieves a webhook event by ID
func (r *HRMSWebhookRepository) GetEventByID(ctx context.Context, eventID string) (*domain.HRMSWebhookEvent, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(hrmsWebhookEventTable).
		Where(sq.Eq{"event_id": eventID})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var event domain.HRMSWebhookEvent
	err = r.db.Get(cCtx, &event, sql, args...)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

// GetEventsByEmployeeID retrieves webhook events for an employee
func (r *HRMSWebhookRepository) GetEventsByEmployeeID(ctx context.Context, employeeID string) ([]domain.HRMSWebhookEvent, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(hrmsWebhookEventTable).
		Where(sq.Eq{"employee_id": employeeID}).
		OrderBy("received_at DESC")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var events []domain.HRMSWebhookEvent
	err = r.db.Select(cCtx, &events, sql, args...)
	if err != nil {
		return nil, err
	}

	return events, nil
}

// GetPendingEvents retrieves events pending processing for retry
func (r *HRMSWebhookRepository) GetPendingEvents(ctx context.Context, limit int) ([]domain.HRMSWebhookEvent, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(hrmsWebhookEventTable).
		Where(sq.Or{
			sq.Eq{"status": domain.WebhookStatusReceived},
			sq.And{
				sq.Eq{"status": domain.WebhookStatusRetrying},
				sq.LtOrEq{"next_retry_at": sq.Expr("NOW()")},
			},
		}).
		OrderBy("received_at ASC").
		Limit(uint64(limit))

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var events []domain.HRMSWebhookEvent
	err = r.db.Select(cCtx, &events, sql, args...)
	if err != nil {
		return nil, err
	}

	return events, nil
}

// IncrementRetryCount increments the retry count for a webhook event
func (r *HRMSWebhookRepository) IncrementRetryCount(ctx context.Context, eventID string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	sql := `
		UPDATE hrms_webhook_events
		SET
			retry_count = retry_count + 1,
			next_retry_at = NOW() + (POWER(2, retry_count + 1) || ' minutes')::INTERVAL,
			status = $1
		WHERE event_id = $2
	`

	_, err := r.db.Exec(cCtx, sql, domain.WebhookStatusRetrying, eventID)
	return err
}

// GetFailedEvents retrieves events that have failed processing
func (r *HRMSWebhookRepository) GetFailedEvents(ctx context.Context, limit int) ([]domain.HRMSWebhookEvent, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(hrmsWebhookEventTable).
		Where(sq.Eq{"status": domain.WebhookStatusFailed}).
		OrderBy("received_at DESC").
		Limit(uint64(limit))

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var events []domain.HRMSWebhookEvent
	err = r.db.Select(cCtx, &events, sql, args...)
	if err != nil {
		return nil, err
	}

	return events, nil
}
