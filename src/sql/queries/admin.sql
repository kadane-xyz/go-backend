-- name: ValidateAdmin :one
SELECT admin FROM account WHERE id = $1;

-- name: GetAdminProblems :many
SELECT 
  p.*,
  COALESCE(json_agg(json_build_object('language', s.language, 'code', s.code)), '[]'::json) as solutions
FROM problem p
LEFT JOIN (
    SELECT problem_id, language, code
    FROM problem_solution
    ORDER BY problem_id, id
) s ON p.id = s.problem_id
GROUP BY p.id;