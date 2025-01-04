-- name: GetStarredProblemByProblemID :one
SELECT COUNT(*) > 0 FROM starred_problem WHERE user_id = $1 AND problem_id = $2;

-- name: GetStarredProblems :many
SELECT * FROM starred_problem WHERE user_id = $1;

-- name: GetStarredSolutionByProblemID :one
SELECT COUNT(*) > 0 FROM starred_solution WHERE user_id = $1 AND solution_id = $2;

-- name: GetStarredSolutionBySolutionID :one
SELECT COUNT(*) > 0 FROM starred_solution WHERE user_id = $1 AND solution_id = $2;

-- name: GetStarredSolutions :many
SELECT * FROM starred_solution WHERE user_id = $1;

-- name: GetStarredSubmissions :many
SELECT * FROM starred_submission WHERE user_id = $1;

-- name: GetStarredSubmissionsBySubmissionID :many
SELECT * FROM starred_submission WHERE user_id = $1 AND submission_id = $2;

-- name: GetStarredProblemsByProblemID :many
SELECT * FROM starred_problem WHERE user_id = $1 AND problem_id = $2;

-- name: GetStarredSolutionsBySolutionID :many
SELECT * FROM starred_solution WHERE user_id = $1 AND solution_id = $2;

-- name: GetStarredSubmissionBySubmissionID :one
SELECT COUNT(*) > 0 FROM starred_submission WHERE user_id = $1 AND submission_id = $2;

-- name: PostStarredProblem :exec
INSERT INTO starred_problem (user_id, problem_id) VALUES ($1, $2);

-- name: PostStarredSolution :exec
INSERT INTO starred_solution (user_id, solution_id) VALUES ($1, $2);

-- name: PostStarredSubmission :exec
INSERT INTO starred_submission (user_id, submission_id) VALUES ($1, $2);

-- name: DeleteStarredProblem :exec
DELETE FROM starred_problem WHERE user_id = $1 AND problem_id = $2;

-- name: DeleteStarredSolution :exec
DELETE FROM starred_solution WHERE user_id = $1 AND solution_id = $2;

-- name: DeleteStarredSubmission :exec
DELETE FROM starred_submission WHERE user_id = $1 AND submission_id = $2;