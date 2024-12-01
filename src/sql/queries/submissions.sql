-- name: CreateSubmission :one
INSERT INTO submission (token, stdout, time, memory_used, stderr, compile_output, message, status, status_id, status_description, language_id, language_name, account_id, problem_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14) RETURNING *;

-- name: GetSubmissionByToken :one
SELECT * FROM submission WHERE token = $1;

-- name: GetSubmissionsByProblemID :many
SELECT * FROM submission WHERE problem_id = $1;