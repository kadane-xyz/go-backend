-- GET -- 

-- name: GetComment :one
SELECT * FROM comment WHERE id = $1;

-- name: GetComments :many
SELECT * FROM comment WHERE solution_id = $1;

-- name: GetCommentsSorted :many
SELECT * FROM get_comments_sorted($1, $2, $3);

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
VALUES ($1, $2, $3, $4)
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

-- name: InsertCommentVote :exec
INSERT INTO comment_user_vote (user_id, comment_id, vote)
VALUES ($1, $2, $3);

-- name: DeleteCommentVote :exec
DELETE FROM comment_user_vote WHERE user_id = $1 AND comment_id = $2;

-- name: UpdateCommentVote :exec
UPDATE comment_user_vote 
SET vote = $3
WHERE user_id = $1 AND comment_id = $2;
