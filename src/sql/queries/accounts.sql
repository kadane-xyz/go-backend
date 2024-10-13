-- GET --

-- name: GetAccount :one
SELECT * FROM account WHERE username = $1;

-- name: GetAccounts :many
SELECT * FROM account;

-- POST --

-- name: CreateAccount :exec
INSERT INTO account (username, email)
VALUES ($1, $2);