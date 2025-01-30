-- name: CreateRoom :one
INSERT INTO room (
    admin,
    name,
    max_players,
    visibility,
    difficulty,
    mode,
    whitelist
) VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: CreateRoomProblem :exec
INSERT INTO room_problems (
    room_id,
    problem_id,
    problem_order
) VALUES ($1, $2, $3);

-- name: GetRooms :many
SELECT 
    r.id, 
    r.admin, 
    r.name, 
    r.created_at, 
    r.max_players, 
    r.visibility, 
    r.difficulty, 
    r.mode,
    COALESCE(
        json_agg(
            row_to_json(p) ORDER BY rp.problem_order
        ) FILTER (WHERE p.id IS NOT NULL),
        '[]'::json
    ) AS problems
FROM room r
LEFT JOIN room_problems rp ON r.id = rp.room_id
LEFT JOIN problem p ON rp.problem_id = p.id
WHERE 
    r.visibility = 'public'
    AND (
        @name::text = '' 
        OR r.name ~* @name::text
    )
    AND (
        @mode::text = '' 
        OR r.mode::text = @mode::text
    )
    AND (
        @difficulty::text = '' 
        OR r.difficulty::text = @difficulty::text
    )
    AND (
        @max_players::int = 0 
        OR r.max_players = @max_players::int
    )
GROUP BY r.id
ORDER BY r.created_at DESC;

-- name: GetRoom :one
SELECT 
    r.id,
    r.admin,
    r.name,
    r.created_at,
    r.max_players,
    r.visibility,
    r.difficulty,
    r.mode,
    COALESCE(
        json_agg(
            row_to_json(p) ORDER BY rp.problem_order
        ) FILTER (WHERE p.id IS NOT NULL),
        '[]'::json
    ) AS problems
FROM room r
LEFT JOIN room_problems rp ON r.id = rp.room_id
LEFT JOIN problem p ON rp.problem_id = p.id
WHERE r.id = $1
GROUP BY r.id; -- If r.id is a primary key, this is sufficient.

-- name: CheckRoomAuthorization :one
SELECT EXISTS (
    SELECT 1
    FROM room
    WHERE id = @room_id
      AND (
        admin = @user_id
        OR visibility = 'public'::visibility
        OR @user_id = ANY(whitelist::text[])
      )
) AS can_access;