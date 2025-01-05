-- GET --

-- name: GetSolution :one
SELECT 
    s.*,
    a.username as user_username,
    a.avatar_url as user_avatar_url,
    a.level as user_level,
    COALESCE(c.comment_count, 0) AS comments_count,
    COALESCE(v.vote_count, 0) AS votes_count,
    COALESCE(uv.vote, 'none') as user_vote
FROM solution s
LEFT JOIN account a ON s.user_id = a.id
LEFT JOIN (
    SELECT solution_id, COUNT(*) AS comment_count
    FROM comment
    GROUP BY solution_id
) c ON s.id = c.solution_id
LEFT JOIN (
    SELECT solution_id, COUNT(*) AS vote_count
    FROM solution_user_vote
    GROUP BY solution_id
) v ON s.id = v.solution_id
LEFT JOIN (
    SELECT solution_id, vote, user_id
    FROM solution_user_vote suv
    WHERE suv.user_id = $2
) uv ON s.id = uv.solution_id
WHERE s.id = $1;

-- name: GetSolutions :many
SELECT 
    s.*,
    a.username as user_username,
    a.avatar_url as user_avatar_url,
    a.level as user_level,
    COALESCE(c.comment_count, 0) AS comments_count,
    COALESCE(v.vote_count, 0) AS votes_count,
    COALESCE(uv.vote, 'none') as user_vote
FROM solution s
LEFT JOIN account a ON s.user_id = a.id
LEFT JOIN (
    SELECT solution_id, COUNT(*) AS comment_count
    FROM comment
    GROUP BY solution_id
) c ON s.id = c.solution_id
LEFT JOIN (
    SELECT solution_id, COUNT(*) AS vote_count
    FROM solution_user_vote
    GROUP BY solution_id
) v ON s.id = v.solution_id
LEFT JOIN (
    SELECT solution_id, vote, user_id
    FROM solution_user_vote suv
    WHERE suv.user_id = $2
) uv ON s.id = uv.solution_id
WHERE s.id = $1;

-- name: GetSolutionsByID :many
SELECT 
    s.id,
    a.username, -- username instead of user_id
    s.title,
    s.body,
    s.problem_id,
    s.tags,
    s.created_at,
    s.votes
FROM solution s
JOIN account a ON s.user_id = a.id
WHERE s.id = ANY(@ids::int[]);

-- name: GetSolutionsCount :one
SELECT COUNT(*) FROM solution
WHERE problem_id = $1
  AND ($2 = '' OR title ILIKE '%' || $2 || '%')
  AND (array_length($3::text[], 1) IS NULL OR tags && $3);

-- name: GetSolutionsPaginated :many
SELECT 
    s.*,
    a.username as user_username,
    a.avatar_url as user_avatar_url,
    a.level as user_level,
    COALESCE(c.comment_count, 0) AS comments_count,
    COALESCE(v.vote_count, 0) AS votes_count
FROM solution s
LEFT JOIN (
    SELECT id, username, avatar_url, level
    FROM account
) a ON s.user_id = a.id
LEFT JOIN (
    SELECT solution_id, COUNT(*) AS comment_count
    FROM comment
    GROUP BY solution_id
) c ON s.id = c.solution_id
LEFT JOIN (
    SELECT solution_id, COUNT(*) AS vote_count
    FROM solution_user_vote
    GROUP BY solution_id
) v ON s.id = v.solution_id
WHERE s.problem_id = $1
  AND ($2::text[] IS NULL OR s.tags && $2)
  AND ($3::text IS NULL OR s.title ILIKE '%' || $3 || '%')
ORDER BY 
    (CASE WHEN $4 = 'id' AND $5 = 'ASC' THEN s.id END) ASC,
    (CASE WHEN $4 = 'id' AND $5 = 'DESC' THEN s.id END) DESC,
    (CASE WHEN $4 = 'created_at' AND $5 = 'ASC' THEN s.created_at END) ASC,
    (CASE WHEN $4 = 'created_at' AND $5 = 'DESC' THEN s.created_at END) DESC,
    (CASE WHEN $4 = 'username' AND $5 = 'ASC' THEN a.username END) ASC,
    (CASE WHEN $4 = 'username' AND $5 = 'DESC' THEN a.username END) DESC,
    (CASE WHEN $4 = 'votes' AND $5 = 'ASC' THEN s.votes END) ASC,
    (CASE WHEN $4 = 'votes' AND $5 = 'DESC' THEN s.votes END) DESC,
    s.id DESC
LIMIT $6 OFFSET $7;

-- name: GetSolutionsWithCommentsCount :many
SELECT s.*, COALESCE(comment_counts.comments_count, 0) AS comments_count
FROM solution s
LEFT JOIN (
    SELECT solution_id, COUNT(*) AS comments_count
    FROM comment
    GROUP BY solution_id
) comment_counts ON s.id = comment_counts.solution_id
WHERE s.problem_id = $1;

SELECT * FROM solution WHERE id = $1;

-- name: GetSolutionVote :one
SELECT vote FROM solution_user_vote
WHERE user_id = $1 AND solution_id = $2;

-- POST --

-- name: CreateSolution :one
INSERT INTO solution (user_id, title, body, problem_id, tags)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- PUT --

-- name: UpdateSolution :one
UPDATE solution
SET title = $1, body = $2, tags = $3
WHERE id = $4 AND user_id = $5
RETURNING *;

-- DELETE --

-- name: DeleteSolution :exec
DELETE FROM solution WHERE id = $1 AND user_id = $2;

-- PATCH --

-- name: InsertSolutionVote :exec
INSERT INTO solution_user_vote (user_id, solution_id, vote)
VALUES ($1, $2, $3);

-- name: DeleteSolutionVote :exec
DELETE FROM solution_user_vote WHERE user_id = $1 AND solution_id = $2;

-- name: UpdateSolutionVote :exec
UPDATE solution_user_vote
SET vote = $3
WHERE user_id = $1 AND solution_id = $2;