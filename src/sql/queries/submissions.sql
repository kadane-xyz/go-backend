-- name: CreateSubmission :one
INSERT INTO submission (id, stdout, time, memory, stderr, compile_output, message, status, language_id, language_name, account_id, problem_id, submitted_code, submitted_stdin) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14) RETURNING *;

-- name: GetSubmissionByID :one
SELECT * FROM submission WHERE id = $1;

-- name: GetSubmissionsByProblemID :many
SELECT * FROM submission WHERE problem_id = $1;

-- name: GetSubmissionsByUsername :many
WITH user_submissions AS (
    SELECT 
        s.id as submission_id,
        s.stdout,
        s.time,
        s.memory,
        s.stderr,
        s.compile_output,
        s.message,
        s.status,
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
SELECT 
    submission_id as id,
    stdout,
    time,
    memory,
    stderr,
    compile_output,
    message,
    status,
    language_id,
    language_name,
    account_id,
    submitted_code,
    submitted_stdin,
    problem_id,
    created_at,
    problem_title,
    problem_description,
    problem_difficulty,
    problem_points,
    username
FROM user_submissions;