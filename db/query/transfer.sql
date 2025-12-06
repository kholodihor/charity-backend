-- name: GetGoalTotalDonations :one
SELECT COALESCE(SUM(amount), 0) AS total_amount
FROM donations
WHERE goal_id = $1;

-- name: GetUserTotalDonations :one
SELECT COALESCE(SUM(amount), 0) AS total_amount
FROM donations
WHERE user_id = $1;

-- name: ListGoalDonors :many
SELECT u.*
FROM users u
JOIN donations d ON d.user_id = u.id
WHERE d.goal_id = $1
GROUP BY u.id
ORDER BY u.id
LIMIT $2
OFFSET $3;