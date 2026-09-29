// internal/adapter/gateway/notifier/webhook.go
package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/y-sugiyama654/study-clean-architecture/internal/usecase"
)

// Webhook は、指定したURLにJSONをPOSTして知らせる usecase.Notifier の実装
type Webhook struct {
	url    string
	client *http.Client
	lg     *slog.Logger
}

var _ usecase.Notifier = (*Webhook)(nil)

func NewWebhook(url string, client *http.Client, lg *slog.Logger) *Webhook {
	return &Webhook{url: url, client: client, lg: lg}
}

type taskCompletedEvent struct {
	Event       string     `json:"event"`
	TaskID      string     `json:"task_id"`
	OwnerID     string     `json:"owner_id"`
	Title       string     `json:"title"`
	CompletedAt *time.Time `json:"completed_at"`
}

func (n *Webhook) TaskCompleted(ctx context.Context, task usecase.TaskOutput) {
	body, _ := json.Marshal(taskCompletedEvent{
		Event:       "task.completed",
		TaskID:      task.ID,
		OwnerID:     task.OwnerID,
		Title:       task.Title,
		CompletedAt: task.CompletedAt,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.url, bytes.NewReader(body))
	if err != nil {
		n.lg.WarnContext(ctx, "通知のリクエストを作れません", slog.Any("error", err))
		return
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := n.client.Do(req)
	if err != nil {
		n.lg.WarnContext(ctx, "通知を送れませんでした", slog.String("task_id", task.ID), slog.Any("error", err))
		return
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		n.lg.WarnContext(ctx, "通知先がエラーを返しました",
			slog.String("task_id", task.ID), slog.Int("status", res.StatusCode))
	}
}
