-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;


-- name: DeleteAllUsers :exec
DELETE FROM users;

-- name: FindUserByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;

-- name: FindUserById :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: UpdateUserCredentials :one
UPDATE users SET email = $1, hashed_password = $2, updated_at = NOW() WHERE id = $3 RETURNING updated_at;
