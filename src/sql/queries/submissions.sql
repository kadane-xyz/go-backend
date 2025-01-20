-- name: CreateSubmission :one
INSERT INTO submission (id, stdout, time, memory, stderr, compile_output, message, status, language_id, language_name, account_id, problem_id, submitted_code, submitted_stdin) VALUES (@id::uuid, @stdout::text, @time::text, @memory::int, @stderr::text, @compile_output::text, @message::text, @status, @language_id, @language_name, @account_id, @problem_id, @submitted_code, @submitted_stdin) RETURNING *;

-- name: GetSubmissionByID :one
SELECT 
    *,
    CASE WHEN EXISTS (SELECT 1 FROM starred_submission WHERE submission_id = s.id AND starred_submission.user_id = @user_id) THEN true ELSE false END AS starred
FROM submission s
WHERE s.id = $1;

-- name: GetSubmissionsByID :many
SELECT * FROM submission WHERE id = ANY(@ids::uuid[]);

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
        a.username,
        CASE WHEN EXISTS (SELECT 1 FROM starred_submission WHERE submission_id = s.id AND starred_submission.user_id = @user_id::text) THEN true ELSE false END AS starred
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
    username,
    starred
FROM user_submissions
ORDER BY
    -- 1) Sort by 'time' (the "time" column cast to float)
    CASE
        WHEN @sort = 'time' AND @sort_direction = 'ASC'
            THEN time 
    END ASC,
    CASE
        WHEN @sort = 'time' AND @sort_direction = 'DESC'
            THEN time 
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

    -- 3) Sort by 'created'
    CASE
        WHEN @sort = 'created_at' AND @sort_direction = 'ASC'
            THEN created_at
    END ASC,
    CASE
        WHEN @sort = 'created_at' AND @sort_direction = 'DESC'
            THEN created_at
    END DESC,

    -- 4) Sort by 'status' (alphabetical)
    CASE
        WHEN @sort = 'status' AND @sort_direction = 'ASC'
            THEN status
    END ASC,
    CASE
        WHEN @sort = 'status' AND @sort_direction = 'DESC'
            THEN status
    END DESC,

    -- 5) Fallback ordering for stability
    submission_id DESC
NULLS LAST;
