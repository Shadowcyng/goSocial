package main

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/Shadowcyng/goSocial/internal/store"
	"github.com/google/uuid"
)

type RegisterUserPayload struct {
	Username string `json:"username" validate:"required,max=100"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

// @Summary		user registeration
// @Description	register user using email, username and password
// @Tags			authentication
// @Accept			json
// @Produce		json
// @Param			payload	body		RegisterUserPayload	true	"User credentials"
// @Success		201		{object}	store.User			"User Registered"
// @Failure		400		{object}	error				"Bad request"
// @Failure		500		{object}	error				"Somehting went wrong"
// @security		ApiKeyAuth
// @Router			/authentication/user	[post]
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegisterUserPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &store.User{
		Username: payload.Username,
		Email:    payload.Email,
	}

	// hash the password
	err := user.Password.Set(payload.Password)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	plainToken := uuid.New().String()

	// store this in user_invitation tables after hash
	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString((hash[:]))

	//store the user
	err = app.store.Users.CreateAndInvite(r.Context(), user, hashToken, app.config.mail.exp)
	if err != nil {
		switch err {
		case store.ErrorDuplicateEmail:
			app.badRequestResponse(w, r, err)
		case store.ErrorDuplicateEmail:
			app.badRequestResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := jsonResponse(w, http.StatusCreated, nil); err != nil {
		app.internalServerError(w, r, err)
	}

}
