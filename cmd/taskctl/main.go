// cmd/taskctl/main.go
//
// タスクをコマンドラインから操作するCLI。HTTPのAPIと同じユースケースを、別の入口から使う。
//
//	export DATABASE_URL=postgres://app@localhost:5432/tasks?sslmode=disable
//	taskctl -user alice add "牛乳を買う"
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/cli"
	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/gateway/notifier"
	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/gateway/postgres"
	"github.com/y-sugiyama654/study-clean-architecture/internal/infrastructure/database"
	"github.com/y-sugiyama654/study-clean-architecture/internal/infrastructure/system"
	"github.com/y-sugiyama654/study-clean-architecture/internal/usecase"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	user := flag.String("user", os.Getenv("TASKCTL_USER"), "操作するユーザーのID（環境変数 TASKCTL_USER でも指定できる）")
	flag.Parse()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return errors.New("環境変数 DATABASE_URL を設定してください")
	}
	ctx := context.Background()
	db, err := database.OpenPostgres(ctx, dbURL)
	if err != nil {
		return err
	}
	defer db.Close()

	// HTTPのAPI（cmd/api）と同じユースケースを組み立てる。違うのは入口（CLI）だけ
	tx := postgres.NewTransactor(db)
	tasks := postgres.NewTaskRepository(db)
	activities := postgres.NewActivityRepository(db)
	clock := system.Clock{}
	app := cli.New(
		usecase.NewCreateTask(tx, tasks, activities, system.UUIDGenerator{}, clock),
		usecase.NewListTasks(tasks),
		usecase.NewCompleteTask(tx, tasks, activities, notifier.Nop{}, clock),
		os.Stdout,
	)
	return app.Run(ctx, *user, flag.Args())
}
