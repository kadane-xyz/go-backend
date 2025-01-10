-- name: CreateFriendRequest :exec
INSERT INTO friendship (user_id_1, user_id_2, status)
SELECT 
    CASE WHEN @user_id < a.id THEN @user_id ELSE a.id END,
    CASE WHEN @user_id < a.id THEN a.id ELSE @user_id END,
    'pending'
FROM account a
WHERE a.username = @friend_name
  AND a.username != (SELECT username FROM account WHERE id = @user_id)
  AND NOT EXISTS (
    SELECT 1 
    FROM friendship f 
    WHERE (f.user_id_1 = LEAST(@user_id, a.id) 
      AND f.user_id_2 = GREATEST(@user_id, a.id))
  )
ON CONFLICT (user_id_1, user_id_2) DO NOTHING; 

-- Friend requests

-- name: GetFriendRequestStatus :one
SELECT status
FROM friendship
WHERE (user_id_1 = @user_id AND user_id_2 = (SELECT a.id FROM account a WHERE a.username = @friend_name))
   OR (user_id_2 = @user_id AND user_id_1 = (SELECT a.id FROM account a WHERE a.username = @friend_name));

-- name: AcceptFriendRequest :exec
UPDATE friendship
SET status = 'accepted'
WHERE ((user_id_1 = @user_id AND user_id_2 = (SELECT a.id FROM account a WHERE a.username = @friend_name))
   OR (user_id_2 = @user_id AND user_id_1 = (SELECT a.id FROM account a WHERE a.username = @friend_name)))
AND status = 'pending';

-- name: BlockFriend :exec
UPDATE friendship
SET status = 'blocked'
WHERE ((user_id_1 = @user_id AND user_id_2 = (SELECT a.id FROM account a WHERE a.username = @friend_name))
   OR (user_id_2 = @user_id AND user_id_1 = (SELECT a.id FROM account a WHERE a.username = @friend_name)))
AND status IN ('pending', 'accepted');


-- name: UnblockFriend :exec
UPDATE friendship
SET status = 'accepted'
WHERE ((user_id_1 = @user_id AND user_id_2 = (SELECT a.id FROM account a WHERE a.username = @friend_name))
   OR (user_id_2 = @user_id AND user_id_1 = (SELECT a.id FROM account a WHERE a.username = @friend_name)))
AND status = 'blocked';

-- name: DeleteFriendship :exec
WITH friend_id AS (
    SELECT id
    FROM account
    WHERE username = @friend_name
      AND id != @user_id
)
DELETE FROM friendship
WHERE (user_id_1 = @user_id::text AND user_id_2 IN (SELECT id FROM friend_id))
   OR (user_id_2 = @user_id::text AND user_id_1 IN (SELECT id FROM friend_id));

-- name: GetFriends :many
SELECT 
    a.id as friend_id,
    a.username as friend_username,
    COALESCE(a.avatar_url, '')::text as avatar_url,
    COALESCE(a.level, 0)::int as level,
    COALESCE(aa.location, '')::text as location,
    f.created_at
FROM friendship f
JOIN account a ON (
    CASE 
        WHEN f.user_id_1 = @user_id THEN a.id = f.user_id_2
        ELSE a.id = f.user_id_1
    END
)
LEFT JOIN account_attribute aa ON a.id = aa.id
WHERE (f.user_id_1 = @user_id OR f.user_id_2 = @user_id) AND f.status = 'accepted';

-- name: GetFriendRequests :many
SELECT 
    a.id as friend_id,
    a.username as friend_username,
    COALESCE(a.avatar_url, '')::text as avatar_url,
    COALESCE(a.level, 0)::int as level,
    COALESCE(aa.location, '')::text as location,
    f.created_at
FROM friendship f
JOIN account a ON ( 
    CASE
        WHEN f.user_id_1 = @user_id THEN a.id = f.user_id_2
        ELSE a.id = f.user_id_1
    END
)
LEFT JOIN account_attribute aa ON a.id = aa.id
WHERE (f.user_id_1 = @user_id OR f.user_id_2 = @user_id) AND f.status = 'pending';