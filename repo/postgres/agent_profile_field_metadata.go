package repo

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"

	"pli-agent-api/core/domain"
)

const agentProfileFieldMetadataTable = "agent_profile_field_metadata"

// AgentProfileFieldMetadataRepository handles field metadata operations
type AgentProfileFieldMetadataRepository struct {
	db  dblib.DB
	cfg *config.Config
}

// NewAgentProfileFieldMetadataRepository creates a new field metadata repository
func NewAgentProfileFieldMetadataRepository(db dblib.DB, cfg *config.Config) *AgentProfileFieldMetadataRepository {
	return &AgentProfileFieldMetadataRepository{
		db:  db,
		cfg: cfg,
	}
}

// GetBySection retrieves all field metadata for a given section
// Phase 6.2: AGT-024 - Get Update Form with dynamic metadata
func (r *AgentProfileFieldMetadataRepository) GetBySection(
	ctx context.Context,
	section string,
) ([]domain.AgentProfileFieldMetadata, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentProfileFieldMetadataTable).
		Where(sq.And{
			sq.Eq{"section": section},
			sq.Eq{"is_active": true},
		}).
		OrderBy("display_order ASC, field_name ASC")

	var results []domain.AgentProfileFieldMetadata
	err := dblib.SelectRows(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentProfileFieldMetadata], &results)
	if err != nil {
		return nil, fmt.Errorf("failed to get field metadata for section %s: %w", section, err)
	}

	return results, nil
}

// GetAll retrieves all active field metadata
// For admin or bulk operations
func (r *AgentProfileFieldMetadataRepository) GetAll(ctx context.Context) ([]domain.AgentProfileFieldMetadata, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentProfileFieldMetadataTable).
		Where(sq.Eq{"is_active": true}).
		OrderBy("section ASC, display_order ASC, field_name ASC")

	var results []domain.AgentProfileFieldMetadata
	err := dblib.SelectRows(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentProfileFieldMetadata], &results)
	if err != nil {
		return nil, fmt.Errorf("failed to get all field metadata: %w", err)
	}

	return results, nil
}

// GetCriticalFields retrieves all fields that require approval
// Used for validation logic
func (r *AgentProfileFieldMetadataRepository) GetCriticalFields(ctx context.Context) ([]domain.AgentProfileFieldMetadata, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentProfileFieldMetadataTable).
		Where(sq.And{
			sq.Eq{"requires_approval": true},
			sq.Eq{"is_active": true},
		}).
		OrderBy("section ASC, display_order ASC")

	var results []domain.AgentProfileFieldMetadata
	err := dblib.SelectRows(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentProfileFieldMetadata], &results)
	if err != nil {
		return nil, fmt.Errorf("failed to get critical fields: %w", err)
	}

	return results, nil
}
