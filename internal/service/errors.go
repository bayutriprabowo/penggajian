package service

import (
	"errors"
)

var (
	ErrBadRequest   = errors.New("permintaan tidak valid")
	ErrUnauthorized = errors.New("kredensial tidak valid")
	ErrForbidden    = errors.New("tidak memiliki akses")
	ErrNotFound     = errors.New("data tidak ditemukan")
	ErrConflict     = errors.New("data sudah ada")
)
