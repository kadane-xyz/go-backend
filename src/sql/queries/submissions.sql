-- name: CreateSubmission :one
INSERT INTO submission (id, stdout, time, memory, stderr, compile_output, message, status, language_id, language_name, account_id, problem_id, submitted_code, submitted_stdin) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14) RETURNING *;

-- name: GetSubmissionByID :one
SELECT * FROM submission WHERE id = $1;

-- name: GetSubmissionsByProblemID :many
SELECT * FROM submission WHERE problem_id = $1;

-- name: GetSubmissionsByUsername :many
WITH user_submissions AS (
    SELECT
        s.id                AS submission_id,
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
        p.title             AS problem_title,
        p.description       AS problem_description,
        p.difficulty        AS problem_difficulty,
        p.points            AS problem_points,
        a.username
    FROM submission s
    JOIN account a
        ON s.account_id = a.id
    JOIN problem p
        ON s.problem_id = p.id
    WHERE
        a.username = @username
        -- Filter by problem_id only if not 0
        AND (
            @problem_id = 0
            OR s.problem_id = @problem_id
        )
        -- Filter by status only if not empty string
        AND (
            @status = ''
            OR s.status = @status::submission_status
        )
)
SELECT
    submission_id                  AS id,
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
FROM user_submissions
ORDER BY
    -- 1) Sort by 'runtime' (the "time" column cast to float)
    CASE
        WHEN @sort = 'runtime' AND @sort_direction = 'ASC'
            THEN CAST(time AS float)
    END ASC,
    CASE
        WHEN @sort = 'runtime' AND @sort_direction = 'DESC'
            THEN CAST(time AS float)
    END DESC,

    -- 2) Sort by 'memory'
    CASE
        WHEN @sort = 'memory' AND @sort_direction = 'ASC'
            THEN memory
    END ASC,
    CASE
        WHEN @sort = 'memory' AND @sort_direction = 'DESC'
            THEN memory
    END DESC,

    -- 3) Sort by 'createdAt'
    CASE
        WHEN @sort = 'createdAt' AND @sort_direction = 'ASC'
            THEN EXTRACT(EPOCH FROM created_at)
    END ASC,
    CASE
        WHEN @sort = 'createdAt' AND @sort_direction = 'DESC'
            THEN EXTRACT(EPOCH FROM created_at)
    END DESC,

    -- 4) OPTIONAL fallback ordering if none of the above matched:
    --    This ensures there is always a stable ordering.
    submission_id DESC
NULLS LAST;
