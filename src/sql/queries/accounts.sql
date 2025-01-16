-- GET --

-- name: GetAccount :one
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
                'websiteUrl', COALESCE(aa.website_url, ''),
                'friends', COALESCE(f.count, 0),
                'blockedUsers', COALESCE(f2.count, 0),
                'friendRequests', COALESCE(f3.count, 0)
            )
        ELSE
            NULL
        END as attributes
    FROM account a
    LEFT JOIN account_attribute aa ON a.id = aa.id
    LEFT JOIN friendship f ON a.id = f.user_id_1 OR a.id = f.user_id_2 AND f.status = 'accepted'
    LEFT JOIN friendship f2 ON a.id = f2.user_id_1 OR a.id = f2.user_id_2 AND f2.status = 'blocked'
    LEFT JOIN friendship f3 ON a.id = f3.user_id_1 OR a.id = f3.user_id_2 AND f3.status = 'pending'
    WHERE
        a.id = $1 AND
        CASE WHEN array_length(@usernames_filter::text[], 1) > 0 THEN
            a.username = ANY(@usernames_filter::text[])
        ELSE
            TRUE
        END
        AND 
        CASE WHEN @include_attributes::boolean AND array_length(@locations_filter::text[], 1) > 0 THEN
            aa.location = ANY(@locations_filter::text[])
        ELSE
            TRUE
        END
    GROUP BY a.id, aa.id
    ORDER BY 
        (CASE WHEN @sort = 'level' AND @sort_direction = 'ASC' THEN a.level END) ASC,
        (CASE WHEN @sort = 'level' AND @sort_direction = 'DESC' THEN a.level END) DESC
    LIMIT 1;

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
                'websiteUrl', COALESCE(aa.website_url, ''),
                'friendCount', COUNT(DISTINCT CASE WHEN f.status = 'accepted' THEN f.user_id_1 END),
                'blockedCount', COUNT(DISTINCT CASE WHEN f2.status = 'blocked' THEN f2.user_id_1 END),
                'friendRequestCount', COUNT(DISTINCT CASE WHEN f3.status = 'pending' THEN f3.user_id_1 END)
            )
        ELSE
            NULL
        END as attributes
    FROM account a
    LEFT JOIN account_attribute aa ON a.id = aa.id
    LEFT JOIN friendship f ON (a.id = f.user_id_1 OR a.id = f.user_id_2)
    LEFT JOIN friendship f2 ON (a.id = f2.user_id_1 OR a.id = f2.user_id_2)
    LEFT JOIN friendship f3 ON (a.id = f3.user_id_1 OR a.id = f3.user_id_2)
    WHERE
        CASE WHEN array_length(@usernames_filter::text[], 1) > 0 THEN
            a.username = ANY(@usernames_filter::text[])
        ELSE
            TRUE
        END
        AND 
        CASE WHEN @include_attributes::boolean AND array_length(@locations_filter::text[], 1) > 0 THEN
            aa.location = ANY(@locations_filter::text[])
        ELSE
            TRUE
        END
    GROUP BY a.id, aa.id, aa.bio, aa.contact_email, aa.location, aa.real_name, 
             aa.github_url, aa.linkedin_url, aa.facebook_url, aa.instagram_url, 
             aa.twitter_url, aa.school, aa.website_url
    ORDER BY 
        (CASE WHEN @sort = 'level' AND @sort_direction = 'ASC' THEN a.level END) ASC,
        (CASE WHEN @sort = 'level' AND @sort_direction = 'DESC' THEN a.level END) DESC;

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
SELECT 
    a.*,
    CASE
        WHEN f.user_id_1 = @user_id::text OR f.user_id_2 = @user_id::text THEN 
            CASE
                WHEN f.status = 'accepted' THEN 'friend'::text
                WHEN f.status = 'blocked' THEN 'blocked'::text
                WHEN f.status = 'pending' THEN 
                    CASE 
                        WHEN f.initiator_id = @user_id::text THEN 'request_sent'::text
                        ELSE 'request_received'::text
                    END
                ELSE 'none'::text
            END
        ELSE 'none'::text
    END as friend_status,
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
                'websiteUrl', COALESCE(aa.website_url, ''),
                'friendCount', COUNT(DISTINCT CASE WHEN f2.status = 'accepted' THEN f2.user_id_1 END),
                'blockedCount', COUNT(DISTINCT CASE WHEN f3.status = 'blocked' THEN f3.user_id_1 END),
                'friendRequestCount', COUNT(DISTINCT CASE WHEN f4.status = 'pending' THEN f4.user_id_1 END)
            )
        ELSE
            NULL
        END as attributes
    FROM account a
    LEFT JOIN account_attribute aa ON a.id = aa.id
    LEFT JOIN friendship f ON (a.id = f.user_id_1 AND f.user_id_2 = @user_id::text) 
        OR (a.id = f.user_id_2 AND f.user_id_1 = @user_id::text)
    LEFT JOIN friendship f2 ON (a.id = f2.user_id_1 OR a.id = f2.user_id_2) AND f2.status = 'accepted'
    LEFT JOIN friendship f3 ON (a.id = f3.user_id_1 OR a.id = f3.user_id_2) AND f3.status = 'blocked'
    LEFT JOIN friendship f4 ON (a.id = f4.user_id_1 OR a.id = f4.user_id_2) AND f4.status = 'pending'
    WHERE a.username = @username::text
    GROUP BY a.id, aa.id, f.status, f.user_id_1, f.user_id_2;

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