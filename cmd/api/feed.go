package main

import (
	"net/http"

	"github.com/Shadowcyng/goSocial/internal/store"
)

// @Summary		get feed
// @Description	get user feed by its followers or own
// @Tags			feed
// @Accept			json
// @Produce		json
// @Param			limit		query		int		false	"Feed limit"
// @Param			offset		query		int		false	"Feed offset"
// @Param			sort_by		query		string	false	"Feed sort_by"
// @Param			sort_order	query		string	false	"Feed sort_order(asc/desc)"
// @Param			tags		query		string	false	"Feed tag comma seprated string max=5"
// @Param			search		query		string	false	"Feed search by title/content"
// @Success		200			{object}	[]store.PostWithMetadata
// @Failure		400			{object}	error	"Bad request"
// @Failure		500			{object}	error	"Somehting went wrong"
// @security		ApiKeyAuth
// @Router			/feed	[get]
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
