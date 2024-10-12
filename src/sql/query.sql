-- name: GetSolutions :many
SELECT * FROM solution WHERE problem_id = $1;

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

-- name: CreateSolution :one
INSERT INTO solution (username, title, body, problem_id, tags)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateSolution :one
UPDATE solution
SET title = $1, body = $2, tags = $3
WHERE id = $4
RETURNING *;

-- name: DeleteSolution :exec
DELETE FROM solution WHERE id = $1;

-- name: GetSolutionVote :one
SELECT vote FROM solution_user_vote
WHERE username = $1 AND solution_id = $2;

-- name: InsertSolutionVote :exec
INSERT INTO solution_user_vote (username, solution_id, vote)
VALUES ($1, $2, $3);

-- name: DeleteSolutionVote :exec
DELETE FROM solution_user_vote WHERE username = $1 AND solution_id = $2;

-- name: UpdateSolutionVote :exec
UPDATE solution_user_vote
SET vote = $3
WHERE username = $1 AND solution_id = $2;

-- name: GetComment :one
SELECT * FROM comment WHERE id = $1;

-- name: GetComments :many
SELECT * FROM comment WHERE solution_id = $1;

-- name: GetCommentsSorted :many
SELECT * FROM get_comments_sorted($1, $2, $3);

-- name: CreateComment :one
INSERT INTO comment (username, body, solution_id, parent_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateComment :one
UPDATE comment
SET body = $1
WHERE id = $2
RETURNING *;

-- name: DeleteComment :exec
DELETE FROM comment WHERE id = $1;

-- name: GetCommentVote :one
SELECT vote FROM comment_user_vote
WHERE username = $1 AND comment_id = $2;

-- name: GetCommentCount :one
SELECT COUNT(*) FROM comment WHERE solution_id = $1;

-- name: InsertCommentVote :exec
INSERT INTO comment_user_vote (username, comment_id, vote)
VALUES ($1, $2, $3);

-- name: DeleteCommentVote :exec
DELETE FROM comment_user_vote WHERE username = $1 AND comment_id = $2;

-- name: UpdateCommentVote :exec
UPDATE comment_user_vote 
SET vote = $3
WHERE username = $1 AND comment_id = $2;
