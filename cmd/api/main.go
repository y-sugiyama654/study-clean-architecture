// cmd/api/main.go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/controller"
	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/decorator"
	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/gateway/postgres"
	"github.com/y-sugiyama654/study-clean-architecture/internal/infrastructure/auth"
	"github.com/y-sugiyama654/study-clean-architecture/internal/infrastructure/config"
	"github.com/y-sugiyama654/study-clean-architecture/internal/infrastructure/database"
	"github.com/y-sugiyama654/study-clean-architecture/internal/infrastructure/logger"
	"github.com/y-sugiyama654/study-clean-architecture/internal/infrastructure/system"
	"github.com/y-sugiyama654/study-clean-architecture/internal/infrastructure/web"
	"github.com/y-sugiyama654/study-clean-architecture/internal/usecase"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run はアプリケーションを起動し、終了のシグナルを受け取ったら後片付けをして戻る
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	lg := logger.New(os.Stdout, cfg.LogLevel)
	slog.SetDefault(lg)

	// Ctrl+C（SIGINT）やコンテナ停止時のSIGTERMで、ctxがキャンセルされる
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.OpenPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close() // サーバーが止まった後に閉じる

	ln, err := net.Listen("tcp", cfg.HTTPAddr)
	if err != nil {
		return err
	}
	srv := &http.Server{
		Handler:           newHandler(cfg, db, lg),
		ReadHeaderTimeout: 5 * time.Second,
	}
	lg.Info("HTTPサーバーを起動します", slog.String("addr", ln.Addr().String()))
	return web.Serve(ctx, srv, ln, lg)
}

// newHandler は、アプリケーションを構成するオブジェクトをすべて組み立てる（Composition Root）
// 具体的な実装（PostgreSQLのリポジトリ、実際の時計など）を選んで結びつけるのは、ここだけ
func newHandler(cfg config.Config, db *sql.DB, lg *slog.Logger) http.Handler {
	// Frameworks & Drivers / Interface Adapters
	tx := postgres.NewTransactor(db)
	repo := postgres.NewTaskRepository(db)
	activities := postgres.NewActivityRepository(db)
	clock := system.Clock{}
	ids := system.UUIDGenerator{}

	verifier := auth.NewStaticTokenVerifier(cfg.APITokens)

	// Use Cases（ログ出力のデコレータで包む）
	createTask := decorator.WithLogging("CreateTask", usecase.NewCreateTask(tx, repo, activities, ids, clock), lg)
	getTask := decorator.WithLogging("GetTask", usecase.NewGetTask(repo), lg)
	listTasks := decorator.WithLogging("ListTasks", usecase.NewListTasks(repo), lg)
	completeTask := decorator.WithLogging("CompleteTask", usecase.NewCompleteTask(tx, repo, activities, clock), lg)

	// Interface Adapters
	tasks := controller.NewTaskController(createTask, getTask, listTasks, completeTask)

	// Frameworks & Drivers
	return web.NewRouter(tasks, verifier, lg)
}
