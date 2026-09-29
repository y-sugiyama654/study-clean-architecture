// internal/testutil/postgrestest/postgrestest.go

// Package postgrestest は、テスト用のPostgreSQLをコンテナで起動するヘルパー
package postgrestest

import (
	"context"
	"database/sql"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/y-sugiyama654/study-clean-architecture/internal/infrastructure/database"
)

// Start はPostgreSQLのコンテナを起動し、db/schema.sql でテーブルを作ってから、接続とURLを返す
// コンテナはテストの終了時に破棄される
func Start(t testing.TB) (*sql.DB, string) {
	t.Helper()
	ctx := context.Background()

	ctr, err := postgres.Run(ctx, "postgres:17",
		postgres.WithDatabase("tasks"),
		postgres.WithUsername("app"),
		postgres.WithInitScripts(schemaPath()),
		// テストの間だけ使う使い捨てのコンテナなので、パスワードなしで接続できるようにする
		testcontainers.WithEnv(map[string]string{"POSTGRES_HOST_AUTH_METHOD": "trust"}),
		postgres.BasicWaitStrategies(),
	)
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			t.Logf("コンテナを破棄できません: %v", err)
		}
	})
	if err != nil {
		t.Fatalf("PostgreSQLのコンテナを起動できません（Dockerは起動していますか？）: %v", err)
	}

	url, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	db, err := database.OpenPostgres(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db, url
}

// schemaPath はリポジトリの db/schema.sql の絶対パスを返す
func schemaPath() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "..", "db", "schema.sql")
}
