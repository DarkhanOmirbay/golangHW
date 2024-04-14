package data

import "database/sql"

type Models struct {
	ModuleInfoModel ModuleInfoModel
}

func NewModels(db *sql.DB) Models {
	return Models{ModuleInfoModel: ModuleInfoModel{DB: db}}
}
