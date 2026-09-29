// cmd/api/main.go
//
// 第6章の時点では、メモリ上のリポジトリを使ってAPIを動かす
package main

import (
	"log"
	"net/http"

	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/controller"
	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/gateway/memory"
	"github.com/y-sugiyama654/study-clean-architecture/internal/infrastructure/system"
	"github.com/y-sugiyama654/study-clean-architecture/internal/usecase"
)

func main() {
	repo := memory.NewTaskRepository()
	clock := system.Clock{}
	ids := system.UUIDGenerator{}

	ctrl := controller.NewTaskController(
		usecase.NewCreateTask(repo, ids, clock),
		usecase.NewGetTask(repo),
		usecase.NewListTasks(repo),
		usecase.NewCompleteTask(repo, clock),
	)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /tasks", ctrl.Create)
	mux.HandleFunc("GET /tasks", ctrl.List)
	mux.HandleFunc("GET /tasks/{id}", ctrl.Get)
	mux.HandleFunc("POST /tasks/{id}/complete", ctrl.Complete)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
