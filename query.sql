-- name: GetUrl :one
SELECT * FROM urls
WHERE id = ? LIMIT 1;

-- name: ListUrls :many
SELECT * FROM urls
ORDER BY original;

-- name: CreateUrl :one
INSERT INTO urls (
  original, short
) VALUES (
  ?, ?
)
RETURNING *;

-- name: DeleteUrls :exec
DELETE FROM urls
WHERE id = ?;