package main

import (
	"errors"
	"fmt"
	"golangHW.darkhanomirbay/internal/data"
	"golangHW.darkhanomirbay/internal/validator"
	"net/http"
)

func (app *application) createModuleInfoHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ModuleName     string `json:"module_name"`
		ModuleDuration int32  `json:"module_duration"`
		ExamType       string `json:"exam_type"`
	}
	//body, err := io.ReadAll(r.Body)
	//if err != nil {
	//	app.serverErrorResponse(w, r, err)
	//	return
	//}
	//err := json.Unmarshal(body, &input)

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
	}
	v := validator.New()
	moduleInfo := &data.ModuleInfo{
		ModuleName:     input.ModuleName,
		ModuleDuration: input.ModuleDuration,
		ExamType:       input.ExamType,
	}
	if data.ValidateModuleInfo(v, moduleInfo); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.models.ModuleInfoModel.Insert(moduleInfo)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/moduleinfo/%d", moduleInfo.ID))

	err = app.writeJSON(w, http.StatusOK, envelope{"module info": moduleInfo}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
func (app *application) getModuleInfoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
	}
	moduleInfo, err := app.models.ModuleInfoModel.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)

		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"module info": moduleInfo}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
func (app *application) getAllModuleInfos(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ModuleName string
		ExamType   string
		Filters    data.Filters
	}
	v := validator.New()
	qs := r.URL.Query()

	input.ModuleName = app.readString(qs, "modulename", "")
	input.ExamType = app.readString(qs, "examtype", "")
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafeList = []string{"modulename", "-modulename", "moduleduration", "-moduleduration", "id", "-id", "moduleduration", "-moduleduration"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	moduleinfo, metadata, err := app.models.ModuleInfoModel.GetAll(input.ModuleName, input.ExamType, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"module infos": moduleinfo, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
func (app *application) editModuleInfoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	moduleInfo, err := app.models.ModuleInfoModel.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
	}
	var input struct {
		ModuleName     *string `json:"module_name"`
		ModuleDuration *int32  `json:"module_duration"`
		ExamType       *string `json:"exam_type"`
	}
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
	}
	if input.ModuleName != nil {
		moduleInfo.ModuleName = *input.ModuleName
	}
	if input.ModuleDuration != nil {
		moduleInfo.ModuleDuration = *input.ModuleDuration
	}
	if input.ExamType != nil {
		moduleInfo.ExamType = *input.ExamType
	}
	v := validator.New()
	if data.ValidateModuleInfo(v, moduleInfo); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
	}
	err = app.models.ModuleInfoModel.Update(moduleInfo)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflicResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)

		}
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"updated module info": moduleInfo}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
func (app *application) deleteModuleInfoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	err = app.models.ModuleInfoModel.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

	}
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "module info successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
