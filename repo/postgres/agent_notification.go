package repo

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"

	"pli-agent-api/core/domain"
)

// AgentNotificationRepository handles all database operations for agent notifications
// Phase 9: Search & Dashboard APIs
// AGT-077: Agent Notification History
// FR-AGT-PRF-021: Self-Service Update notifications
type AgentNotificationRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewAgentNotificationRepository creates a new agent notification repository
func NewAgentNotificationRepository(db *dblib.DB, cfg *config.Config) *AgentNotificationRepository {
	return &AgentNotificationRepository{
		db:  db,
		cfg: cfg,
	}
}

const agentNotificationTable = "agent_notifications"

// Create inserts a new notification record
// Used by workflow activities to record notifications sent
func (r *AgentNotificationRepository) Create(ctx context.Context, notification *domain.AgentNotification) (*domain.AgentNotification, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Insert(agentNotificationTable).
		Columns(
			"agent_id", "notification_type", "template", "recipient",
			"subject", "message", "sent_at", "status", "metadata",
		).
		Values(
			notification.AgentID, notification.NotificationType, notification.Template,
			notification.Recipient, notification.Subject, notification.Message,
			notification.SentAt, notification.Status, notification.Metadata,
		).
		Suffix("RETURNING *")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var result domain.AgentNotification
	err = r.db.Get(cCtx, &result, sql, args...)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetByAgentID fetches notifications by agent ID with pagination and filters
// AGT-077: Agent Notification History
// SINGLE database query with all filters
func (r *AgentNotificationRepository) GetByAgentID(
	ctx context.Context,
	agentID string,
	notificationType *string,
	fromDate *time.Time,
	toDate *time.Time,
	page int,
	limit int,
) ([]domain.AgentNotification, *domain.PaginationMetadata, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMedium"))
	defer cancel()

	// Build base query with filters
	query := dblib.Psql.Select("*").
		From(agentNotificationTable).
		Where(sq.Eq{"agent_id": agentID})

	// Apply optional filters
	if notificationType != nil {
		query = query.Where(sq.Eq{"notification_type": *notificationType})
	}
	if fromDate != nil {
		query = query.Where(sq.GtOrEq{"sent_at": *fromDate})
	}
	if toDate != nil {
		query = query.Where(sq.LtOrEq{"sent_at": *toDate})
	}

	// Count total results for pagination
	countQuery := query
	countSQL, countArgs, err := countQuery.ToSql()
	if err != nil {
		return nil, nil, err
	}

	var totalResults int
	countSQLWithCount := "SELECT COUNT(*) FROM (" + countSQL + ") AS count_query"
	err = r.db.Get(cCtx, &totalResults, countSQLWithCount, countArgs...)
	if err != nil {
		return nil, nil, err
	}

	// Apply pagination and ordering
	offset := (page - 1) * limit
	query = query.OrderBy("sent_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, nil, err
	}

	var notifications []domain.AgentNotification
	err = r.db.Select(cCtx, &notifications, sql, args...)
	if err != nil {
		return nil, nil, err
	}

	// Calculate pagination metadata
	totalPages := (totalResults + limit - 1) / limit
	pagination := &domain.PaginationMetadata{
		CurrentPage:    page,
		TotalPages:     totalPages,
		TotalResults:   totalResults,
		ResultsPerPage: limit,
	}

	return notifications, pagination, nil
}

// GetRecentByAgentID fetches recent notifications for dashboard
// AGT-068: Agent Dashboard (recent notifications section)
// Returns last N notifications without pagination
func (r *AgentNotificationRepository) GetRecentByAgentID(ctx context.Context, agentID string, limit int) ([]domain.AgentNotification, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentNotificationTable).
		Where(sq.Eq{"agent_id": agentID}).
		OrderBy("sent_at DESC").
		Limit(uint64(limit))

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var notifications []domain.AgentNotification
	err = r.db.Select(cCtx, &notifications, sql, args...)
	if err != nil {
		return nil, err
	}

	return notifications, nil
}

// UpdateStatus updates notification status (for delivery tracking)
// Used when we receive delivery confirmation from notification service
func (r *AgentNotificationRepository) UpdateStatus(
	ctx context.Context,
	notificationID string,
	status string,
	deliveredAt *time.Time,
) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Update(agentNotificationTable).
		Set("status", status)

	if deliveredAt != nil {
		query = query.Set("delivered_at", *deliveredAt)
	}

	query = query.Where(sq.Eq{"notification_id": notificationID})

	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(cCtx, sql, args...)
	return err
}

// MarkAsRead marks a notification as read (for internal notifications)
// Used by agent portal when agent views notification
func (r *AgentNotificationRepository) MarkAsRead(ctx context.Context, notificationID string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Update(agentNotificationTable).
		Set("status", domain.NotificationStatusRead).
		Set("read_at", time.Now()).
		Where(sq.Eq{"notification_id": notificationID})

	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(cCtx, sql, args...)
	return err
}
