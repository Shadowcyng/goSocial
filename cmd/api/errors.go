package main

import (
	"fmt"
	"net/http"
)

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("internal server error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	writeJSONError(w, http.StatusInternalServerError, "the server encountered a problem")
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnf("bad request error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	writeJSONError(w, http.StatusBadRequest, err.Error())
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnf("not found error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	writeJSONError(w, http.StatusNotFound, "resource not found")
}

func (app *application) conflictError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("conflict error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	writeJSONError(w, http.StatusConflict, "resource already exists")
}

func (app *application) unauthorizedBasicError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("unauthorized basic error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", chartset="UTF-8`)
	writeJSONError(w, http.StatusUnauthorized, "unauthorized")
}
func (app *application) unauthorizedError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("unauthorized error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	writeJSONError(w, http.StatusUnauthorized, "unauthorized")
}

func (app *application) invalidCredentials(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("invalid credentials", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	writeJSONError(w, http.StatusBadRequest, "invalid credentials")
}

func (app *application) forbiddenError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("forbidden", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	writeJSONError(w, http.StatusForbidden, "forbidden")
}
func (app *application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request, retryAfter string) {
	app.logger.Warnw("rate limit exceeded", "method", r.Method, "path", r.URL.Path)
	w.Header().Set("Retry-After", retryAfter)
	writeJSONError(w, http.StatusTooManyRequests, fmt.Sprintf("rate limit exceeded , retry after %s", retryAfter))
}
