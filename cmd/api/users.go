package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Shadowcyng/goSocial/internal/store"
	"github.com/go-chi/chi/v5"
)

type userKey string

const userCtx userKey = "user"

type AuthUser struct {
	UserID int64 `json:"user_id" validate:"required"`
}

// GetUser godoc
//
//	@Summary		Fetches a user profile
//	@Description	Feches a user profile by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	store.User
//	@Failure		400	{object}	error	"Bad request"
//	@Failure		404	{object}	error	"User not found"
//	@Failure		500	{object}	error	"Something went wrong"
//	@security		ApiKeyAuth
//	@Router			/users/{id}	[get]
func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if err := jsonResponse(w, http.StatusCreated, user); err != nil {
		app.internalServerError(w, r, err)
	}
}

// FollowUser godoc
//
//	@Summary		Follow a user profile
//	@Description	Follow a user profile by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int		true	"User ID"
//	@Success		204	{string}	string	"User followed"
//	@Failure		400	{object}	error	"Bad request"
//	@Failure		409	{object}	error	"Conflict: user already in followings"
//	@Failure		404	{object}	error	"User not found"
//	@Failure		500	{object}	error	"Somehting went wrong"
//	@security		ApiKeyAuth
//	@Router			/users/{id}/follow	[put]
func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	followerUser := getUserFromContext(r)

	// TODO: Revert back to auth userID using ctx
	var payload AuthUser
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
	}
	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	fmt.Println(followerUser.ID)
	ctx := r.Context()
	err := app.store.Followers.Follow(ctx, followerUser.ID, payload.UserID)
	if err != nil {
		switch err {
		case store.ErrorConflict:
			app.conflictError(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	if err := jsonResponse(w, http.StatusNoContent, followerUser); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// UnfollowUser godoc
//
//	@Summary		Unfollow a user profile
//	@Description	Unfollow a user profile by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int		true	"User ID"
//	@Success		204	{string}	string	"User unfollowed"
//	@Failure		400	{object}	error	"Bad request"
//	@Failure		404	{object}	error	"User not found"
//	@Failure		500	{object}	error	"Somehting went wrong"
//	@security		ApiKeyAuth
//	@Router			/users/{id}/unfollow	[put]
func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	unfollowedUser := getUserFromContext(r)

	// TODO: Revert back to auth userID using ctx
	var payload AuthUser
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
	}
	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()
	err := app.store.Followers.Unfollow(ctx, unfollowedUser.ID, payload.UserID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusNoContent, unfollowedUser); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) userContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "userID")
		userId, err := strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}
		user, err := app.store.Users.GetById(r.Context(), userId)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrorNotFound):
				app.notFoundResponse(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}
		ctx := context.WithValue(r.Context(), userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserFromContext(r *http.Request) *store.User {
	user, _ := r.Context().Value(userCtx).(*store.User)
	return user
}
