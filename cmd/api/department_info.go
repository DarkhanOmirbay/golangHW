package main

import (
	"errors"
	"fmt"
	"golangHW.darkhanomirbay/internal/data"
	"golangHW.darkhanomirbay/internal/validator"
	"net/http"
)

func (app *application) CreateDepInfoHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		DepartmentName     string `json:"department_name"`
		StaffQuantity      int32  `json:"staff_quantity"`
		DepartmentDirector string `json:"department_director"`
		ModuleID           int64  `json:"module_id"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
	}
	v := validator.New()

	departmentInfo := &data.DepartmentInfo{
		DepartmentName:     input.DepartmentName,
		StaffQuantity:      input.StaffQuantity,
		DepartmentDirector: input.DepartmentDirector,
		ModuleID:           input.ModuleID,
	}
	if data.ValidateDepartmentInfo(v, departmentInfo); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.models.DepartmentInfoModel.Insert(departmentInfo)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/departmentinfo/%d", departmentInfo.ID))

	err = app.writeJSON(w, http.StatusOK, envelope{"department info": departmentInfo}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
func (app *application) GetDepInfoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
	}
	departmentInfo, err := app.models.DepartmentInfoModel.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)

		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"department info": departmentInfo}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
