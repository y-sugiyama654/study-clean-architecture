// internal/adapter/controller/task_controller.go
package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/middleware"
	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/presenter"
	"github.com/y-sugiyama654/study-clean-architecture/internal/domain"
	"github.com/y-sugiyama654/study-clean-architecture/internal/usecase"
)

// コントローラーが使うユースケースを、コントローラー側で小さなインターフェースとして定義する（入力ポート）
type (
	CreateTaskUsecase interface {
		Execute(ctx context.Context, in usecase.CreateTaskInput) (usecase.TaskOutput, error)
	}
	GetTaskUsecase interface {
		Execute(ctx context.Context, in usecase.GetTaskInput) (usecase.TaskOutput, error)
	}
	ListTasksUsecase interface {
		Execute(ctx context.Context, in usecase.ListTasksInput) ([]usecase.TaskOutput, error)
	}
	CompleteTaskUsecase interface {
		Execute(ctx context.Context, in usecase.CompleteTaskInput) (usecase.TaskOutput, error)
	}
)

// TaskController はHTTPリクエストをユースケースの入力に変換し、結果をPresenterに渡す
type TaskController struct {
	create   CreateTaskUsecase
	get      GetTaskUsecase
	list     ListTasksUsecase
	complete CompleteTaskUsecase
}

func NewTaskController(create CreateTaskUsecase, get GetTaskUsecase, list ListTasksUsecase, complete CompleteTaskUsecase) *TaskController {
	return &TaskController{create: create, get: get, list: list, complete: complete}
}

// ErrMalformedRequest はリクエストのJSONが読めないことを表す
var ErrMalformedRequest = fmt.Errorf("%w: リクエストの形式が不正です", domain.ErrValidation)

type createTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// Create は POST /tasks を処理する
func (c *TaskController) Create(w http.ResponseWriter, r *http.Request) {
	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		presenter.Error(w, r, ErrMalformedRequest)
		return
	}
	out, err := c.create.Execute(r.Context(), usecase.CreateTaskInput{
		UserID:      currentUserID(r),
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		presenter.Error(w, r, err)
		return
	}
	presenter.JSON(w, http.StatusCreated, presenter.NewTaskResponse(out))
}

// List は GET /tasks を処理する
func (c *TaskController) List(w http.ResponseWriter, r *http.Request) {
	outs, err := c.list.Execute(r.Context(), usecase.ListTasksInput{UserID: currentUserID(r)})
	if err != nil {
		presenter.Error(w, r, err)
		return
	}
	presenter.JSON(w, http.StatusOK, presenter.NewTaskListResponse(outs))
}

// Get は GET /tasks/{id} を処理する
func (c *TaskController) Get(w http.ResponseWriter, r *http.Request) {
	out, err := c.get.Execute(r.Context(), usecase.GetTaskInput{
		UserID: currentUserID(r),
		TaskID: r.PathValue("id"),
	})
	if err != nil {
		presenter.Error(w, r, err)
		return
	}
	presenter.JSON(w, http.StatusOK, presenter.NewTaskResponse(out))
}

// Complete は POST /tasks/{id}/complete を処理する
func (c *TaskController) Complete(w http.ResponseWriter, r *http.Request) {
	out, err := c.complete.Execute(r.Context(), usecase.CompleteTaskInput{
		UserID: currentUserID(r),
		TaskID: r.PathValue("id"),
	})
	if err != nil {
		presenter.Error(w, r, err)
		return
	}
	presenter.JSON(w, http.StatusOK, presenter.NewTaskResponse(out))
}

// currentUserID は、認証ミドルウェアが context に入れた、操作しているユーザーのIDを返す
func currentUserID(r *http.Request) string {
	return middleware.UserID(r.Context()).String()
}
