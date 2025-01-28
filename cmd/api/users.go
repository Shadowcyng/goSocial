package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/Shadowcyng/goSocial/internal/store"
	"github.com/go-chi/chi/v5"
)

type userKey string

const userCtx userKey = "user"

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
	if err := jsonResponse(w, http.StatusOK, user); err != nil {
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
	followedUser := getUserFromContext(r)
	authUser := getAuthUserFromContext(r)
	ctx := r.Context()
	err := app.store.Followers.Follow(ctx, followedUser.ID, authUser.ID)
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

	if err := jsonResponse(w, http.StatusCreated, followedUser); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// @Summary		Unfollow a user profile
// @Description	Unfollow a user profile by ID
// @Tags			users
// @Accept			json
// @Produce		json
// @Param			id	path		int		true	"User ID"
// @Success		204	{string}	string	"User unfollowed"
// @Failure		400	{object}	error	"Bad request"
// @Failure		404	{object}	error	"User not found"
// @Failure		500	{object}	error	"Somehting went wrong"
// @security		ApiKeyAuth
// @Router			/users/{id}/unfollow	[put]
func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	followerUser := getUserFromContext(r)
	authUser := getAuthUserFromContext(r)
	ctx := r.Context()
	err := app.store.Followers.Unfollow(ctx, followerUser.ID, authUser.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusCreated, followerUser); err != nil {
		app.internalServerError(w, r, err)
	}
}

// @Summary		activate user
// @Description	Activates/Register a user by invitation token
// @Tags			users
// @Produce		json
// @Param			token	path		string	true	"Invitation token"
// @Success		204		{string}	string	"User activated"
// @Failure		400		{object}	error	"Bad request"
// @Failure		404		{object}	error	"User not found"
// @Failure		500		{object}	error	"Somehting went wrong"
// @security		ApiKeyAuth
// @Router			/users/activate/{token} [put]
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	err := app.store.Users.Activate(r.Context(), token)
	if err != nil {
		switch err {
		case store.ErrorNotFound:
			app.badRequestResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}
	if err := jsonResponse(w, http.StatusNoContent, ""); err != nil {
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
		user, err := app.getUser(r.Context(), userId)
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
