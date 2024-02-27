package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

type Models struct {
	Movies MovieDAO
}

func NewModels(db *sql.DB) Models {
	return Models {
		Movies: MovieDAO{DB: db},
	}
}