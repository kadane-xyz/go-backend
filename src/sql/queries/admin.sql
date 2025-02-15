-- name: ValidateAdmin :one
SELECT admin FROM account WHERE id = $1;