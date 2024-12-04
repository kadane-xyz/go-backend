-- name: CreateRoom :one
INSERT INTO room (admin, name, max_players, visibility, problems, difficulty, mode, whitelist)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetRooms :many
SELECT * FROM room;

-- name: GetRoom :one
SELECT * FROM room WHERE id = $1;