// internal/adapter/gateway/postgres/transactor.go
package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/gateway/postgres/sqlcgen"
	"github.com/y-sugiyama654/study-clean-architecture/internal/usecase"
)

// txKey は context にトランザクションを入れるときのキー（他のパッケージのキーと衝突しないよう非公開の型にする）
type txKey struct{}

// Transactor は usecase.Transactor のPostgreSQL実装
// 開始したトランザクション（*sql.Tx）を context に入れて fn に渡し、
// リポジトリは context にトランザクションがあればそれを使う
type Transactor struct {
	db *sql.DB
}

var _ usecase.Transactor = (*Transactor)(nil)

func NewTransactor(db *sql.DB) *Transactor {
	return &Transactor{db: db}
}

func (t *Transactor) WithinTx(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	// すでにトランザクションの中なら、そのトランザクションをそのまま使う
	if _, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return fn(ctx)
	}

	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("postgres: トランザクションを開始できません: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback() // パニックしても、トランザクションを開いたままにしない
			panic(p)
		}
	}()

	if err := fn(context.WithValue(ctx, txKey{}, tx)); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("%w（ロールバックにも失敗しました: %v）", err, rbErr)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("postgres: コミットできません: %w", err)
	}
	return nil
}

// queries は、context にトランザクションがあればその中で、なければ直接 db に対してクエリを実行する
func queries(ctx context.Context, db *sql.DB) *sqlcgen.Queries {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return sqlcgen.New(tx)
	}
	return sqlcgen.New(db)
}
