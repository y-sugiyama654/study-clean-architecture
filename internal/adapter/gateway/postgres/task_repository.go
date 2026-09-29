// internal/adapter/gateway/postgres/task_repository.go
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/gateway/postgres/sqlcgen"
	"github.com/y-sugiyama654/study-clean-architecture/internal/domain"
	"github.com/y-sugiyama654/study-clean-architecture/internal/usecase"
)

// TaskRepository はタスクをPostgreSQLに保存する usecase.TaskRepository の実装
type TaskRepository struct {
	q *sqlcgen.Queries
}

var _ usecase.TaskRepository = (*TaskRepository)(nil)

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{q: sqlcgen.New(db)}
}

func (r *TaskRepository) Save(ctx context.Context, task *domain.Task) error {
	err := r.q.UpsertTask(ctx, sqlcgen.UpsertTaskParams{
		ID:          task.ID().String(),
		OwnerID:     task.OwnerID().String(),
		Title:       task.Title().String(),
		Description: task.Description().String(),
		Status:      task.Status().String(),
		CreatedAt:   task.CreatedAt(),
		CompletedAt: toNullTime(task.CompletedAt()),
	})
	if err != nil {
		return fmt.Errorf("postgres: タスクを保存できません: %w", err)
	}
	return nil
}

func (r *TaskRepository) FindByID(ctx context.Context, id domain.TaskID) (*domain.Task, error) {
	row, err := r.q.GetTask(ctx, id.String())
	if errors.Is(err, sql.ErrNoRows) {
		return nil, usecase.ErrTaskNotFound // データベースの都合のエラーを、ユースケースが決めたエラーに変換する
	}
	if err != nil {
		return nil, fmt.Errorf("postgres: タスクを取得できません: %w", err)
	}
	return toEntity(row)
}

func (r *TaskRepository) ListByOwner(ctx context.Context, ownerID domain.UserID) ([]*domain.Task, error) {
	rows, err := r.q.ListTasksByOwner(ctx, ownerID.String())
	if err != nil {
		return nil, fmt.Errorf("postgres: タスクの一覧を取得できません: %w", err)
	}
	tasks := make([]*domain.Task, 0, len(rows))
	for _, row := range rows {
		t, err := toEntity(row)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

// toEntity はテーブルの1行をエンティティに変換する
// 値オブジェクトのコンストラクタを通すので、万一データベースに不正な値があれば気づける
func toEntity(row sqlcgen.Task) (*domain.Task, error) {
	title, err := domain.NewTitle(row.Title)
	if err != nil {
		return nil, fmt.Errorf("postgres: タスク %s のタイトルが不正です: %w", row.ID, err)
	}
	desc, err := domain.NewDescription(row.Description)
	if err != nil {
		return nil, fmt.Errorf("postgres: タスク %s の説明が不正です: %w", row.ID, err)
	}
	status, err := domain.ParseStatus(row.Status)
	if err != nil {
		return nil, fmt.Errorf("postgres: タスク %s のステータスが不正です: %w", row.ID, err)
	}
	var completedAt *time.Time
	if row.CompletedAt.Valid {
		completedAt = &row.CompletedAt.Time
	}
	return domain.ReconstructTask(domain.TaskID(row.ID), domain.UserID(row.OwnerID),
		title, desc, status, row.CreatedAt, completedAt), nil
}

func toNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}
