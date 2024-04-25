package data

import (
	"context"
	"database/sql"
	"errors"
	"golangHW.darkhanomirbay/internal/validator"
	"time"
)

type DepartmentInfo struct {
	ID                 int64  `json:"id"`
	DepartmentName     string `json:"department_name"`
	StaffQuantity      int32  `json:"staff_quantity"`
	DepartmentDirector string `json:"department_director"`
	ModuleID           int64  `json:"module_id"`
}
type DepartmentInfoModel struct {
	DB *sql.DB
}

func ValidateDepartmentInfo(v *validator.Validator, departmentInfo *DepartmentInfo) {
	v.Check(departmentInfo.DepartmentName != "", "departmentName", "must be provided")
	v.Check(len(departmentInfo.DepartmentName) <= 500, "departmentName", "must not be more than 500 bytes long")
	v.Check(departmentInfo.StaffQuantity != 0, "StaffQuantity", "must be provided")
	v.Check(departmentInfo.StaffQuantity <= 10, "StaffQuantity", "must not be more than 10")
	v.Check(departmentInfo.DepartmentDirector != "", "DepartmentDirector", "must be provided")
	v.Check(len(departmentInfo.DepartmentDirector) <= 500, "DepartmentDirector", "must not be more than 500 bytes long")
	v.Check(departmentInfo.ModuleID != 0, "ModuleID", "must be provided")
	v.Check(departmentInfo.ModuleID > 0, "ModuleID", "must be positive number")
}
func (m *DepartmentInfoModel) Insert(departmentInfo *DepartmentInfo) error {
	query := `INSERT INTO department_info(department_name,staff_quantity,department_director,module_id) VALUES($1,$2,$3,$4) RETURNING ID`
	args := []any{departmentInfo.DepartmentName, departmentInfo.StaffQuantity, departmentInfo.DepartmentDirector, departmentInfo.ModuleID}
	return m.DB.QueryRow(query, args...).Scan(&departmentInfo.ID)

}
func (m *DepartmentInfoModel) Get(id int64) (*DepartmentInfo, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	query := `SELECT id,department_name,staff_quantity,department_director,module_id FROM department_info WHERE id=$1`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var departmentInfo DepartmentInfo
	err := m.DB.QueryRowContext(ctx, query, id).Scan(&departmentInfo.ID, &departmentInfo.DepartmentName, &departmentInfo.StaffQuantity, &departmentInfo.DepartmentDirector, &departmentInfo.ModuleID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &departmentInfo, nil
}
