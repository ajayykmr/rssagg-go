-- name: CreateFeed :one
INSERT INTO feeds (
    id,
    created_at,
    updated_at,
    name,
    url,
    user_id --created by user_id
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetAllFeeds :many
SELECT *
FROM feeds ORDER BY created_at ASC;

-- name: GetFeedsByUserID :many
select * from feeds where user_id = $1;

-- name: DeleteFeed :exec
DELETE FROM feeds
WHERE id = $1 AND user_id = $2;

-- name: GetNextFeedsToFetch :many
SELECT *
FROM feeds
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT $1;

-- name: MarkFeedAsFetched :one
UPDATE feeds
SET last_fetched_at = NOW(), updated_at = NOW()
WHERE id = $1
RETURNING *;
