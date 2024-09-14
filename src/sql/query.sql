-- name: GetSolutions :many
SELECT * FROM solution WHERE problem_id = $1;

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
INSERT INTO solution (email, title, body, problem_id, tags)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateSolution :one
UPDATE solution
SET title = $1, body = $2, tags = $3
WHERE id = $4
RETURNING *;

-- name: DeleteSolution :exec
DELETE FROM solution WHERE id = $1;

-- name: UpVoteSolution :exec
UPDATE solution SET votes = votes + 1 WHERE id = $1;

-- name: DownVoteSolution :exec
UPDATE solution SET votes = votes - 1 WHERE id = $1;

-- name: GetSolutionVote :one
SELECT vote FROM solution_user_vote
WHERE email = $1 AND solution_id = $2;

-- name: InsertSolutionVote :exec
INSERT INTO solution_user_vote (email, solution_id, vote)
VALUES ($1, $2, $3);

-- name: UpdateSolutionVote :exec
UPDATE solution_user_vote
SET vote = $3
WHERE email = $1 AND solution_id = $2;

-- name: GetComments :many
SELECT * FROM comment WHERE solution_id = $1;

-- name: GetComment :one
SELECT * FROM comment WHERE id = $1;

-- name: CreateComment :one
INSERT INTO comment (email, body, solution_id, parent_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateComment :one
UPDATE comment
SET body = $1
WHERE id = $2
RETURNING *;

-- name: DeleteComment :exec
DELETE FROM comment WHERE id = $1;
