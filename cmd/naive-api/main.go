// cmd/naive-api/main.go
//
// 第3章: アーキテクチャを考えずに、1つのファイルにすべてを書いたタスク管理API。
// 以降の章で、このコードをクリーンアーキテクチャに沿って作り直していく。
package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQLのドライバ
)

var db *sql.DB

type task struct {
	ID          string     `json:"id"`
	OwnerID     string     `json:"owner_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

func main() {
	var err error
	db, err = sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.HandleFunc("POST /tasks", createTask)
	http.HandleFunc("GET /tasks", listTasks)
	http.HandleFunc("GET /tasks/{id}", getTask)
	http.HandleFunc("POST /tasks/{id}/complete", completeTask)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func createTask(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "X-User-ID ヘッダが必要です", http.StatusUnauthorized)
		return
	}

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "リクエストの形式が不正です", http.StatusBadRequest)
		return
	}

	// 入力チェック
	title := strings.TrimSpace(req.Title)
	if title == "" {
		http.Error(w, "タイトルは必須です", http.StatusBadRequest)
		return
	}
	if utf8.RuneCountInString(title) > 100 {
		http.Error(w, "タイトルは100文字以内にしてください", http.StatusBadRequest)
		return
	}
	if utf8.RuneCountInString(req.Description) > 1000 {
		http.Error(w, "説明は1000文字以内にしてください", http.StatusBadRequest)
		return
	}

	t := task{
		ID:          uuid.NewString(),
		OwnerID:     userID,
		Title:       title,
		Description: req.Description,
		Status:      "todo",
		CreatedAt:   time.Now(),
	}
	_, err := db.ExecContext(r.Context(),
		`INSERT INTO tasks (id, owner_id, title, description, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		t.ID, t.OwnerID, t.Title, t.Description, t.Status, t.CreatedAt)
	if err != nil {
		log.Println(err)
		http.Error(w, "サーバーエラー", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

func listTasks(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "X-User-ID ヘッダが必要です", http.StatusUnauthorized)
		return
	}

	rows, err := db.QueryContext(r.Context(),
		`SELECT id, owner_id, title, description, status, created_at, completed_at
		 FROM tasks WHERE owner_id = $1 ORDER BY created_at`, userID)
	if err != nil {
		log.Println(err)
		http.Error(w, "サーバーエラー", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	tasks := []task{}
	for rows.Next() {
		var t task
		if err := rows.Scan(&t.ID, &t.OwnerID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.CompletedAt); err != nil {
			log.Println(err)
			http.Error(w, "サーバーエラー", http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func getTask(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "X-User-ID ヘッダが必要です", http.StatusUnauthorized)
		return
	}

	var t task
	err := db.QueryRowContext(r.Context(),
		`SELECT id, owner_id, title, description, status, created_at, completed_at
		 FROM tasks WHERE id = $1 AND owner_id = $2`, r.PathValue("id"), userID).
		Scan(&t.ID, &t.OwnerID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.CompletedAt)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "タスクが見つかりません", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Println(err)
		http.Error(w, "サーバーエラー", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

func completeTask(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "X-User-ID ヘッダが必要です", http.StatusUnauthorized)
		return
	}

	var t task
	err := db.QueryRowContext(r.Context(),
		`SELECT id, owner_id, title, description, status, created_at, completed_at
		 FROM tasks WHERE id = $1 AND owner_id = $2`, r.PathValue("id"), userID).
		Scan(&t.ID, &t.OwnerID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.CompletedAt)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "タスクが見つかりません", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Println(err)
		http.Error(w, "サーバーエラー", http.StatusInternalServerError)
		return
	}

	// 業務ルール: 完了済みのタスクは、もう一度完了にはできない
	if t.Status == "done" {
		http.Error(w, "タスクは既に完了しています", http.StatusConflict)
		return
	}

	now := time.Now()
	t.Status = "done"
	t.CompletedAt = &now
	_, err = db.ExecContext(r.Context(),
		`UPDATE tasks SET status = $1, completed_at = $2 WHERE id = $3`,
		t.Status, t.CompletedAt, t.ID)
	if err != nil {
		log.Println(err)
		http.Error(w, "サーバーエラー", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}
