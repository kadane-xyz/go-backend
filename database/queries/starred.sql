-- name: GetStarredProblemByProblemID :one
SELECT COUNT(*) > 0 
FROM starred_problem 
WHERE starred_problem.user_id = $1 
  AND problem_id = $2;

-- name: GetStarredProblems :many
SELECT 
    sqlc.embed(p),
    EXISTS (
        SELECT 1 
        FROM starred_problem sp
        WHERE sp.user_id = @user_id::text 
          AND sp.problem_id = p.id
    ) AS starred
FROM problem p;

-- name: GetStarredSolutionByProblemID :one
SELECT COUNT(*) > 0 FROM starred_solution WHERE user_id = $1 AND solution_id = $2;

-- name: GetStarredSolutionBySolutionID :one
SELECT COUNT(*) > 0 
FROM starred_solution 
WHERE starred_solution.user_id = $1 
  AND solution_id = $2;

-- name: GetStarredSolutions :many
SELECT 
    s.*,
    EXISTS (
        SELECT 1
        FROM starred_solution ss
        WHERE ss.user_id = @user_id 
          AND ss.solution_id = s.id
    ) AS starred,
    (SELECT username FROM account WHERE id = @user_id) AS username
FROM solution s;

-- name: GetStarredSubmissions :many
SELECT 
    s.*,
    EXISTS (
        SELECT 1
        FROM starred_submission ss
        WHERE ss.user_id = $1 
          AND ss.submission_id = s.id
    ) AS starred,
    (SELECT account.id FROM account WHERE account.id = s.account_id) AS account_id,
    (SELECT submitted_code FROM submission WHERE id = s.id) AS submitted_code,
    (SELECT submitted_stdin FROM submission WHERE id = s.id) AS submitted_stdin,
    (SELECT created_at FROM submission WHERE id = s.id) AS created_at
FROM submission s;

-- name: GetStarredProblemsByProblemID :many
SELECT * FROM starred_problem WHERE user_id = $1 AND problem_id = $2;

-- name: GetStarredSolutionsBySolutionID :many
SELECT * FROM starred_solution WHERE user_id = $1 AND solution_id = $2;

-- name: GetStarredSubmissionBySubmissionID :one
SELECT COUNT(*) > 0 FROM starred_submission WHERE user_id = $1 AND submission_id = $2;

-- name: PutStarredProblem :one
WITH deleted(starred) AS (
    DELETE FROM starred_problem AS sp
    WHERE sp.user_id = @user_id AND sp.problem_id = @problem_id
    RETURNING false AS starred
),
inserted(starred) AS (
    INSERT INTO starred_problem (user_id, problem_id)
    SELECT v.user_id, v.problem_id
    FROM (VALUES (@user_id, @problem_id)) AS v(user_id, problem_id)
    WHERE NOT EXISTS (SELECT 1 FROM deleted)
    RETURNING true AS starred
)
SELECT starred
FROM deleted
UNION ALL
SELECT starred
FROM inserted
LIMIT 1;

-- name: PutStarredSubmission :one
WITH deleted(starred) AS (
    DELETE FROM starred_submission AS ss
    WHERE ss.user_id = $1 AND ss.submission_id = $2
    RETURNING false AS starred
),
inserted(starred) AS (
    INSERT INTO starred_submission (user_id, submission_id)
    SELECT v.user_id, v.submission_id
    FROM (VALUES (@user_id, @submission_id)) AS v(user_id, submission_id)
    WHERE NOT EXISTS (SELECT 1 FROM deleted)
    RETURNING true AS starred
)
SELECT starred
FROM deleted
UNION ALL
SELECT starred
FROM inserted
LIMIT 1;

-- name: PutStarredSolution :one
WITH deleted(starred) AS (
    DELETE FROM starred_solution AS ss
    WHERE ss.user_id = $1 
      AND ss.solution_id = $2
    RETURNING false AS starred
),
inserted(starred) AS (
    INSERT INTO starred_solution (user_id, solution_id)
    SELECT v.user_id, v.solution_id
    FROM (VALUES (@user_id, @solution_id)) AS v(user_id, solution_id)
    WHERE NOT EXISTS (SELECT 1 FROM deleted)
    RETURNING true AS starred
)
SELECT starred
FROM deleted
UNION ALL
SELECT starred
FROM inserted
LIMIT 1;