-- name: CreateUser :one
INSERT INTO
    users (login, hash)
VALUES ($1, $2)
RETURNING
    id,
    login,
    hash;

-- name: GetUserByLogin :one
SELECT id, login, hash FROM users WHERE login = $1 LIMIT 1;