package eventhistory

import (
	"context"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
)

type EventUpdateRepository interface {
	Create(ctx context.Context, ev *EventUpdate) error
}

type DBEventUpdateRepository struct {
	db        *sqlx.DB
	tableName string
}

func NewDBEventUpdateRepository(db *sqlx.DB, tableName string) *DBEventUpdateRepository {
	return &DBEventUpdateRepository{
		db:        db,
		tableName: tableName,
	}
}

func (r *DBEventUpdateRepository) Create(ctx context.Context, ev *EventUpdate) error {
	q, _, err := goqu.Insert(r.tableName).Rows(ev).ToSQL()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, q)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}
