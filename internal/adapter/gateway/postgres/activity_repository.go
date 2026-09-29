// internal/adapter/gateway/postgres/activity_repository.go
package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/gateway/postgres/sqlcgen"
	"github.com/y-sugiyama654/study-clean-architecture/internal/domain"
	"github.com/y-sugiyama654/study-clean-architecture/internal/usecase"
)

// ActivityRepository は操作の履歴をPostgreSQLに記録する usecase.ActivityRepository の実装
type ActivityRepository struct {
	db *sql.DB
}

var _ usecase.ActivityRepository = (*ActivityRepository)(nil)

func NewActivityRepository(db *sql.DB) *ActivityRepository {
	return &ActivityRepository{db: db}
}

func (r *ActivityRepository) Add(ctx context.Context, a domain.Activity) error {
	err := queries(ctx, r.db).InsertActivity(ctx, sqlcgen.InsertActivityParams{
		TaskID:     a.TaskID.String(),
		ActorID:    a.ActorID.String(),
		Action:     string(a.Action),
		OccurredAt: a.OccurredAt,
	})
	if err != nil {
		return fmt.Errorf("postgres: 操作の履歴を記録できません: %w", err)
	}
	return nil
}
