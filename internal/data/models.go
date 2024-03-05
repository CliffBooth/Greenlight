package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict = errors.New("edit conflict")
)

type Models struct {
	Movies MovieDAO
	Users UserDao
}

func NewModels(db *sql.DB) Models {
	return Models {
		Movies: MovieDAO{DB: db},
		Users: UserDao{DB: db},
	}
}