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

-- name: GetAccountAttributes :one
SELECT * FROM account_attributes WHERE id = $1;

-- name: GetAccountAttributesWithAccount :one
SELECT * FROM account_attributes
JOIN account ON account_attributes.id = account.id
WHERE account.id = $1;

-- name: GetAccountByUsername :one
SELECT * FROM account WHERE username = $1;

-- POST --

-- name: CreateAccount :exec
INSERT INTO account (id, username, email, created_at) VALUES ($1, $2, $3, $4);

-- name: CreateAccountAttributes :one
INSERT INTO account_attributes (id, bio, location, real_name, github_url, linkedin_url, facebook_url, instagram_url, twitter_url, school)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id;

-- PUT --

-- name: UpdateAvatar :exec
UPDATE account
SET avatar_url = $1
WHERE id = $2;

-- name: UpdateAccountAttributes :one
UPDATE account_attributes
SET bio = $1, contact_email = $2, location = $3, real_name = $4, github_url = $5, linkedin_url = $6, facebook_url = $7, instagram_url = $8, twitter_url = $9, school = $10
WHERE id = $11 RETURNING *;

-- DELETE --

-- name: DeleteAccount :exec
DELETE FROM account WHERE id = $1;