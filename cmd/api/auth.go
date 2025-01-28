package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/Shadowcyng/goSocial/internal/mailer"
	"github.com/Shadowcyng/goSocial/internal/store"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type authKey string

const authCtx authKey = "authUser"

type UserWithToken struct {
	*store.User
	Token string `json:"token"`
}
type RegisterUserPayload struct {
	Username string `json:"username" validate:"required,max=100"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

type CreateUserTokenPayload struct {
	Email    string `json:"email" validate:"omitempty,email,max=255"`
	Username string `json:"username" validate:"omitempty,min=3,max=50"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

// @Summary		user registeration
// @Description	register user using email, username and password
// @Tags			authentication
// @Accept			json
// @Produce		json
// @Param			payload	body		RegisterUserPayload	true	"User credentials"
// @Success		200		{object}	UserWithToken
// @Failure		400		{object}	error	"Bad request"
// @Failure		500		{object}	error	"Somehting went wrong"
// @security		ApiKeyAuth
// @Router			/authentication/user	[post]
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegisterUserPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.notFoundResponse(w, r, err)
		return
	}
	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &store.User{
		Username: payload.Username,
		Email:    payload.Email,
		Role:     store.Role{Name: "user"},
	}

	// hash the password
	err := user.Password.Set(payload.Password)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	plainToken := uuid.New().String()

	// store this in user_invitation tables after hash but keep plain token for email
	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString((hash[:]))

	//store the user
	err = app.store.Users.CreateAndInvite(r.Context(), user, hashToken, app.config.mail.exp)
	if err != nil {
		switch err {
		case store.ErrorDuplicateEmail:
			app.badRequestResponse(w, r, err)
		case store.ErrorDuplicateUsername:
			app.badRequestResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	userWithToken := UserWithToken{
		User:  user,
		Token: plainToken,
	}

	// send email
	activationURL := fmt.Sprintf("%s/confirm/%s", app.config.frontendURL, plainToken)
	isProdEnv := app.config.env == "production"
	vars := struct {
		Username      string
		ActivationURL string
	}{
		Username:      payload.Username,
		ActivationURL: activationURL,
	}
	err = app.mailer.Send(mailer.UserWelcomeTemplate, payload.Username, payload.Email, vars, !isProdEnv)
	if err != nil {
		app.logger.Errorw("error sending welcome email", "error", err)
		// rollback user creation if email fails (SAGA pattern)
		if err := app.store.Users.Delete(r.Context(), user.ID); err != nil {
			app.logger.Errorw("error deleting user", "error", err)
		}
		app.internalServerError(w, r, err)
	}
	app.logger.Infof("Email sent successfully to %s", payload.Email)
	if err := jsonResponse(w, http.StatusOK, userWithToken); err != nil {
		app.internalServerError(w, r, err)
	}

}

// @Summary		Creates a token
// @Description	Creates a token for user
// @Tags			authentication
// @Accept			json
// @Produce		json
// @Param			payload	body		CreateUserTokenPayload	true	"User credentials"
// @Success		200		{string}	string					"Token"
// @Failure		400		{object}	error					"Bad request"
// @Failure		500		{object}	error					"Somehting went wrong"
// @security		ApiKeyAuth
// @Router			/authentication/token	[post]
func (app *application) createTokenHandler(w http.ResponseWriter, r *http.Request) {
	// parse payload credentials
	var payload CreateUserTokenPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if payload.Email == "" && payload.Username == "" {
		app.badRequestResponse(w, r, fmt.Errorf("email or username is required"))
		return
	}

	// fetch the user (check if user exist) from the payload
	var user *store.User
	var err error
	if payload.Email != "" {
		user, err = app.store.Users.GetByEmail(r.Context(), payload.Email)
	} else {
		user, err = app.store.Users.GetByUsername(r.Context(), payload.Username)
	}
	if err != nil {
		switch err {
		case store.ErrorNotFound:
			app.unauthorizedError(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	err = user.Password.Validate(payload.Password)
	if err != nil {
		app.invalidCredentials(w, r, err)
		return
	}

	// generate the token -> add claims
	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(app.config.auth.token.exp).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": app.config.auth.token.issuer,
		"aud": app.config.auth.token.issuer,
	}

	token, err := app.authenticator.GenerateToken(claims)
	if err != nil {
		app.internalServerError(w, r, err)
		return

	}
	// send it to the client
	if err := jsonResponse(w, http.StatusCreated, token); err != nil {
		app.internalServerError(w, r, err)
	}
}

func getAuthUserFromContext(r *http.Request) *store.User {
	user, _ := r.Context().Value(authCtx).(*store.User)
	return user
}
