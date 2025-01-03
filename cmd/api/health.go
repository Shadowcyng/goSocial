package main

import (
	"net/http"
)

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":  "ok",
		"env":     app.config.env,
		"version": "v1",
	}
	err := jsonResponse(w, http.StatusOK, data)
	if err != nil {
		app.internalServerError(w, r, err)
	}
}