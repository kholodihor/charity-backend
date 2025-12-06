-- name: CreateGoal :one
INSERT INTO goals (
  title,
  description,
  target_amount
) VALUES (
  $1, $2, $3
) RETURNING *;

-- name: GetGoal :one
SELECT * FROM goals
WHERE id = $1 LIMIT 1;

-- name: ListGoals :many
SELECT * FROM goals
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: ListActiveGoals :many
SELECT * FROM goals
WHERE is_active = true
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateGoal :one
UPDATE goals
SET
  title = COALESCE(sqlc.narg(title), title),
  description = COALESCE(sqlc.narg(description), description),
  target_amount = COALESCE(sqlc.narg(target_amount), target_amount),
  is_active = COALESCE(sqlc.narg(is_active), is_active)
WHERE id = sqlc.arg(id)
RETURNING *;