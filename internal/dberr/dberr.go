package dberr

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// IsUniqueViolation mengecek apakah error berasal dari pelanggaran
// unique constraint PostgreSQL (SQLSTATE 23505).
func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// IsForeignKeyViolation mengecek apakah error berasal dari pelanggaran
// foreign key PostgreSQL: 23503 (foreign_key_violation) atau
// 23001 (restrict_violation).
func IsForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && (pgErr.Code == "23503" || pgErr.Code == "23001")
}
