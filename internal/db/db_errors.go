package db

import (
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
)

// ConstraintError is the error returned when a constraint violation occurs
type ConstraintError struct {
	Field      string
	Constraint string
	Message    string
	Code       string
}

// ParseDBError parses the error returned by the database and returns
// a ConstraintError if the error is a constraint violation
func ParseDBError(err error) *ConstraintError {
	var pqErr *pgconn.PgError
	if errors.As(err, &pqErr) {
		return &ConstraintError{
			Field:      pqErr.ColumnName,
			Constraint: pqErr.ConstraintName,
			Message:    pqErr.Message,
			Code:       pqErr.Code,
		}
	}
	return nil
}
