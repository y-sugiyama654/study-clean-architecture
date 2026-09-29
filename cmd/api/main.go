// cmd/api/main.go
//
// 第7章の時点では、設定・ログ・PostgreSQL・ルーターを組み合わせてAPIを動かす
package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/controller"
	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/gateway/postgres"
	"github.com/y-sugiyama654/study-clean-architecture/internal/infrastructure/config"
	"github.com/y-sugiyama654/study-clean-architecture/internal/infrastructure/database"
	"github.com/y-sugiyama654/study-clean-architecture/internal/infrastructure/logger"
	"github.com/y-sugiyama654/study-clean-architecture/internal/infrastructure/system"
	"github.com/y-sugiyama654/study-clean-architecture/internal/infrastructure/web"
	"github.com/y-sugiyama654/study-clean-architecture/internal/usecase"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	lg := logger.New(os.Stdout, cfg.LogLevel)

	db, err := database.OpenPostgres(context.Background(), cfg.DatabaseURL)
	if err != nil {
		lg.Error("データベースに接続できません", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	repo := postgres.NewTaskRepository(db)
	clock := system.Clock{}
	ids := system.UUIDGenerator{}
	ctrl := controller.NewTaskController(
		usecase.NewCreateTask(repo, ids, clock),
		usecase.NewGetTask(repo),
		usecase.NewListTasks(repo),
		usecase.NewCompleteTask(repo, clock),
	)

	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: web.NewRouter(ctrl)}
	lg.Info("HTTPサーバーを起動します", slog.String("addr", cfg.HTTPAddr))
	if err := srv.ListenAndServe(); err != nil {
		lg.Error("HTTPサーバーが停止しました", "error", err)
		os.Exit(1)
	}
}
