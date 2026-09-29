// internal/adapter/presenter/task_presenter.go
package presenter

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/y-sugiyama654/study-clean-architecture/internal/domain"
	"github.com/y-sugiyama654/study-clean-architecture/internal/usecase"
)

// TaskResponse はAPIが返すタスクのJSON表現（ビューモデル）
type TaskResponse struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// ErrorResponse はエラーのJSON表現
// code はプログラムで分岐するための値、message は人が読むための説明
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// InternalErrorMessage は、サーバー内部のエラーのときにクライアントへ返す文言
const InternalErrorMessage = "サーバー内部でエラーが発生しました"

// NewTaskResponse はユースケースの出力をレスポンスの形に変換する
func NewTaskResponse(out usecase.TaskOutput) TaskResponse {
	return TaskResponse{
		ID:          out.ID,
		Title:       out.Title,
		Description: out.Description,
		Status:      out.Status,
		CreatedAt:   out.CreatedAt,
		CompletedAt: out.CompletedAt,
	}
}

// NewTaskListResponse は一覧の出力をレスポンスの形に変換する
func NewTaskListResponse(outs []usecase.TaskOutput) []TaskResponse {
	res := make([]TaskResponse, len(outs))
	for i, out := range outs {
		res[i] = NewTaskResponse(out)
	}
	return res
}

// JSON はステータスコードとJSONのボディを書き込む
func JSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

// Error はエラーの分類に応じたステータスコードとコードでエラーを返す
func Error(w http.ResponseWriter, r *http.Request, err error) {
	status, code := classify(err)
	msg := err.Error()
	if status == http.StatusInternalServerError {
		// 内部の詳細（SQLのエラーなど）はクライアントに見せず、ログにだけ残す
		slog.ErrorContext(r.Context(), "リクエストの処理中にエラーが発生しました", slog.Any("error", err))
		msg = InternalErrorMessage
	}
	JSON(w, status, ErrorResponse{Error: ErrorBody{Code: code, Message: msg}})
}

// classify はエラーを、HTTPステータスコードとエラーコードに対応づける
func classify(err error) (status int, code string) {
	switch {
	case errors.Is(err, domain.ErrValidation):
		return http.StatusBadRequest, "invalid_argument"
	case errors.Is(err, usecase.ErrNotFound):
		return http.StatusNotFound, "not_found"
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict, "conflict"
	default:
		return http.StatusInternalServerError, "internal"
	}
}
