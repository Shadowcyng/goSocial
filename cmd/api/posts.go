package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/Shadowcyng/goSocial/internal/store"
	"github.com/go-chi/chi/v5"
)

type postKey string

const postCtx postKey = "post"

type CreatePostPayload struct {
	Title   string   `json:"title" validate:"required,max=100"`
	Content string   `json:"content" validate:"required,max=1000"`
	Tags    []string `json:"tags"`
}

// @Summary		Creates post
// @Description	creates post for a user
// @Tags			posts
// @Accept			json
// @Produce		json
// @Param			payload	body		CreatePostPayload	true	"Post title"f
// @Success		201		{object}	store.Post
// @Failure		400		{object}	error	"Bad request"
// @Failure		500		{object}	error	"Somehting went wrong"
// @security		ApiKeyAuth
// @Router			/posts	[post]
func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	userId := 1
	var payload CreatePostPayload
	var post store.Post
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	post.Title = payload.Title
	post.Content = payload.Content
	post.Tags = payload.Tags
	// TODO change after auth
	post.UserID = int64(userId)
	err := app.store.Posts.Create(r.Context(), &post)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	if err := jsonResponse(w, http.StatusCreated, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// @Summary		get post
// @Description	get post by post id
// @Tags			posts
// @Accept			json
// @Produce		json
// @Param			id	path		int	true	"Post id"
// @Success		200	{object}	store.Post
// @Failure		400	{object}	error	"Bad request"
// @Failure		404	{object}	error	"Post not found"
// @Failure		500	{object}	error	"Somehting went wrong"
// @security		ApiKeyAuth
// @Router			/posts/{id}	[get]
func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromContext(r)
	comments, err := app.store.Comments.GetByPostID(r.Context(), post.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	post.Comments = comments
	if err := jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// @Summary		delete post
// @Description	delete post by post id
// @Tags			posts
// @Accept			json
// @Produce		json
// @Param			id	path		int		true	"Post id"
// @Success		200	{string}	string	"Post deleted successfully"
// @Failure		400	{object}	error	"Bad request"
// @Failure		404	{object}	error	"Post not found"
// @Failure		500	{object}	error	"Somehting went wrong"
// @security		ApiKeyAuth
// @Router			/posts/{id}	[delete]
func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "post_id")
	postId, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	err = app.store.Posts.DeleteById(r.Context(), postId)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrorNotFound):
			app.notFoundResponse(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

type UpdatePostPayload struct {
	Title   *string  `json:"title" validate:"omitempty,max=100"`
	Content *string  `json:"content" validate:"omitempty,max=1000"`
	Tags    []string `json:"tags"  validate:"omitempty,max=1000"`
}

// @Summary		update post
// @Description	update post by post id
// @Tags			posts
// @Accept			json
// @Produce		json
// @Param			id		path		int					true	"Post id"
// @Param			payload	body		UpdatePostPayload	false	"Post title"
// @Param			content	body		string				false	"Post content"
// @Param			tags	body		[]string			false	"Post tags"
// @Success		200		{object}	store.Post
// @Failure		400		{object}	error	"Bad request"
// @Failure		404		{object}	error	"Post not found"
// @Failure		500		{object}	error	"Somehting went wrong"
// @security		ApiKeyAuth
// @Router			/posts/{id}	[patch]
func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromContext(r)
	var payload UpdatePostPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if payload.Content != nil {
		post.Content = *payload.Content
	}
	if payload.Title != nil {
		post.Title = *payload.Title
	}
	if payload.Tags != nil {
		post.Tags = payload.Tags
	}

	err := app.store.Posts.UpdatePostById(r.Context(), post)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrorNotFound):
			app.notFoundResponse(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}
	err = jsonResponse(w, http.StatusOK, post)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
	}

}

func (app *application) postContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "postID")
		postId, err := strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}
		post, err := app.store.Posts.GetById(r.Context(), postId)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrorNotFound):
				app.notFoundResponse(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}
		ctx := context.WithValue(r.Context(), postCtx, post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPostFromContext(r *http.Request) *store.Post {
	post, _ := r.Context().Value(postCtx).(*store.Post)
	return post
}
