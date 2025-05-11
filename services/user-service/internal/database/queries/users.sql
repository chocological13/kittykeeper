-- name: CreateUser :one
INSERT INTO users (
  username,
  email,
  password_hash,
  first_name,
  last_name
) VALUES (
  $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetUserById :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = $1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;