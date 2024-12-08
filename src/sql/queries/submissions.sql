-- name: CreateSubmission :one
INSERT INTO submission (token, stdout, time, memory_used, stderr, compile_output, message, status, status_id, status_description, language_id, language_name, account_id, problem_id, submitted_code, submitted_stdin) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16) RETURNING *;

-- name: GetSubmissionByToken :one
SELECT * FROM submission WHERE token = $1;

-- name: GetSubmissionsByProblemID :many
SELECT * FROM submission WHERE problem_id = $1;

-- name: GetSubmissionsByUsername :many
WITH user_submissions AS (
    SELECT 
        s.token,
        s.stdout,
        s.time,
        s.memory_used,
        s.stderr,
        s.compile_output,
        s.message,
        s.status_id,
        s.status_description,
        s.language_id,
        s.language_name,
        s.account_id,
        s.submitted_code,
        s.submitted_stdin,
        s.problem_id,
        s.created_at,
        p.title as problem_title,
        p.description as problem_description,
        p.difficulty as problem_difficulty,
        p.points as problem_points,
        a.username
    FROM submission s
    JOIN account a ON s.account_id = a.id
    JOIN problem p ON s.problem_id = p.id
    WHERE 
        a.username = @username
        AND (@problem_id::uuid IS NULL OR s.problem_id = @problem_id)
    ORDER BY s.created_at DESC
)
SELECT * FROM user_submissions;