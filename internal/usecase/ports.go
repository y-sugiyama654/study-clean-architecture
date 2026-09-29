// internal/usecase/ports.go
package usecase

import (
	"context"
	"time"

	"github.com/y-sugiyama654/study-clean-architecture/internal/domain"
)

// ユースケースが外の世界に求める機能を、インターフェースとしてここ（使う側）で定義する。
// 実装（PostgreSQLへの保存、現在時刻の取得など）は外側の層が用意する

// TaskRepository はタスクの保存と取得を行う
type TaskRepository interface {
	// Save はタスクを保存する（新規なら追加、既存なら更新）
	Save(ctx context.Context, task *domain.Task) error
	// FindByID はIDでタスクを探す。見つからなければ ErrTaskNotFound を返す
	FindByID(ctx context.Context, id domain.TaskID) (*domain.Task, error)
	// ListByOwner は所有者のタスクを作成日時の古い順に返す
	ListByOwner(ctx context.Context, ownerID domain.UserID) ([]*domain.Task, error)
}

// Clock は現在時刻を返す
type Clock interface {
	Now() time.Time
}

// IDGenerator は新しいタスクIDを払い出す
type IDGenerator interface {
	NewTaskID() domain.TaskID
}
