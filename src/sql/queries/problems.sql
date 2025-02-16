-- name: CreateProblem :one
INSERT INTO problem (title, description, function_name, points, tags, difficulty) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id;

-- name: CreateProblemCode :exec
INSERT INTO problem_code (problem_id, language, code) VALUES ($1, $2, $3);

-- name: CreateProblemHint :exec
INSERT INTO problem_hint (problem_id, description, answer) VALUES ($1, $2, $3);

-- name: GetProblemsById :many
SELECT * FROM problem WHERE id = ANY(@ids::int[]);

-- name: GetProblem :one
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
    ) AS code,
    
    (
        SELECT COALESCE(
            json_agg(
                json_build_object(
                    'description', ph.description,
                    'answer', ph.answer
                )
            ),
            '[]'
        )
        FROM problem_hint ph 
        WHERE ph.problem_id = p.id
    ) AS hints,
    
    (
        SELECT COALESCE(
            json_agg(
                json_build_object(
                    'description', pt.description,
                    'input',
                        (
                            SELECT json_agg(
                                json_build_object(
                                    'name', pti.name,
                                    'type', pti.type,
                                    'value', pti.value
                                )
                            )
                            FROM problem_test_case_input pti
                            WHERE pti.problem_test_case_id = pt.id
                        ),
                    'output',
                        (
                            SELECT pto.value
                            FROM problem_test_case_output pto
                            WHERE pto.problem_test_case_id = pt.id
                        )
                )
            ),
            '[]'
        )
        FROM problem_test_case pt 
        WHERE pt.problem_id = p.id AND pt.visibility = 'public'
    ) AS test_cases,
    
    (
        SELECT COALESCE(
            json_agg(solution),
            '[]'
        )
        FROM problem_solution ps 
        WHERE ps.problem_id = p.id
    ) AS solutions,
    CASE WHEN EXISTS (SELECT 1 FROM starred_problem sp WHERE sp.problem_id = p.id AND sp.user_id = @user_id) THEN true ELSE false END AS starred,
    CASE WHEN EXISTS (SELECT 1 FROM submission s WHERE s.problem_id = p.id AND s.status = 'Accepted' AND s.account_id = @user_id) THEN true ELSE false END AS solved,
    COUNT(s.id) as total_attempts,
    COUNT(s.id) FILTER (WHERE s.status = 'Accepted') as total_correct
FROM problem p
LEFT JOIN submission s ON p.id = s.problem_id
WHERE p.id = @problem_id
GROUP BY p.id;

-- name: GetProblems :many
WITH problem_data AS (
    SELECT 
        p.id,
        p.title,
        p.description,
        p.tags,
        p.difficulty,
        p.points,
        p.function_name,
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
            (
        SELECT COALESCE(
            json_agg(
                json_build_object(
                    'description', pt.description,
                    'input',
                        (
                            SELECT json_agg(
                                json_build_object(
                                    'name', pti.name,
                                    'type', pti.type,
                                    'value', pti.value
                                )
                            )
                            FROM problem_test_case_input pti
                            WHERE pti.problem_test_case_id = pt.id
                        ),
                    'output',
                        (
                            SELECT pto.value
                            FROM problem_test_case_output pto
                            WHERE pto.problem_test_case_id = pt.id
                        )
                )
            ),
            '[]'
        )
        FROM problem_test_case pt 
        WHERE pt.problem_id = p.id AND pt.visibility = 'public'
    ) AS test_cases_json,
        COUNT(s.id) as totalAttempts,
        COUNT(s.id) FILTER (WHERE s.status = 'completed' AND s.correct = true) as totalCorrect,
        CASE WHEN EXISTS (SELECT 1 FROM submission s WHERE s.problem_id = p.id AND s.status = 'Accepted' AND s.account_id = @user_id) THEN true ELSE false END AS solved
    FROM problem p
    LEFT JOIN problem_code pc ON p.id = pc.problem_id
    LEFT JOIN problem_hint ph ON p.id = ph.problem_id
    LEFT JOIN problem_test_case pt ON p.id = pt.problem_id
    LEFT JOIN submission s ON p.id = s.problem_id
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
    COALESCE(test_cases_json, '[]'::json) as test_cases_json
FROM problem_data
ORDER BY points DESC;

-- name: GetProblemsFilteredPaginated :many
WITH filtered_problems AS (
    SELECT p.id
    FROM problem p
    LEFT JOIN submission s ON p.id = s.problem_id
    LEFT JOIN starred_problem sp ON p.id = sp.problem_id
    LEFT JOIN (
        SELECT status, problem_id
        FROM submission
        WHERE account_id = @user_id
    ) AS user_submissions ON p.id = user_submissions.problem_id
    WHERE
        (@title = '' OR p.title ILIKE '%' || @title || '%')
        AND (@difficulty = '' OR p.difficulty = @difficulty::problem_difficulty)
    GROUP BY p.id, sp.problem_id

    -- No ORDER BY / LIMIT / OFFSET here: this is the "full" matching set
)
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
    ) AS code,
    
    (
        SELECT COALESCE(
            json_agg(
                json_build_object(
                    'description', ph.description,
                    'answer', ph.answer
                )
            ),
            '[]'
        )
        FROM problem_hint ph 
        WHERE ph.problem_id = p.id
    ) AS hints,

    (
        SELECT COALESCE(
            json_agg(
                json_build_object(
                    'description', pt.description,
                    'input',
                        (
                            SELECT json_agg(
                                json_build_object(
                                    'name', pti.name,
                                    'type', pti.type,
                                    'value', pti.value
                                )
                            )
                            FROM problem_test_case_input pti
                            WHERE pti.problem_test_case_id = pt.id
                        ),
                    'output',
                        (
                            SELECT pto.value
                            FROM problem_test_case_output pto
                            WHERE pto.problem_test_case_id = pt.id
                        )
                )
            ),
            '[]'
        )
        FROM problem_test_case pt 
        WHERE pt.problem_id = p.id AND pt.visibility = 'public'
    ) AS test_cases, 
    
    (
        SELECT COALESCE(
            json_agg(solution),
            '[]'
        )
        FROM problem_solution ps 
        WHERE ps.problem_id = p.id
    ) AS solutions,
    COUNT(s.id) as total_attempts,
    COUNT(s.id) FILTER (WHERE s.status = 'Accepted') as total_correct,
    CASE
        WHEN sp.problem_id IS NOT NULL THEN TRUE
        ELSE FALSE
    END AS starred,
    CASE WHEN EXISTS (SELECT 1 FROM submission s WHERE s.problem_id = p.id AND s.status = 'Accepted' AND s.account_id = @user_id) THEN true ELSE false END AS solved,
    (SELECT COUNT(*) FROM filtered_problems) AS total_count
FROM problem p
LEFT JOIN submission s ON p.id = s.problem_id
LEFT JOIN starred_problem sp ON p.id = sp.problem_id
LEFT JOIN (
    SELECT status, problem_id
    FROM submission
    WHERE account_id = @user_id
) AS user_submissions ON p.id = user_submissions.problem_id
WHERE
    (@title = '' OR p.title ILIKE '%' || @title || '%')
    AND (@difficulty = '' OR p.difficulty = @difficulty::problem_difficulty)
GROUP BY p.id, sp.problem_id
ORDER BY
    CASE WHEN @sort = 'alpha' AND @sort_direction = 'asc' THEN p.title END ASC,
    CASE WHEN @sort = 'alpha' AND @sort_direction = 'desc' THEN p.title END DESC,
    CASE WHEN @sort = 'index' AND @sort_direction = 'asc' THEN p.id END ASC,
    CASE WHEN @sort = 'index' AND @sort_direction = 'desc' THEN p.id END DESC,
    p.id DESC
LIMIT @per_page
OFFSET (@page - 1) * @per_page;

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
INSERT INTO problem_test_case (problem_id, description, visibility) VALUES ($1, $2, $3) RETURNING *;

-- name: CreateProblemTestCaseInput :one
INSERT INTO problem_test_case_input (problem_test_case_id, name, value, type) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: CreateProblemTestCaseOutput :one
INSERT INTO problem_test_case_output (problem_test_case_id, value) VALUES ($1, $2) RETURNING *;

-- name: GetProblemTestCases :many
SELECT ptc.*,
    (
        SELECT COALESCE(json_agg(
            json_build_object(
                'type', pti.type,
                'value', pti.value
            )
        ), '[]')
        FROM problem_test_case_input pti 
        WHERE pti.problem_test_case_id = ptc.id
    ) AS input,
    (
        SELECT COALESCE(value,'') FROM problem_test_case_output pto WHERE pto.problem_test_case_id = ptc.id
    ) AS output
FROM problem_test_case ptc
WHERE ptc.problem_id = @problem_id::int 
    AND (@visibility::visibility IS NULL OR ptc.visibility = @visibility::visibility);