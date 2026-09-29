-- db/queries/activities.sql

-- name: InsertActivity :exec
INSERT INTO task_activities (task_id, actor_id, action, occurred_at)
VALUES ($1, $2, $3, $4);
