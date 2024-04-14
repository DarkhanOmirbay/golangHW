package data

import (
	"database/sql"
	"time"
)

type ModuleInfo struct {
	ID             int64     `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	ModuleName     string    `json:"module_name"`
	ModuleDuration int32     `json:"module_duration"`
	ExamType       string    `json:"exam_type"`
	Version        int32     `json:"version"`
}
type ModuleInfoModel struct {
	DB *sql.DB
}

//migrate -path=./migrations -database=postgres://postgres:703905@localhost/d.omirbayDB?sslmode=disable up

func (m *ModuleInfoModel) Insert() {}
func (m *ModuleInfoModel) Get()    {}
func (m *ModuleInfoModel) Update() {}
func (m *ModuleInfoModel) Delete() {}
