// cmd/api/main.go
package main

import (
	"context"
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
	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/gateway/memory"
	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/gateway/notifier"
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

	st, err := openStorage(ctx, cfg)
	if err != nil {
		return err
	}
	defer st.close() // サーバーが止まった後に閉じる

	ln, err := net.Listen("tcp", cfg.HTTPAddr)
	if err != nil {
		return err
	}
	srv := &http.Server{
		Handler:           newHandler(cfg, st, lg),
		ReadHeaderTimeout: 5 * time.Second,
	}
	lg.Info("HTTPサーバーを起動します", slog.String("addr", ln.Addr().String()), slog.String("storage", cfg.Storage))
	return web.Serve(ctx, srv, ln, lg)
}

// storage は、データの保存先に関わる実装をまとめたもの
type storage struct {
	tx         usecase.Transactor
	tasks      usecase.TaskRepository
	activities usecase.ActivityRepository
	close      func() error
}

// openStorage は設定に応じて、PostgreSQLかメモリのどちらかの実装を用意する
// 保存先を切り替えても、ユースケースより内側のコードは一切変わらない
func openStorage(ctx context.Context, cfg config.Config) (storage, error) {
	if cfg.Storage == "memory" {
		s := memory.NewStore()
		return storage{
			tx:         memory.NewTransactor(s),
			tasks:      memory.NewTaskRepository(s),
			activities: memory.NewActivityRepository(s),
			close:      func() error { return nil },
		}, nil
	}

	db, err := database.OpenPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		return storage{}, err
	}
	return storage{
		tx:         postgres.NewTransactor(db),
		tasks:      postgres.NewTaskRepository(db),
		activities: postgres.NewActivityRepository(db),
		close:      db.Close,
	}, nil
}

// newHandler は、アプリケーションを構成するオブジェクトをすべて組み立てる（Composition Root）
// 具体的な実装（どのリポジトリ、どの通知先を使うかなど）を選んで結びつけるのは、ここだけ
func newHandler(cfg config.Config, st storage, lg *slog.Logger) http.Handler {
	// Frameworks & Drivers / Interface Adapters
	clock := system.Clock{}
	ids := system.UUIDGenerator{}
	verifier := auth.NewStaticTokenVerifier(cfg.APITokens)
	var notify usecase.Notifier = notifier.Nop{}
	if cfg.NotifyWebhookURL != "" {
		notify = notifier.NewWebhook(cfg.NotifyWebhookURL, &http.Client{Timeout: 5 * time.Second}, lg)
	}

	// Use Cases（ログ出力のデコレータで包む）
	createTask := decorator.WithLogging("CreateTask", usecase.NewCreateTask(st.tx, st.tasks, st.activities, ids, clock), lg)
	getTask := decorator.WithLogging("GetTask", usecase.NewGetTask(st.tasks), lg)
	listTasks := decorator.WithLogging("ListTasks", usecase.NewListTasks(st.tasks), lg)
	completeTask := decorator.WithLogging("CompleteTask", usecase.NewCompleteTask(st.tx, st.tasks, st.activities, notify, clock), lg)

	// Interface Adapters
	tasks := controller.NewTaskController(createTask, getTask, listTasks, completeTask)

	// Frameworks & Drivers
	return web.NewRouter(tasks, verifier, lg)
}
