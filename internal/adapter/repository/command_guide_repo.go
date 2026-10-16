package repository

import (
	"context"
	"database/sql"
	"fmt"

	"McQueens_Tea_Cup/internal/config"
	"McQueens_Tea_Cup/internal/domain/entity"
	"McQueens_Tea_Cup/internal/domain/port"
)

type CommandGuideRepo struct {
	DB        *sql.DB
	tableName string
}

func NewCommandGuideRepository(db *sql.DB, tables config.DatabaseTablesConfig) port.CommandGuideRepository {
	return &CommandGuideRepo{DB: db, tableName: quoteIdent(tables.BotCommandGuides)}
}

// commandGuideColumns is the explicit column list (in scan order) shared by the
// read queries so an added/reordered column can't silently corrupt a Scan.
const commandGuideColumns = `id, command_key, title, content, type, sort_order, enabled, created_at, updated_at`

func scanCommandGuide(s interface{ Scan(...any) error }) (entity.CommandGuide, error) {
	var g entity.CommandGuide
	err := s.Scan(&g.ID, &g.CommandKey, &g.Title, &g.Content, &g.Type, &g.SortOrder, &g.Enabled, &g.CreatedAt, &g.UpdatedAt)
	return g, err
}

// GetByCommandKey returns the enabled guide for a command key, or (nil, nil) when
// none exists, so callers can show a friendly "no guide yet" message.
func (r *CommandGuideRepo) GetByCommandKey(ctx context.Context, commandKey string) (*entity.CommandGuide, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM %s WHERE command_key = $1 AND enabled = TRUE`,
		commandGuideColumns, r.tableName,
	)
	g, err := scanCommandGuide(r.DB.QueryRowContext(ctx, query, commandKey))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &g, nil
}

// ListEnabled returns all enabled guides ordered for display (sort_order, then
// title).
func (r *CommandGuideRepo) ListEnabled(ctx context.Context) ([]entity.CommandGuide, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM %s WHERE enabled = TRUE ORDER BY sort_order ASC, title ASC`,
		commandGuideColumns, r.tableName,
	)
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var guides []entity.CommandGuide
	for rows.Next() {
		g, err := scanCommandGuide(rows)
		if err != nil {
			return nil, err
		}
		guides = append(guides, g)
	}
	return guides, rows.Err()
}
