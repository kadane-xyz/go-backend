-- name: CreateProblem :one
INSERT INTO problem (title, description, points, tags, difficulty) VALUES ($1, $2, $3, $4, $5) RETURNING id;

-- name: CreateProblemCode :exec
INSERT INTO problem_code (problem_id, language, code) VALUES ($1, $2, $3);

-- name: CreateProblemHint :exec
INSERT INTO problem_hint (problem_id, description, answer) VALUES ($1, $2, $3);

-- name: GetProblemsById :many
SELECT * FROM problem WHERE id = ANY(@ids::int[]);

-- name: GetProblem :one
SELECT 
    p.id,
    p.title,
    p.description,
    p.tags,
    p.difficulty,
    p.points,
    json_agg(
        json_build_object(
            'language', pc.language,
            'code', pc.code
        )
    ) FILTER (WHERE pc.id IS NOT NULL) as code,
    json_agg(
        json_build_object(
            'id', ph.id,
            'description', ph.description,
            'answer', ph.answer
        ) ORDER BY ph.id
    ) FILTER (WHERE ph.id IS NOT NULL) as hint,
    json_agg(
        json_build_object(
            'id', pt.id,
            'input', pt.input,
            'output', pt.output
        ) ORDER BY pt.id
    ) FILTER (WHERE pt.id IS NOT NULL) as test_cases
FROM problem p
LEFT JOIN problem_code pc ON p.id = pc.problem_id
LEFT JOIN problem_hint ph ON p.id = ph.problem_id
LEFT JOIN problem_test_case pt ON p.id = pt.problem_id
WHERE p.id = @id
GROUP BY p.id, p.title, p.description, p.tags, p.difficulty, p.points;

-- name: GetProblems :many
WITH problem_data AS (
    SELECT 
        p.id,
        p.title,
        p.description,
        p.tags,
        p.difficulty,
        p.points,
        json_agg(
            json_build_object(
                'language', pc.language,
                'code', pc.code
            )
        ) FILTER (WHERE pc.id IS NOT NULL) as code,
        json_agg(
            json_build_object(
                'description', ph.description,
                'answer', ph.answer
            )
        ) FILTER (WHERE ph.id IS NOT NULL) as hint,
        json_agg(
            json_build_object(
                'input', pt.input,
                'output', pt.output
            )
        ) FILTER (WHERE pt.id IS NOT NULL) as test_cases
    FROM problem p
    LEFT JOIN problem_code pc ON p.id = pc.problem_id
    LEFT JOIN problem_hint ph ON p.id = ph.problem_id
    LEFT JOIN problem_test_case pt ON p.id = pt.problem_id
    GROUP BY p.id, p.title, p.description, p.tags, p.difficulty, p.points
)
SELECT 
    id,
    title,
    description,
    tags,
    difficulty,
    points,
    COALESCE(code, '[]'::json) as code,
    COALESCE(hint, '[]'::json) as hint,
    COALESCE(test_cases, '[]'::json) as test_cases
FROM problem_data
ORDER BY points DESC;

-- name: GetProblemsFilteredPaginated :many
SELECT
    p.*,
    (
        SELECT COALESCE(
            json_agg(
                json_build_object('language', pc.language, 'code', pc.code)
            ),
            '[]'
        )
        FROM problem_code pc 
        WHERE pc.problem_id = p.id
    ) AS code_json,
    
    (
        SELECT COALESCE(
            json_agg(
                json_build_object(
                    'description', encode(ph.description, 'escape'),
                    'answer', encode(ph.answer, 'escape')
                )
            ),
            '[]'
        )
        FROM problem_hint ph 
        WHERE ph.problem_id = p.id
    ) AS hints_json,
    
    (
        SELECT COALESCE(
            json_agg(
                json_build_object('input', pt.input, 'output', pt.output)
            ),
            '[]'
        )
        FROM problem_test_case pt 
        WHERE pt.problem_id = p.id
    ) AS test_cases_json,
    
    (
        SELECT COALESCE(
            json_agg(solution),
            '[]'
        )
        FROM problem_solution ps 
        WHERE ps.problem_id = p.id
    ) AS solutions_json,
    
    EXISTS(
        SELECT 1 
        FROM starred_problem sp 
        WHERE sp.problem_id = p.id 
        AND sp.user_id = @user_id
    ) AS starred
FROM problem p
WHERE
    (@title = '' OR p.title ILIKE '%' || @title || '%')
    AND (@difficulty = '' OR p.difficulty = @difficulty::problem_difficulty)
ORDER BY
    CASE WHEN @sort = 'alpha' AND @sort_direction = 'asc' THEN p.title END ASC,
    CASE WHEN @sort = 'alpha' AND @sort_direction = 'desc' THEN p.title END DESC,
    CASE WHEN @sort = 'index' AND @sort_direction = 'asc' THEN p.id END ASC,
    CASE WHEN @sort = 'index' AND @sort_direction = 'desc' THEN p.id END DESC,
    p.id DESC
LIMIT @per_page OFFSET @page;

-- name: GetProblemCodeByLanguage :one
SELECT * FROM problem_code WHERE problem_id = $1 AND language = $2;

-- name: GetProblemCode :one
SELECT * FROM problem_code WHERE problem_id = $1;

-- name: GetProblemCodes :many
SELECT * FROM problem_code WHERE problem_id = $1;

-- name: GetProblemHints :many
SELECT * FROM problem_hint WHERE problem_id = $1;

-- name: GetProblemByTitle :one
SELECT * FROM problem WHERE title = $1;

-- name: GetProblemByDifficulty :many
SELECT * FROM problem WHERE difficulty = $1 ORDER BY RANDOM()LIMIT $2;

-- name: CreateProblemTestCase :one
INSERT INTO problem_test_case (problem_id, input, output, visibility) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetProblemTestCases :many
SELECT * FROM problem_test_case WHERE problem_id = $1;