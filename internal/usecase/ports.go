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
	// FindByIDForUpdate は、読んだタスクを更新するつもりで探す。
	// トランザクションの中で呼ぶと、そのトランザクションが終わるまで、同じタスクを更新しようとする他の処理を待たせる
	FindByIDForUpdate(ctx context.Context, id domain.TaskID) (*domain.Task, error)
	// ListByOwner は所有者のタスクを作成日時の古い順に返す
	ListByOwner(ctx context.Context, ownerID domain.UserID) ([]*domain.Task, error)
}

// ActivityRepository はタスクに対する操作の履歴を記録する
type ActivityRepository interface {
	Add(ctx context.Context, activity domain.Activity) error
}

// Transactor は、複数の保存処理を1つのトランザクションとして実行する
type Transactor interface {
	// WithinTx はfnをトランザクションの中で実行する。
	// fnがエラーを返したらすべての変更を取り消し（ロールバック）、nilを返したら確定（コミット）する。
	// fnの中では、引数で渡されたctxを使ってリポジトリを呼び出すこと
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// Clock は現在時刻を返す
type Clock interface {
	Now() time.Time
}

// IDGenerator は新しいタスクIDを払い出す
type IDGenerator interface {
	NewTaskID() domain.TaskID
}
