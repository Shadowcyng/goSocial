package main

import (
	"net/http"
)

// @Summary		Server Health
// @Description	Check server health
// @Tags			ops
// @Accept			json
// @Produce		json
// @Success		200	{string}	string	"Server is up and running"
// @Failure		400	{object}	error	"Bad request"
// @Failure		500	{object}	error	"Somehting went wrong"
// @security		ApiKeyAuth
// @Router			/health	[get]
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
