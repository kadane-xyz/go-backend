-- GET --

-- name: GetSolutions :many
SELECT * FROM solution WHERE problem_id = $1;

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
SELECT * FROM get_solutions_paginated($1, $2, $3, $4, $5, $6::text[], $7);

-- name: GetSolutionsWithCommentsCount :many
SELECT s.*, COALESCE(comment_counts.comments_count, 0) AS comments_count
FROM solution s
LEFT JOIN (
    SELECT solution_id, COUNT(*) AS comments_count
    FROM comment
    GROUP BY solution_id
) comment_counts ON s.id = comment_counts.solution_id
WHERE s.problem_id = $1;

-- name: GetSolution :one
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