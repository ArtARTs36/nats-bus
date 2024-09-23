package eventhistory

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
)

type EventRepository interface {
	Exists(ctx context.Context, ev *Event) (bool, error)
	Create(ctx context.Context, ev *Event) error
}

type DBEventRepository struct {
	db        *sqlx.DB
	tableName string
}

func NewDBEventRepository(db *sqlx.DB, tableName string) *DBEventRepository {
	return &DBEventRepository{
		db:        db,
		tableName: tableName,
	}
}

func (r *DBEventRepository) Exists(ctx context.Context, ev *Event) (bool, error) {
	q, _, err := goqu.Select("id").Where(goqu.C("id").Eq(ev.ID)).ToSQL()
	if err != nil {
		return false, fmt.Errorf("failed to build query: %w", err)
	}

	type row struct {
		ID string `db:"id"`
	}

	err = r.db.SelectContext(ctx, &row{}, q)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, fmt.Errorf("failed to execute query: %w", err)
	}

	return true, err
}

func (r *DBEventRepository) Create(ctx context.Context, ev *Event) error {
	q, _, err := goqu.
		Insert(r.tableName).
		Rows(ev).
		OnConflict(goqu.DoNothing()).
		ToSQL()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, q)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}
