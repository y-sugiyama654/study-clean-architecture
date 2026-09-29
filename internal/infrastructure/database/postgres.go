// internal/infrastructure/database/postgres.go
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // database/sql 用のPostgreSQLドライバ（"pgx" という名前で登録される）
)

// OpenPostgres はPostgreSQLに接続し、接続できることを確かめてから *sql.DB を返す
func OpenPostgres(ctx context.Context, url string) (*sql.DB, error) {
	db, err := sql.Open("pgx", url)
	if err != nil {
		return nil, fmt.Errorf("database: 接続を開けません: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)

	// sql.Open は実際には接続しないので、Pingで接続できることを確かめる
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("database: PostgreSQLに接続できません: %w", err)
	}
	return db, nil
}
