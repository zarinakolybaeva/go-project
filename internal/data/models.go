package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Tasks       TaskModel
	Categories  CategoryModel // Add the Categories field.
	Permissions PermissionModel
	Tokens      TokenModel
	Users       UserModel
}

// NewModels returns a Models struct containing the initialized TaskModel, CategoryModel, etc.
func NewModels(db *sql.DB) Models {
	return Models{
		Tasks:       TaskModel{DB: db},
		Categories:  CategoryModel{DB: db}, // Initialize the CategoryModel instance.
		Permissions: PermissionModel{DB: db},
		Tokens:      TokenModel{DB: db},
		Users:       UserModel{DB: db},
	}
}
