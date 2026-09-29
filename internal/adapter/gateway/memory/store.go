// internal/adapter/gateway/memory/store.go
package memory

import (
	"maps"
	"slices"
	"sync"

	"github.com/y-sugiyama654/study-clean-architecture/internal/domain"
)

// Store はメモリ上のデータ置き場。タスクと操作の履歴のリポジトリで共有する
type Store struct {
	mu         sync.RWMutex
	tasks      map[domain.TaskID]*domain.Task
	activities []domain.Activity

	txMu sync.Mutex // トランザクションを1つずつ順番に実行するためのロック
}

func NewStore() *Store {
	return &Store{tasks: map[domain.TaskID]*domain.Task{}}
}

// snapshot は現在の中身のコピーを返す（ロールバックで元に戻すために使う）
func (s *Store) snapshot() (map[domain.TaskID]*domain.Task, []domain.Activity) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return maps.Clone(s.tasks), slices.Clone(s.activities)
}

// restore はsnapshotで取っておいた中身に戻す
func (s *Store) restore(tasks map[domain.TaskID]*domain.Task, activities []domain.Activity) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks, s.activities = tasks, activities
}

// Activities は記録された操作の履歴のコピーを返す（テストや確認用）
func (s *Store) Activities() []domain.Activity {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return slices.Clone(s.activities)
}
