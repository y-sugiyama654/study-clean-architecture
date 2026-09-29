// internal/adapter/gateway/memory/transactor.go
package memory

import (
	"context"

	"github.com/y-sugiyama654/study-clean-architecture/internal/usecase"
)

type txKey struct{}

// Transactor は usecase.Transactor のメモリ実装
// トランザクションを1つずつ順番に実行し、fnがエラーを返したら開始前の中身に戻す
type Transactor struct {
	s *Store
}

var _ usecase.Transactor = (*Transactor)(nil)

func NewTransactor(s *Store) *Transactor {
	return &Transactor{s: s}
}

func (t *Transactor) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if ctx.Value(txKey{}) != nil {
		return fn(ctx) // すでにトランザクションの中
	}
	t.s.txMu.Lock()
	defer t.s.txMu.Unlock()

	tasks, activities := t.s.snapshot()
	if err := fn(context.WithValue(ctx, txKey{}, true)); err != nil {
		t.s.restore(tasks, activities)
		return err
	}
	return nil
}
