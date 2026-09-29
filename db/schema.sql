-- db/schema.sql
CREATE TABLE tasks (
    id           TEXT PRIMARY KEY,
    owner_id     TEXT NOT NULL,
    title        TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    status       TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ
);

CREATE INDEX tasks_owner_id_created_at_idx ON tasks (owner_id, created_at);

-- タスクに対する操作の履歴（第9章で追加）
CREATE TABLE task_activities (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    task_id     TEXT NOT NULL REFERENCES tasks (id),
    actor_id    TEXT NOT NULL,
    action      TEXT NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX task_activities_task_id_idx ON task_activities (task_id);
