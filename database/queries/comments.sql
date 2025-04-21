-- GET -- 

-- name: GetComment :one
SELECT 
    c.*,
    a.username AS user_username,
    a.avatar_url AS user_avatar_url,
    a.level AS user_level,
    COALESCE(v.vote_count, 0)::integer AS votes_count,
    COALESCE(uv.vote, 'none') AS user_vote
FROM comment c
LEFT JOIN (
    SELECT id, username, avatar_url, level
    FROM account
) a ON c.user_id = a.id
LEFT JOIN (
    SELECT comment_id, COUNT(*) AS vote_count
    FROM comment_user_vote
    GROUP BY comment_id
) v ON c.id = v.comment_id
LEFT JOIN (
    SELECT comment_id, vote
    FROM comment_user_vote
    WHERE user_id = @user_id::text 
) uv ON c.id = uv.comment_id
WHERE c.id = @id;

-- name: GetCommentById :one
SELECT 1 FROM comment WHERE id = @id;

-- name: GetComments :many
SELECT 
    c.*,
    a.username AS user_username,
    a.avatar_url AS user_avatar_url,
    a.level AS user_level,
    COALESCE(v.vote_count, 0)::integer AS votes_count,
    COALESCE(uv.vote, 'none') AS user_vote
FROM comment c
LEFT JOIN (
    SELECT id, username, avatar_url, level
    FROM account
) a ON c.user_id = a.id
LEFT JOIN (
    SELECT comment_id, COUNT(*) AS vote_count
    FROM comment_user_vote
    GROUP BY comment_id
) v ON c.id = v.comment_id
LEFT JOIN (
    SELECT comment_id, vote
    FROM comment_user_vote
    WHERE user_id = @user_id::text 
) uv ON c.id = uv.comment_id
WHERE c.id = ANY(@ids::bigint[]);

-- name: GetCommentsSorted :many
SELECT 
    c.*,
    a.username AS user_username,
    a.avatar_url AS user_avatar_url,
    a.level AS user_level,
    COALESCE(v.vote_count, 0)::integer AS votes_count,
    COALESCE(uv.vote, 'none') AS user_vote
FROM comment c
LEFT JOIN (
    SELECT id, username, avatar_url, level
    FROM account
) a ON c.user_id = a.id
LEFT JOIN (
    SELECT comment_id, COUNT(*) AS vote_count
    FROM comment_user_vote
    GROUP BY comment_id
) v ON c.id = v.comment_id
LEFT JOIN (
    SELECT comment_id, vote
    FROM comment_user_vote
    WHERE user_id = @user_id::text 
) uv ON c.id = uv.comment_id
WHERE c.solution_id = @solution_id
ORDER BY 
    (CASE WHEN @sort = 'id'         AND @sort_direction = 'ASC'  THEN c.id          END) ASC,
    (CASE WHEN @sort = 'id'         AND @sort_direction = 'DESC' THEN c.id          END) DESC,
    (CASE WHEN @sort = 'created_at' AND @sort_direction = 'ASC'  THEN c.created_at   END) ASC,
    (CASE WHEN @sort = 'created_at' AND @sort_direction = 'DESC' THEN c.created_at   END) DESC,
    (CASE WHEN @sort = 'username'   AND @sort_direction = 'ASC'  THEN a.username     END) ASC,
    (CASE WHEN @sort = 'username'   AND @sort_direction = 'DESC' THEN a.username     END) DESC,
    (CASE WHEN @sort = 'votes'      AND @sort_direction = 'ASC'  THEN v.vote_count   END) ASC,
    (CASE WHEN @sort = 'votes'      AND @sort_direction = 'DESC' THEN v.vote_count   END) DESC,
    c.id DESC;


-- name: GetCommentVotesBatch :many
SELECT comment_id, vote
FROM comment_user_vote
WHERE user_id = $1 AND comment_id = ANY($2::bigint[]);

-- name: GetCommentCount :one
SELECT COUNT(*) FROM comment WHERE solution_id = $1;

-- name: GetCommentVote :one
SELECT vote FROM comment_user_vote
WHERE user_id = $1 AND comment_id = $2;

-- POST --

-- name: CreateComment :one
INSERT INTO comment (user_id, body, solution_id, parent_id)
VALUES (@user_id::text, @body::text, @solution_id::bigint, sqlc.narg('parent_id')) 
RETURNING *; 

-- PUT --

-- name: UpdateComment :one
UPDATE comment
SET body = $1
WHERE id = $2
AND user_id = $3
RETURNING *;

-- DELETE --

-- name: DeleteComment :exec
DELETE FROM comment WHERE id = $1 AND user_id = $2;

-- PATCH

-- name: VoteComment :exec
SELECT set_comment_vote(@user_id::text, @comment_id::bigint, @vote::vote_type);