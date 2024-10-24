-- GET --

-- name: GetAccount :one
SELECT * FROM account WHERE id = $1;

-- name: GetAccounts :many
SELECT * FROM account;

-- name: GetAccountUsername :one
SELECT username from account WHERE id = $1;

-- name: GetAccountAvatarUrl :one
SELECT avatar_url from account WHERE id = $1;

-- name: GetAccountLevel :one
SELECT level from account WHERE id = $1;

-- POST --

-- name: CreateAccount :exec
INSERT INTO account (id, username, email)
VALUES ($1, $2, $3);

-- PUT --

-- name: UpdateAvatar :exec
UPDATE account
SET avatar_url = $1
WHERE id = $2;

-- DELETE --

-- name: DeleteAccount :exec
DELETE FROM account WHERE id = $1;