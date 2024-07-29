-- name: CreateUser :one
INSERT INTO users (
    id,
    created_at,
    updated_at,
    name,
    api_key,
    email,
    password
)
VALUES ($1, $2, $3, $4,
    encode(sha256(random()::text::bytea), 'hex'), $5, $6
)
RETURNING *;

-- name: GetUserByAPIKey :one
SELECT *
FROM users
WHERE api_key = $1;

-- name: GetUserByEmail :one
Select * from users where email = $1;