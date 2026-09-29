-- db/queries/tasks.sql

-- name: UpsertTask :exec
INSERT INTO tasks (id, owner_id, title, description, status, created_at, completed_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (id) DO UPDATE
SET title        = EXCLUDED.title,
    description  = EXCLUDED.description,
    status       = EXCLUDED.status,
    completed_at = EXCLUDED.completed_at;

-- name: GetTask :one
SELECT id, owner_id, title, description, status, created_at, completed_at
FROM tasks
WHERE id = $1;

-- name: GetTaskForUpdate :one
-- 行ロックを取って読む。同じタスクを更新しようとする他のトランザクションは、コミットまで待たされる
SELECT id, owner_id, title, description, status, created_at, completed_at
FROM tasks
WHERE id = $1
FOR UPDATE;

-- name: ListTasksByOwner :many
SELECT id, owner_id, title, description, status, created_at, completed_at
FROM tasks
WHERE owner_id = $1
ORDER BY created_at, id;
