package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.HandlerFunc(http.MethodPost, "/v1/moduleinfo", app.requirePermission("movies:read", app.createModuleInfoHandler))
	router.HandlerFunc(http.MethodGet, "/v1/moduleinfo/:id", app.requirePermission("movies:read", app.getModuleInfoHandler))
	router.HandlerFunc(http.MethodGet, "/v1/moduleinfo", app.requirePermission("movies:read", app.getAllModuleInfos))
	router.HandlerFunc(http.MethodPatch, "/v1/moduleinfo/:id", app.requirePermission("movies:read", app.editModuleInfoHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/moduleinfo/:id", app.requirePermission("movies:read", app.deleteModuleInfoHandler))

	router.HandlerFunc(http.MethodPost, "/v1/departmentinfo", app.requirePermission("movies:read", app.CreateDepInfoHandler))
	router.HandlerFunc(http.MethodGet, "/v1/departmentinfo/:id", app.requirePermission("movies:read", app.GetDepInfoHandler))

	//USER
	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)
	router.HandlerFunc(http.MethodGet, "/v1/users/:id", app.requirePermission("movies:read", app.getUserInfoHandler))
	router.HandlerFunc(http.MethodGet, "/v1/users", app.requirePermission("movies:read", app.getAllUserInfos))
	router.HandlerFunc(http.MethodPatch, "/v1/users/:id", app.requirePermission("movies:read", app.editUserInfoHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/users/:id", app.requirePermission("movies:read", app.deleteUserInfoHandler))
	return (app.recoverPanic(app.rateLimit(app.authenticate(router))))
}
