-- GET --

-- name: GetRoom :one
SELECT * FROM room WHERE id = $1;

-- name: GetRooms :many
SELECT * FROM room;

-- name: GetRoomsByProblemID :many
SELECT * FROM room WHERE problem_id = $1;

-- name: GetRoomParticipants :many
SELECT * FROM room_participant WHERE room_id = $1;

-- name: GetRoomMessages :many
SELECT * FROM room_message WHERE room_id = $1 ORDER BY created_at ASC;

-- name: GetRoomCount :one
SELECT COUNT(*) FROM room;

-- POST --

-- name: CreateRoom :one
INSERT INTO room (name, problem_id, max_participants, time_limit, creator_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;

-- name: AddParticipantToRoom :exec
INSERT INTO room_participant (room_id, account_id)
VALUES ($1, $2);

-- name: SaveRoomMessage :exec
INSERT INTO room_message (room_id, account_id, content)
VALUES ($1, $2, $3);

-- PUT --

-- name: UpdateRoom :one
UPDATE room
SET name = $2, status = $3, max_participants = $4, time_limit = $5
WHERE id = $1
RETURNING *;

-- name: UpdateParticipantStatus :exec
UPDATE room_participant
SET status = $3
WHERE room_id = $1 AND account_id = $2;

-- DELETE --

-- name: DeleteRoom :exec
DELETE FROM room WHERE id = $1;

-- name: RemoveParticipantFromRoom :exec
DELETE FROM room_participant WHERE room_id = $1 AND account_id = $2;

-- PATCH --

-- name: UpdateRoomStatus :exec
UPDATE room
SET status = $2
WHERE id = $1;

-- Additional Queries --

-- name: GetActiveRooms :many
SELECT * FROM room WHERE status = 'open';

-- name: GetRoomsByCreator :many
SELECT * FROM room WHERE creator_id = $1;

-- name: GetParticipantCount :one
SELECT COUNT(*) FROM room_participant WHERE room_id = $1 AND status = 'active';

-- name: GetLatestRoomMessage :one
SELECT * FROM room_message
WHERE room_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: IsParticipantInRoom :one
SELECT EXISTS(
    SELECT 1 FROM room_participant
    WHERE room_id = $1 AND account_id = $2 AND status = 'active'
);

-- name: GetRoomsPaginated :many
SELECT * FROM room
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: SearchRooms :many
SELECT * FROM room
WHERE name ILIKE '%' || $1 || '%'
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

