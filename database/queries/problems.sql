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
            json_agg(
                json_build_object(
                    'language', ps.language,
                    'code', ps.code
                )
            ),
            '[]'
        )
        FROM problem_solution ps 
        WHERE ps.problem_id = p.id
    ) AS solutions,
    CASE WHEN EXISTS (SELECT 1 FROM starred_problem sp WHERE sp.problem_id = p.id AND sp.user_id = @user_id) THEN true ELSE false END AS starred,
    CASE WHEN EXISTS (SELECT 1 FROM submission s WHERE s.problem_id = p.id AND s.status = 'Accepted' AND s.account_id = @user_id) THEN true ELSE false END AS solved,
    COUNT(s.id)::int as total_attempts,
    COUNT(s.id) FILTER (WHERE s.status = 'Accepted')::int as total_correct
FROM problem p
LEFT JOIN submission s ON p.id = s.problem_id
WHERE p.id = @problem_id::int
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
        COUNT(s.id)::int as totalAttempts,
        COUNT(s.id) FILTER (WHERE s.status = 'completed' AND s.correct = true)::int as totalCorrect,
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
        (@title::text = '' OR p.title ILIKE '%' || @title::text || '%')
        AND (@difficulty::text = '' OR p.difficulty = @difficulty::problem_difficulty)
    GROUP BY p.id, sp.problem_id
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
            json_agg(
                json_build_object(
                    'language', ps.language,
                    'code', ps.code
                )
            ),
            '[]'
        )
        FROM problem_solution ps 
        WHERE ps.problem_id = p.id
    ) AS solutions,
    COUNT(s.id)::int as total_attempts,
    COUNT(s.id) FILTER (WHERE s.status = 'Accepted')::int as total_correct,
    CASE
        WHEN sp.problem_id IS NOT NULL THEN TRUE
        ELSE FALSE
    END AS starred,
    CASE WHEN EXISTS (SELECT 1 FROM submission s WHERE s.problem_id = p.id AND s.status = 'Accepted' AND s.account_id = @user_id) THEN true ELSE false END AS solved,
    (SELECT COUNT(*) FROM filtered_problems)::int AS total_count
FROM problem p
LEFT JOIN submission s ON p.id = s.problem_id
LEFT JOIN starred_problem sp ON p.id = sp.problem_id
LEFT JOIN (
    SELECT status, problem_id
    FROM submission
    WHERE account_id = @user_id
) AS user_submissions ON p.id = user_submissions.problem_id
WHERE
    (@title::text = '' OR p.title ILIKE '%' || @title::text || '%')
    AND (@difficulty::text = '' OR p.difficulty = @difficulty::problem_difficulty)
GROUP BY p.id, sp.problem_id
ORDER BY
    CASE WHEN @sort::problem_sort = 'alpha' AND @sort_direction::sort_direction = 'asc' THEN p.title END ASC,
    CASE WHEN @sort::problem_sort = 'alpha' AND @sort_direction::sort_direction = 'desc' THEN p.title END DESC,
    CASE WHEN @sort::problem_sort = 'index' AND @sort_direction::sort_direction = 'asc' THEN p.id END ASC,
    CASE WHEN @sort::problem_sort = 'index' AND @sort_direction::sort_direction = 'desc' THEN p.id END DESC,
    p.id DESC
LIMIT @per_page::int
OFFSET ((@page::int) - 1) * @per_page::int;

-- name: GetProblemCodeByLanguage :one
SELECT * FROM problem_code WHERE problem_id = @problem_id::int AND language = @language::problem_language;

-- name: GetProblemCode :one
SELECT * FROM problem_code WHERE problem_id = @problem_id::int;

-- name: GetProblemCodes :many
SELECT * FROM problem_code WHERE problem_id = @problem_id::int;

-- name: GetProblemHints :many
SELECT * FROM problem_hint WHERE problem_id = @problem_id::int;

-- name: GetProblemByTitle :one
SELECT * FROM problem WHERE title = @title::text;

-- name: GetProblemByDifficulty :many
SELECT * FROM problem WHERE difficulty = @difficulty::problem_difficulty ORDER BY RANDOM() LIMIT @per_page::int OFFSET ((@page::int) - 1) * @per_page::int;

-- name: CreateProblem :one
WITH inserted_problem AS (
  INSERT INTO problem (title, description, function_name, points, tags, difficulty)
  VALUES (
    @title,
    @description::text,
    @function_name,
    @points,
    @tags,
    @difficulty
  )
  RETURNING id
),
inserted_problem_hints AS (
  INSERT INTO problem_hint (problem_id, description, answer)
  SELECT 
    (SELECT id FROM inserted_problem),
    d.description,
    a.answer
  FROM 
    unnest(@hint_descriptions::text[]) WITH ORDINALITY AS d(description, idx)
    JOIN unnest(@hint_answers::text[]) WITH ORDINALITY AS a(answer, idx) USING (idx)
  RETURNING *
),
inserted_problem_codes AS (
  INSERT INTO problem_code (problem_id, language, code)
  SELECT 
    (SELECT id FROM inserted_problem),
    l.language,
    c.code  
  FROM 
    unnest(@code_languages::problem_language[]) WITH ORDINALITY AS l(language, idx)
    JOIN unnest(@code_bodies::text[]) WITH ORDINALITY AS c(code, idx) USING (idx)
  RETURNING *
),
inserted_problem_solutions AS (
  INSERT INTO problem_solution (problem_id, language, code)
  SELECT 
    (SELECT id FROM inserted_problem),
    l.language,
    c.code
  FROM 
    unnest(@solution_languages::problem_language[]) WITH ORDINALITY AS l(language, idx)
    JOIN unnest(@solution_codes::text[]) WITH ORDINALITY AS c(code, idx) USING (idx)
  RETURNING *
),
inserted_test_case AS (
  INSERT INTO problem_test_case (problem_id, description, visibility)
  VALUES (
    (SELECT id FROM inserted_problem),
    @test_case_description::text,
    @test_case_visibility::visibility
  )
  RETURNING id
),
inserted_test_case_inputs AS (
  INSERT INTO problem_test_case_input (problem_test_case_id, name, value, type)
  SELECT 
    (SELECT id FROM inserted_test_case),
    n.name,
    v.value,
    t.type
  FROM 
    unnest(@test_case_input_names::text[]) WITH ORDINALITY AS n(name, idx)
    JOIN unnest(@test_case_input_values::text[]) WITH ORDINALITY AS v(value, idx) USING (idx)
    JOIN unnest(@test_case_input_types::problem_test_case_type[]) WITH ORDINALITY AS t(type, idx) USING (idx)
  RETURNING *
),
inserted_test_case_output AS (
  INSERT INTO problem_test_case_output (problem_test_case_id, value)
  VALUES (
    (SELECT id FROM inserted_test_case),
    @test_case_output::text
  )
  RETURNING *
)
SELECT
  (SELECT id FROM inserted_problem) AS problem_id,
  (SELECT COALESCE(json_agg(row_to_json(iphs)), '[]'::json) FROM inserted_problem_hints iphs) AS hints,
  (SELECT COALESCE(json_agg(row_to_json(ipcs)), '[]'::json) FROM inserted_problem_codes ipcs) AS codes,
  (SELECT COALESCE(json_agg(row_to_json(ipss)), '[]'::json) FROM inserted_problem_solutions ipss) AS solutions,
  (SELECT row_to_json(itc) FROM inserted_test_case itc) AS test_case,
  (SELECT COALESCE(json_agg(row_to_json(itcis)), '[]'::json) FROM inserted_test_case_inputs itcis) AS test_case_inputs,
  (SELECT row_to_json(itco) FROM inserted_test_case_output itco) AS test_case_output;

-- name: GetProblemTestCases :many
SELECT 
    ptc.*,
    (SELECT COALESCE(json_agg(
            json_build_object(
                'type', pti.type,
                'value', pti.value,
                'name', pti.name
            )
        ), '[]')
        FROM problem_test_case_input pti 
        WHERE pti.problem_test_case_id = ptc.id
    ) AS input,
    (SELECT COALESCE(value,'') FROM problem_test_case_output pto WHERE pto.problem_test_case_id = ptc.id) AS output
FROM problem_test_case ptc
WHERE ptc.problem_id = @problem_id::int 
    AND (@visibility::text = '' OR ptc.visibility = @visibility::visibility);