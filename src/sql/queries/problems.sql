-- name: CreateProblem :one
INSERT INTO problem (title, description, prompt, points, tags) VALUES ($1, $2, $3, $4, $5) RETURNING id;

-- name: CreateProblemCode :exec
INSERT INTO problem_code (problem_id, language, code) VALUES ($1, $2, $3);

-- name: CreateProblemHint :exec
INSERT INTO problem_hint (problem_id, description, answer) VALUES ($1, $2, $3);

-- name: GetProblem :one
SELECT * FROM problem WHERE id = $1;

-- name: GetProblems :many
SELECT * FROM problem;

-- name: GetProblemCodes :many
SELECT * FROM problem_code WHERE problem_id = $1;

-- name: GetProblemHints :many
SELECT * FROM problem_hint WHERE problem_id = $1;

-- name: GetProblemByTitle :one
SELECT * FROM problem WHERE title = $1;

-- name: GetProblemSolutionExpectedOutputHash :one
SELECT expected_output_hash FROM problem_solution WHERE problem_id = $1;

-- name: CreateProblemSolution :exec
INSERT INTO problem_solution (problem_id, expected_output, expected_output_hash) VALUES ($1, $2, $3);