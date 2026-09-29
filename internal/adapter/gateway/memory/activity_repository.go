// internal/adapter/gateway/memory/activity_repository.go
package memory

import (
	"context"

	"github.com/y-sugiyama654/study-clean-architecture/internal/domain"
	"github.com/y-sugiyama654/study-clean-architecture/internal/usecase"
)

// ActivityRepository は操作の履歴をメモリ上に記録する usecase.ActivityRepository の実装
type ActivityRepository struct {
	s *Store
}

var _ usecase.ActivityRepository = (*ActivityRepository)(nil)

func NewActivityRepository(s *Store) *ActivityRepository {
	return &ActivityRepository{s: s}
}

func (r *ActivityRepository) Add(_ context.Context, a domain.Activity) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.activities = append(r.s.activities, a)
	return nil
}
