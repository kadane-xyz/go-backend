-- name: CreateFriendRequest :exec
INSERT INTO friendship (user_id_1, user_id_2, initiator_id, status)
SELECT 
    CASE WHEN @user_id < a.id THEN @user_id ELSE a.id END,
    CASE WHEN @user_id < a.id THEN a.id ELSE @user_id END,
    @user_id::text,
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
SET status = 'accepted', accepted_at = NOW()
WHERE ((user_id_1 = @user_id AND user_id_2 = (SELECT a.id FROM account a WHERE a.username = @friend_name))
   OR (user_id_2 = @user_id AND user_id_1 = (SELECT a.id FROM account a WHERE a.username = @friend_name)))
AND status = 'pending';

-- name: BlockFriend :exec
UPDATE friendship
SET status = 'blocked', accepted_at = NULL
WHERE ((user_id_1 = @user_id AND user_id_2 = (SELECT a.id FROM account a WHERE a.username = @friend_name))
   OR (user_id_2 = @user_id AND user_id_1 = (SELECT a.id FROM account a WHERE a.username = @friend_name)))
AND status IN ('pending', 'accepted');


-- name: UnblockFriend :exec
UPDATE friendship
SET status = 'accepted', accepted_at = NOW()
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
    f.accepted_at
FROM friendship f
JOIN account a ON (
    CASE 
        WHEN f.user_id_1 = @user_id THEN a.id = f.user_id_2
        ELSE a.id = f.user_id_1
    END
)
LEFT JOIN account_attribute aa ON a.id = aa.id
WHERE (f.user_id_1 = @user_id OR f.user_id_2 = @user_id) AND f.status = 'accepted';

-- name: GetFriendRequestsSent :many
SELECT 
    a.id AS friend_id,
    a.username AS friend_username,
    COALESCE(a.avatar_url, '')::text AS avatar_url,
    COALESCE(a.level, 0)::int AS level,
    COALESCE(aa.location, '')::text AS location,
    f.created_at
FROM friendship f
JOIN account a ON a.id = CASE 
    WHEN f.user_id_1 = @user_id::text THEN f.user_id_2
    WHEN f.user_id_2 = @user_id::text THEN f.user_id_1
    ELSE f.user_id_2  -- fallback, though this case shouldn't occur
END
LEFT JOIN account_attribute aa ON a.id = aa.id
WHERE f.initiator_id = @user_id::text
AND f.status = 'pending';

-- name: GetFriendRequestsReceived :many
SELECT
    friend.id                         AS friend_id,
    friend.username                   AS friend_username,
    COALESCE(friend.avatar_url, '')::text AS avatar_url,
    COALESCE(friend.level, 0)::int        AS level,
    COALESCE(aa.location, '')::text      AS location,
    f.created_at
FROM friendship f
JOIN account friend
  ON friend.id = CASE
      WHEN f.user_id_1 = @user_id::text THEN f.user_id_2
      ELSE f.user_id_1
    END
LEFT JOIN account_attribute aa
  ON friend.id = aa.id
WHERE f.status = 'pending'
  AND (f.user_id_1 = @user_id::text OR f.user_id_2 = @user_id::text)
  AND f.initiator_id != @user_id::text;

-- name: GetFriendsByUsername :many
WITH user_info AS (
    SELECT a.id, a.username
    FROM account a
    WHERE a.username = @friend_name
)
SELECT 
    a.id as friend_id,
    a.username as friend_username,
    COALESCE(a.avatar_url, '')::text as avatar_url,
    COALESCE(a.level, 0)::int as level,
    COALESCE(aa.location, '')::text as location,
    f.accepted_at,
    CASE 
        WHEN (f.user_id_1 = u.id OR f.user_id_2 = u.id) AND f.status = 'accepted' THEN TRUE
        ELSE FALSE
    END AS is_friend
FROM friendship f
JOIN user_info u ON (f.user_id_1 = u.id OR f.user_id_2 = u.id)
JOIN account a ON (
    CASE 
        WHEN f.user_id_1 = u.id THEN a.id = f.user_id_2
        ELSE a.id = f.user_id_1
    END
)
LEFT JOIN account_attribute aa ON a.id = aa.id
WHERE f.status = 'accepted';