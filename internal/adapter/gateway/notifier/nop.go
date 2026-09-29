// internal/adapter/gateway/notifier/nop.go
package notifier

import (
	"context"

	"github.com/y-sugiyama654/study-clean-architecture/internal/usecase"
)

// Nop は何もしない usecase.Notifier の実装（通知先を設定しない場合に使う）
type Nop struct{}

var _ usecase.Notifier = Nop{}

func (Nop) TaskCompleted(context.Context, usecase.TaskOutput) {}
