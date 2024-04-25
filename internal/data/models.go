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
	ModuleInfoModel     ModuleInfoModel
	DepartmentInfoModel DepartmentInfoModel
	UserInfoModel       UserInfoModel
	Permissions         PermissionModel // Add a new Permissions field.
	Tokens              TokenModel
}

func NewModels(db *sql.DB) Models {
	return Models{ModuleInfoModel: ModuleInfoModel{DB: db},
		DepartmentInfoModel: DepartmentInfoModel{DB: db},
		UserInfoModel:       UserInfoModel{DB: db},
		Permissions:         PermissionModel{DB: db},
		Tokens:              TokenModel{DB: db},
	}
}
