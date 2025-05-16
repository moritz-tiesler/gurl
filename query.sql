-- name: GetUrlById :one
SELECT * FROM urls
WHERE id = ? LIMIT 1;

-- name: GetUrlByShortUrl :one
SELECT * FROM urls
WHERE short = ? LIMIT 1;

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

-- name: UpdateUrl :exec
UPDATE urls
  set short = ?
WHERE id = ?;

-- name: DeleteUrls :exec
DELETE FROM urls
WHERE id = ?;