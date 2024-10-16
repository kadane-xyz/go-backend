-- GET --

-- name: GetAccount :one
SELECT * FROM account WHERE username = $1;

-- name: GetAccounts :many
SELECT * FROM account;

-- POST --

-- name: CreateAccount :exec
INSERT INTO account (id, username, email)
VALUES ($1, $2, $3);

-- name: UpdateAvatar :exec
UPDATE account
SET avatar_url = $1
WHERE id = $2;