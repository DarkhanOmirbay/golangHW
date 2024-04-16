package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"golangHW.darkhanomirbay/internal/validator"
	"time"
)

// migrate -path=./migrations -database=postgres://postgres:703905@localhost/d.omirbayDB?sslmode=disable up
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

func ValidateModuleInfo(v *validator.Validator, moduleInfo *ModuleInfo) {
	v.Check(moduleInfo.ModuleName != "", "moduleName", "must be provided")
	v.Check(len(moduleInfo.ModuleName) <= 500, "title", "must not be more than 500 bytes long")
	v.Check(moduleInfo.ModuleDuration != 0, "moduleDuration", "must be provided")
	v.Check(moduleInfo.ModuleDuration <= 10, "moduleDuration", "must not be more than 10")
	v.Check(moduleInfo.ExamType != "", "examType", "must be provided")
	v.Check(len(moduleInfo.ExamType) <= 500, "examType", "must not be more than 500 bytes long")

}

func (m *ModuleInfoModel) Insert(moduleInfo *ModuleInfo) error {
	query := `INSERT INTO module_info(module_name,module_duration,exam_type) VALUES($1,$2,$3) RETURNING ID,created_at,updated_at,version`
	args := []any{moduleInfo.ModuleName, moduleInfo.ModuleDuration, moduleInfo.ExamType}
	return m.DB.QueryRow(query, args...).Scan(&moduleInfo.ID, &moduleInfo.CreatedAt, &moduleInfo.UpdatedAt, &moduleInfo.Version)

}
func (m *ModuleInfoModel) Get(id int64) (*ModuleInfo, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	query := `SELECT id,created_at,updated_at,module_name,module_duration,exam_type,version FROM module_info WHERE id=$1`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var moduleInfo ModuleInfo
	err := m.DB.QueryRowContext(ctx, query, id).Scan(&moduleInfo.ID, &moduleInfo.CreatedAt, &moduleInfo.UpdatedAt, &moduleInfo.ModuleName, &moduleInfo.ModuleDuration, &moduleInfo.ExamType, &moduleInfo.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &moduleInfo, nil
}
func (m *ModuleInfoModel) Update(moduleInfo *ModuleInfo) error {
	query := `UPDATE module_info SET module_name = $1,module_duration = $2,exam_type=$3,version = version +1 WHERE id=$4 AND version=$5 RETURNING version`
	args := []any{moduleInfo.ModuleName, moduleInfo.ModuleDuration, moduleInfo.ExamType, moduleInfo.ID, moduleInfo.Version}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&moduleInfo.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err

		}
	}
	return nil
}
func (m *ModuleInfoModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}
	query := `DELETE FROM module_info WHERE id=$1`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}
func (m *ModuleInfoModel) GetAll(ModuleName string, ExamType string, filters Filters) ([]*ModuleInfo, Metadata, error) {
	query := fmt.Sprintf(`SELECT count(*) OVER(), id, created_at, updated_at,module_name,module_duration,exam_type, version
	FROM module_info
	WHERE (to_tsvector('simple', module_name) @@ plainto_tsquery('simple', $1) OR $1 = '')
	AND (to_tsvector('simple', exam_type) @@ plainto_tsquery('simple', $2) OR $2 = '')
	ORDER BY %s %s,id ASC
	LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, ModuleName, ExamType, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()
	moduleInfos := []*ModuleInfo{}
	totalRecords := 0
	for rows.Next() {
		var moduleInfo ModuleInfo

		err := rows.Scan(&totalRecords, &moduleInfo.ID, &moduleInfo.CreatedAt, &moduleInfo.UpdatedAt, &moduleInfo.ModuleName, &moduleInfo.ModuleDuration, &moduleInfo.ExamType, &moduleInfo.Version)
		if err != nil {
			return nil, Metadata{}, err
		}
		moduleInfos = append(moduleInfos, &moduleInfo)
	}
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return moduleInfos, metadata, nil
}
