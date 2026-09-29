// internal/adapter/presenter/task_presenter.go
package presenter

import (
	"encoding/json"
	"errors"
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
type ErrorResponse struct {
	Error string `json:"error"`
}

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

// Error はエラーの分類に応じたステータスコードでエラーを返す
func Error(w http.ResponseWriter, err error) {
	status := statusOf(err)
	msg := err.Error()
	if status == http.StatusInternalServerError {
		msg = "サーバー内部でエラーが発生しました" // 内部の詳細はクライアントに見せない
	}
	JSON(w, status, ErrorResponse{Error: msg})
}

// statusOf はエラーをHTTPステータスコードに対応づける
func statusOf(err error) int {
	switch {
	case errors.Is(err, domain.ErrValidation):
		return http.StatusBadRequest
	case errors.Is(err, usecase.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
