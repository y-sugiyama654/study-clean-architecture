// internal/usecase/list_tasks.go
package usecase

import (
	"context"

	"github.com/y-sugiyama654/study-clean-architecture/internal/domain"
)

// ListTasksInput はタスク一覧取得の入力
type ListTasksInput struct {
	UserID string
}

// ListTasks は自分のタスクの一覧を取得するユースケース
type ListTasks struct {
	repo TaskRepository
}

func NewListTasks(repo TaskRepository) *ListTasks {
	return &ListTasks{repo: repo}
}

func (uc *ListTasks) Execute(ctx context.Context, in ListTasksInput) ([]TaskOutput, error) {
	ownerID, err := domain.NewUserID(in.UserID)
	if err != nil {
		return nil, err
	}
	tasks, err := uc.repo.ListByOwner(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	out := make([]TaskOutput, len(tasks))
	for i, t := range tasks {
		out[i] = newTaskOutput(t)
	}
	return out, nil
}
