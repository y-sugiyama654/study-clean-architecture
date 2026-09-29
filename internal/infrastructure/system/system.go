// internal/infrastructure/system/system.go
package system

import (
	"time"

	"github.com/google/uuid"

	"github.com/y-sugiyama654/study-clean-architecture/internal/domain"
)

// Clock は実際の現在時刻を返す usecase.Clock の実装
type Clock struct{}

func (Clock) Now() time.Time { return time.Now() }

// UUIDGenerator はUUID（バージョン7）でタスクIDを払い出す usecase.IDGenerator の実装
// バージョン7のUUIDは先頭が時刻なので、作成順に並べやすい
type UUIDGenerator struct{}

func (UUIDGenerator) NewTaskID() domain.TaskID {
	return domain.TaskID(uuid.Must(uuid.NewV7()).String())
}
