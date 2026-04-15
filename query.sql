-- name: GetHubs :many
SELECT * FROM hubs;

-- name: CreateHub :one
INSERT INTO hubs (id, name, description)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetTasks :many
SELECT * FROM tasks;

-- name: CreateTask :one
INSERT INTO tasks (title, priority)
VALUES ($1, $2)
RETURNING *;
