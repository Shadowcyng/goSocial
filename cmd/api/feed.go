package main

import (
	"net/http"

	"github.com/Shadowcyng/goSocial/internal/store"
)

func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: paginaton, filters, sort
	fq := store.PaginatedFeedQuery{
		Limit:     20,
		Offset:    0,
		SortBy:    "created_at",
		SortOrder: "desc",
		Tags:      nil,
		Search:    "",
	}
	fq = fq.Parse(r)

	if err := Validate.Struct(fq); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	feeds, err := app.store.Posts.GetUserFeed(r.Context(), int64(1), fq)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	if err = jsonResponse(w, http.StatusOK, feeds); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
