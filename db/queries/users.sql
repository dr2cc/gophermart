-- name: CreateUser :one
INSERT INTO
    users (login, password_hash)
VALUES ($1, $2)
RETURNING
    id,
    login,
    password_hash;

-- name: GetUserByLogin :one
SELECT id, login, password_hash FROM users WHERE login = $1 LIMIT 1;