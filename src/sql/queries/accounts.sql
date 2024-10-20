-- GET --

-- name: GetAccount :one
SELECT * FROM account WHERE id = $1;

-- name: GetAccounts :many
SELECT * FROM account;

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