package errors

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func HandleDatabaseError(err error, resourceName string) *ApiError {
	if errors.Is(err, pgx.ErrNoRows) {
		return NewNotFoundError(resourceName)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505", "23514": // Constraint violations
			return NewApiError(
				err,
				pgErr.Detail,
				http.StatusConflict,
			)
		case "P0001": // Invalid function argument
			return NewApiError(
				err,
				pgErr.Detail,
				http.StatusBadRequest,
			)
		default:
			// Unknown database error - this should be reported
			return NewInternalServerError("A database error occurred")
		}
	}

	// Other unexpected errors
	return NewInternalServerError("An unexpected error occurred")
}
