//go:build wireinject

// cmd/api-wire/wire.go
//
// Wireに「何を使って何を作るか」を宣言するファイル。
// go generate を実行すると、この宣言から wire_gen.go が生成される。
// このファイル自体は wireinject ビルドタグが付いているので、通常のビルドには含まれない
package main

import (
	"database/sql"
	"net/http"

	"github.com/google/wire"

	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/controller"
	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/gateway/postgres"
	"github.com/y-sugiyama654/study-clean-architecture/internal/infrastructure/system"
	"github.com/y-sugiyama654/study-clean-architecture/internal/infrastructure/web"
	"github.com/y-sugiyama654/study-clean-architecture/internal/usecase"
)

func initializeHandler(db *sql.DB) http.Handler {
	wire.Build(
		// 実装を作るプロバイダ
		postgres.NewTaskRepository,
		wire.Value(system.Clock{}),
		wire.Value(system.UUIDGenerator{}),
		usecase.NewCreateTask,
		usecase.NewGetTask,
		usecase.NewListTasks,
		usecase.NewCompleteTask,
		controller.NewTaskController,
		web.NewRouter,

		// インターフェースと実装の対応
		wire.Bind(new(usecase.TaskRepository), new(*postgres.TaskRepository)),
		wire.Bind(new(usecase.Clock), new(system.Clock)),
		wire.Bind(new(usecase.IDGenerator), new(system.UUIDGenerator)),
		wire.Bind(new(controller.CreateTaskUsecase), new(*usecase.CreateTask)),
		wire.Bind(new(controller.GetTaskUsecase), new(*usecase.GetTask)),
		wire.Bind(new(controller.ListTasksUsecase), new(*usecase.ListTasks)),
		wire.Bind(new(controller.CompleteTaskUsecase), new(*usecase.CompleteTask)),
	)
	return nil // wire_gen.go では、実際に組み立てるコードに置き換わる
}
