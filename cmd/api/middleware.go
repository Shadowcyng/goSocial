package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Shadowcyng/goSocial/internal/store"
	"github.com/golang-jwt/jwt/v5"
)

func (app *application) BasicAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// read the auth header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				app.unauthorizedBasicError(w, r, fmt.Errorf("authorization header is missing"))
				return
			}
			// parse it -> get the base 64
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Basic" {
				app.unauthorizedBasicError(w, r, fmt.Errorf("authorization header is malformed"))
				return
			}
			// decode it
			decoded, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				app.unauthorizedBasicError(w, r, err)
				return
			}
			username := app.config.auth.basic.user
			pass := app.config.auth.basic.pass
			creds := strings.SplitN(string(decoded), ":", 2)
			if len(creds) != 2 || creds[0] != username || creds[1] != pass {
				app.unauthorizedBasicError(w, r, fmt.Errorf("invalid credentials"))
				return
			}
			// check credentials
			next.ServeHTTP(w, r)
		})
	}
}

func (app *application) AuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// read the auth header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.unauthorizedError(w, r, fmt.Errorf("authorization header is missing"))
			return
		}
		// parse it -> get the token
		parts := strings.Split(authHeader, " ") // authorization: Bearer <token>
		if len(parts) != 2 || parts[0] != "Bearer" {
			app.unauthorizedError(w, r, fmt.Errorf("authorization header is malformed"))
			return
		}
		// validate token
		token := parts[1]
		jwtToken, err := app.authenticator.ValidateToken(token)
		if err != nil {
			app.unauthorizedError(w, r, fmt.Errorf("invalid token"))
			return
		}

		// parse the claims
		claims, _ := jwtToken.Claims.(jwt.MapClaims)
		userID, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["sub"]), 10, 64)

		if err != nil {
			app.unauthorizedError(w, r, fmt.Errorf("invalid token"))
			return
		}

		ctx := r.Context()
		user, err := app.getUser(ctx, userID)
		if err != nil {
			switch err {
			case store.ErrorNotFound:
				app.unauthorizedError(w, r, fmt.Errorf("invalid token"))
				return
			default:
				app.internalServerError(w, r, err)
				return
			}
		}
		ctx = context.WithValue(ctx, authCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) checkPostOwnership(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getAuthUserFromContext(r)
		post := getPostFromContext(r)

		// if it is user's post
		if post.UserID == user.ID {
			next.ServeHTTP(w, r)
			return
		}
		// role precendece check
		allowed, err := app.checkRolePrecedence(r.Context(), user, requiredRole)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}
		if !allowed {
			app.forbiddenError(w, r, fmt.Errorf("user is not allowed to perform this action"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) checkRolePrecedence(ctx context.Context, user *store.User, roleName string) (bool, error) {
	role, err := app.store.Role.GetByName(ctx, roleName)
	if err != nil {
		return false, err
	}
	return user.Role.Level >= role.Level, nil
}

func (app *application) getUser(ctx context.Context, userID int64) (*store.User, error) {
	if !app.config.redis.enabled {
		return app.store.Users.GetById(ctx, userID)
	}
	app.logger.Infow("cache hit", "key", "user", "id", userID)
	user, err := app.cacheStorage.Users.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		app.logger.Infow("fetching from db", "id", userID)
		user, err := app.store.Users.GetById(ctx, userID)
		if err != nil {
			return nil, err
		}
		err = app.cacheStorage.Users.Set(ctx, user)
		if err != nil {
			return nil, err
		}
		return user, nil
	}
	return user, nil
}

func (app *application) rateLimiterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.rateLimiter.Enabled {
			if allow, retryAfter := app.rateLimiter.Allow(r.RemoteAddr); !allow {
				app.rateLimitExceededResponse(w, r, retryAfter.String())
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
