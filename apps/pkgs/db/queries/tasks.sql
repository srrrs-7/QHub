-- name: CreateTask :one
INSERT INTO tasks (
    title,
    description,
    status,
    priority,
    due_date,
    user_id
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetTask :one
SELECT * FROM tasks
WHERE id = $1;

-- name: ListTasks :many
SELECT * FROM tasks
ORDER BY created_at DESC;

-- name: ListTasksByUser :many
SELECT * FROM tasks
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: ListTasksByStatus :many
SELECT * FROM tasks
WHERE status = $1
ORDER BY created_at DESC;

-- name: ListTasksByUserAndStatus :many
SELECT * FROM tasks
WHERE user_id = $1 AND status = $2
ORDER BY created_at DESC;

-- name: UpdateTask :one
UPDATE tasks
SET
    title = COALESCE(sqlc.narg('title'), title),
    description = COALESCE(sqlc.narg('description'), description),
    status = COALESCE(sqlc.narg('status'), status),
    priority = COALESCE(sqlc.narg('priority'), priority),
    due_date = COALESCE(sqlc.narg('due_date'), due_date),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateTaskStatus :one
UPDATE tasks
SET
    status = $2,
    completed_at = CASE WHEN $2 = 'completed' THEN NOW() ELSE NULL END,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteTask :exec
DELETE FROM tasks
WHERE id = $1;

-- name: CountTasksByStatus :one
SELECT COUNT(*) FROM tasks
WHERE status = $1;

-- name: CountTasksByUser :one
SELECT COUNT(*) FROM tasks
WHERE user_id = $1;

-- name: ListOverdueTasks :many
SELECT * FROM tasks
WHERE due_date < NOW()
  AND status NOT IN ('completed', 'cancelled')
ORDER BY due_date ASC;

-- name: ListUpcomingTasks :many
SELECT * FROM tasks
WHERE due_date BETWEEN NOW() AND $1
  AND status NOT IN ('completed', 'cancelled')
ORDER BY due_date ASC;
