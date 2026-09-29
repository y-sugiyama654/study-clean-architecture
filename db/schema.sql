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
