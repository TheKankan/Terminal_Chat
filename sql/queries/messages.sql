-- name: CreateMessage :one
INSERT INTO messages (user_id, created_at, content)
VALUES ($1, NOW(), $2)
RETURNING *;

-- name: GetRecentMessages :many
SELECT m.id, m.created_at, m.user_id, m.content, u.username
FROM messages m
JOIN users u ON m.user_id = u.id
ORDER BY m.created_at DESC
LIMIT $1;