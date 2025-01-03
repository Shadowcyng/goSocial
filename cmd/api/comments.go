package main

import (
	"net/http"

	"github.com/Shadowcyng/goSocial/internal/store"
)

type CreateCommentPayload struct {
	UserId  int64       `json:"user_id" validate:"required"`
	User    *store.User `json:"user" validate:"required"`
	Content string      `json:"content" validate:"required,max=500"`
}

func (app *application) createCommentHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromContext(r)
	var payload CreateCommentPayload
	var comment store.Comment
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	comment.Content = payload.Content
	comment.UserID = payload.UserId
	comment.PostID = post.ID
	comment.User = *payload.User
	err := app.store.Comments.Create(r.Context(), &comment)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	jsonResponse(w, http.StatusCreated, comment)
}
