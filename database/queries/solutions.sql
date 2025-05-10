-- GET --

-- name: GetSolution :one
SELECT 
    s.*,
    a.username as user_username,
    a.avatar_url as user_avatar_url,
    a.level as user_level,
    COALESCE(c.comment_count, 0)::int as comments_count,
    COALESCE(v.vote_count, 0)::int as votes_count,
    COALESCE(uv.vote, 'none') as user_vote,
    CASE WHEN EXISTS (SELECT 1 FROM starred_solution WHERE solution_id = s.id AND starred_solution.user_id = @user_id) THEN true ELSE false END AS starred
FROM solution s
LEFT JOIN (
    SELECT id, username, avatar_url, level
    FROM account
) a ON s.user_id = a.id
LEFT JOIN (
    SELECT solution_id, COUNT(*) AS comment_count
    FROM comment
    GROUP BY solution_id
) c ON s.id = c.solution_id
LEFT JOIN (
    SELECT solution_id, COUNT(*) AS vote_count
    FROM solution_user_vote
    GROUP BY solution_id
) v ON s.id = v.solution_id
LEFT JOIN (
    SELECT solution_id, vote, user_id
    FROM solution_user_vote suv
    WHERE suv.user_id = @user_id
) uv ON s.id = uv.solution_id
WHERE s.id = @id;

-- name: GetSolutionsByID :many
SELECT 
    s.id,
    a.username, -- username instead of user_id
    s.title,
    s.body,
    s.problem_id,
    s.tags,
    s.created_at,
    s.votes
FROM solution s
JOIN account a ON s.user_id = a.id
WHERE s.id = ANY(@ids::int[]);

-- name: GetSolutionsCount :one
SELECT COUNT(*) FROM solution
WHERE problem_id = $1
  AND (@title = '' OR title ILIKE '%' || @title || '%')
  AND (array_length(@tags::text[], 1) IS NULL OR tags && @tags);

-- name: GetSolutions :many
WITH filtered_solutions AS (
    SELECT s.id
    FROM solution s
    LEFT JOIN comment c ON s.id = c.solution_id
    LEFT JOIN solution_user_vote suv ON s.id = suv.solution_id
    LEFT JOIN starred_solution sp ON s.id = sp.solution_id
    WHERE
        (@title = '' OR s.title ILIKE '%' || @title || '%')
        AND (@tags::text[] IS NULL OR s.tags && @tags::text[])
        AND (@user_id IS NULL OR s.user_id = @user_id)
        AND (@problem_id IS NULL OR s.problem_id = @problem_id)
    GROUP BY s.id, sp.solution_id
)
SELECT
    s.*,
    a.username as user_username,
    a.avatar_url as user_avatar_url,
    a.level as user_level,
    COALESCE(c.comment_count, 0)::int as comments_count,
    COALESCE(v.vote_count, 0)::int as votes_count,
    COALESCE(uv.vote, 'none') as user_vote,
    CASE WHEN EXISTS (SELECT 1 FROM starred_solution WHERE solution_id = s.id AND starred_solution.user_id = @user_id) THEN true ELSE false END AS starred,
    (SELECT COUNT(*) FROM filtered_solutions)::int as total_count
FROM solution s
LEFT JOIN (
    SELECT id, username, avatar_url, level
    FROM account
) a ON s.user_id = a.id
LEFT JOIN (
    SELECT solution_id, COUNT(*) AS comment_count
    FROM comment
    GROUP BY solution_id
) c ON s.id = c.solution_id
LEFT JOIN (
    SELECT solution_id, COUNT(*) AS vote_count
    FROM solution_user_vote
    GROUP BY solution_id
) v ON s.id = v.solution_id
LEFT JOIN (
    SELECT solution_id, vote, user_id
    FROM solution_user_vote suv
    WHERE suv.user_id = @user_id
) uv ON s.id = uv.solution_id
WHERE s.problem_id = @problem_id
  AND (@tags::text[] IS NULL OR s.tags && @tags::text[])
  AND (@title::text IS NULL OR s.title ILIKE '%' || @title || '%')
ORDER BY 
    (CASE WHEN @sort = 'id' AND @sort_direction = 'ASC' THEN s.id END) ASC,
    (CASE WHEN @sort = 'id' AND @sort_direction = 'DESC' THEN s.id END) DESC,
    (CASE WHEN @sort = 'created_at' AND @sort_direction = 'ASC' THEN s.created_at END) ASC,
    (CASE WHEN @sort = 'created_at' AND @sort_direction = 'DESC' THEN s.created_at END) DESC,
    (CASE WHEN @sort = 'username' AND @sort_direction = 'ASC' THEN a.username END) ASC,
    (CASE WHEN @sort = 'username' AND @sort_direction = 'DESC' THEN a.username END) DESC,
    (CASE WHEN @sort = 'votes' AND @sort_direction = 'ASC' THEN s.votes END) ASC,
    (CASE WHEN @sort = 'votes' AND @sort_direction = 'DESC' THEN s.votes END) DESC,
    s.id DESC
LIMIT @per_page::int
OFFSET ((@page::int) - 1) * @per_page::int;

-- name: GetSolutionById :one
SELECT 1 FROM solution WHERE id = @id;

-- name: GetSolutionsWithCommentsCount :many
SELECT s.*, COALESCE(comment_counts.comments_count, 0) AS comments_count
FROM solution s
LEFT JOIN (
    SELECT solution_id, COUNT(*) AS comments_count
    FROM comment
    GROUP BY solution_id
) comment_counts ON s.id = comment_counts.solution_id
WHERE s.problem_id = $1;

SELECT * FROM solution WHERE id = $1;

-- name: GetSolutionVote :one
SELECT vote FROM solution_user_vote
WHERE user_id = $1 AND solution_id = $2;

-- POST --

-- name: CreateSolution :one
INSERT INTO solution (user_id, title, body, problem_id, tags)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- PUT --

-- name: UpdateSolution :one
UPDATE solution
SET title = $1, body = $2, tags = $3
WHERE id = $4 AND user_id = $5
RETURNING *;

-- DELETE --

-- name: DeleteSolution :exec
DELETE FROM solution WHERE id = $1 AND user_id = $2;

-- PATCH --

-- name: VoteSolution :exec
SELECT set_solution_vote(@user_id::text, @solution_id::int, @vote::vote_type);