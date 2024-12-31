-- GET --

-- name: GetAccount :one
SELECT 
    a.id,
    a.username,
    a.email,
    a.avatar_url,
    a.level,
    a.created_at,
    CASE WHEN @include_attributes::boolean THEN
        json_build_object(
            'bio', COALESCE(aa.bio, ''),
            'contactEmail', COALESCE(aa.contact_email, ''),
            'location', COALESCE(aa.location, ''),
            'realName', COALESCE(aa.real_name, ''),
            'githubUrl', COALESCE(aa.github_url, ''),
            'linkedinUrl', COALESCE(aa.linkedin_url, ''),
            'facebookUrl', COALESCE(aa.facebook_url, ''),
            'instagramUrl', COALESCE(aa.instagram_url, ''),
            'twitterUrl', COALESCE(aa.twitter_url, ''),
            'school', COALESCE(aa.school, ''),
            'websiteUrl', COALESCE(aa.website_url, '')
        )
    ELSE
        NULL
    END as attributes
FROM account a
LEFT JOIN account_attribute aa ON a.id = aa.id
WHERE a.id = @id;

-- name: GetAccounts :many
SELECT 
    a.*,
    CASE 
        WHEN @include_attributes::boolean THEN
            json_build_object(
                'bio', COALESCE(aa.bio, ''),
                'contactEmail', COALESCE(aa.contact_email, ''),
                'location', COALESCE(aa.location, ''),
                'realName', COALESCE(aa.real_name, ''),
                'githubUrl', COALESCE(aa.github_url, ''),
                'linkedinUrl', COALESCE(aa.linkedin_url, ''),
                'facebookUrl', COALESCE(aa.facebook_url, ''),
                'instagramUrl', COALESCE(aa.instagram_url, ''),
                'twitterUrl', COALESCE(aa.twitter_url, ''),
                'school', COALESCE(aa.school, ''),
                'websiteUrl', COALESCE(aa.website_url, '')
            )
        ELSE
            NULL
    END as attributes
FROM account a
LEFT JOIN account_attribute aa ON a.id = aa.id;

-- name: GetAccountExists :one
SELECT EXISTS (SELECT 1 FROM account WHERE id = $1);

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
INSERT INTO account_attribute (
    id,
    bio,
    contact_email,
    location,
    real_name,
    github_url,
    linkedin_url,
    facebook_url,
    instagram_url,
    twitter_url,
    school,
    website_url
) VALUES (
    @id,
    @bio,
    @contact_email,
    @location,
    @real_name,
    @github_url,
    @linkedin_url,
    @facebook_url,
    @instagram_url,
    @twitter_url,
    @school,
    @website_url
)
RETURNING 
    id,
    bio,
    contact_email,
    location,
    real_name,
    github_url,
    linkedin_url,
    facebook_url,
    instagram_url,
    twitter_url,
    school,
    website_url;

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