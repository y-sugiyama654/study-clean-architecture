// internal/usecase/output.go
package usecase

import (
	"time"

	"github.com/y-sugiyama654/study-clean-architecture/internal/domain"
)

// TaskOutput はユースケースが返すタスクの情報
// エンティティをそのまま外に渡さず、外側の層が必要とする値だけを詰め直す
type TaskOutput struct {
	ID          string
	OwnerID     string
	Title       string
	Description string
	Status      string
	CreatedAt   time.Time
	CompletedAt *time.Time
}

func newTaskOutput(t *domain.Task) TaskOutput {
	return TaskOutput{
		ID:          t.ID().String(),
		OwnerID:     t.OwnerID().String(),
		Title:       t.Title().String(),
		Description: t.Description().String(),
		Status:      t.Status().String(),
		CreatedAt:   t.CreatedAt(),
		CompletedAt: t.CompletedAt(),
	}
}
