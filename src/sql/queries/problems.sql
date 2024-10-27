-- name: GetProblemSolutionExpectedOutputHash :one
SELECT expected_output_hash FROM problem_solution WHERE problem_id = $1;
