-- name: GetSolutions :many
SELECT * FROM solutions WHERE problem_id = $1;

-- name: GetSolutionsWithCommentsCount :many
SELECT s.*, COALESCE(COUNT(c.id), 0) AS comments_count
FROM solutions s
LEFT JOIN comments c ON c.solution_id = s.id
WHERE s.problem_id = $1
GROUP BY s.id;

-- name: GetSolution :one
SELECT * FROM solutions WHERE id = $1;

-- name: CreateSolution :one
INSERT INTO solutions (email, title, body, problem_id, tags) VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: UpdateSolution :one
UPDATE solutions SET title = $1, body = $2, tags = $3 WHERE id = $4 RETURNING *;

-- name: DeleteSolution :exec
DELETE FROM solutions WHERE id = $1;

-- name: UpVoteSolution :exec
UPDATE solutions SET votes = votes + 1 WHERE id = $1; 

-- name: DownVoteSolution :exec
UPDATE solutions SET votes = votes - 1 WHERE id = $1;

-- name: GetSolutionVote :one
SELECT vote FROM solutions_user_votes WHERE email = $1 AND solution_id = $2;

-- name: InsertSolutionVote :exec
INSERT INTO solutions_user_votes (email, solution_id, vote) VALUES ($1, $2, $3);

-- name: UpdateSolutionVote :exec
UPDATE solutions_user_votes SET vote = $3 WHERE email = $1 AND solution_id = $2;

-- name: GetComments :many
SELECT * FROM comments WHERE solution_id = $1;

-- name: GetComment :one
SELECT * FROM comments WHERE id = $1;

-- name: CreateComment :one
INSERT INTO comments (email, body, solution_id, parent_id) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: UpdateComment :one
UPDATE comments SET body = $1 WHERE id = $2 RETURNING *;

-- name: DeleteComment :exec
DELETE FROM comments WHERE id = $1; 