-- name: CreateDonation :one
INSERT INTO donations (
  user_id,
  goal_id,
  amount,
  currency,
  is_anonymous
) VALUES (
  $1, $2, $3, $4, $5
) RETURNING *;

-- name: CreateAnonymousDonation :one
INSERT INTO donations (
  goal_id,
  amount,
  currency,
  is_anonymous
) VALUES (
  $1, $2, $3, TRUE
) RETURNING *;

-- name: GetDonation :one
SELECT * FROM donations
WHERE id = $1 LIMIT 1;

-- name: ListDonationsByGoal :many
SELECT * FROM donations
WHERE goal_id = $1
ORDER BY created_at DESC
LIMIT $2
OFFSET $3;

-- name: ListDonationsByUser :many
SELECT * FROM donations
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2
OFFSET $3;