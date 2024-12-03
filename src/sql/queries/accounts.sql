-- GET --

-- name: GetAccount :one
SELECT * FROM account WHERE id = $1;

-- name: GetAccounts :many
SELECT * FROM account;

-- name: GetAccountIDByUsername :one
SELECT id from account WHERE username = $1;

-- name: GetAccountUsername :one
SELECT username from account WHERE id = $1;

-- name: GetAccountAvatarUrl :one
SELECT avatar_url from account WHERE id = $1;

-- name: GetAccountLevel :one
SELECT level from account WHERE id = $1;

-- name: GetAccountAttributes :one
SELECT * FROM account_attribute WHERE id = $1;

-- name: GetAccountAttributesWithAccount :one
SELECT * FROM account_attribute
JOIN account ON account_attribute.id = account.id
WHERE account.id = $1;

-- name: GetAccountByUsername :one
SELECT * FROM account WHERE username = $1;

-- POST --

-- name: CreateAccount :exec
INSERT INTO account (id, username, email) VALUES ($1, $2, $3);

-- name: CreateAccountAttributes :one
INSERT INTO account_attribute (id, bio, contact_email, location, real_name, github_url, linkedin_url, facebook_url, instagram_url, twitter_url, school, website_url)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING id;

-- PUT --

-- name: UpdateAvatar :exec
UPDATE account
SET avatar_url = $1
WHERE id = $2;

-- name: UpdateAccountAttributes :one
UPDATE account_attribute
SET bio = $1, contact_email = $2, location = $3, real_name = $4, github_url = $5, linkedin_url = $6, facebook_url = $7, instagram_url = $8, twitter_url = $9, school = $10, website_url = $11
WHERE id = $12 RETURNING *;

-- DELETE --

-- name: DeleteAccount :exec
DELETE FROM account WHERE id = $1;