package main

import (
	"context"
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Shadowcyng/goSocial/docs" // This is required to generate a swagger docs
	"github.com/Shadowcyng/goSocial/internal/auth"
	"github.com/Shadowcyng/goSocial/internal/mailer"
	"github.com/Shadowcyng/goSocial/internal/ratelimiter"
	"github.com/Shadowcyng/goSocial/internal/store"
	"github.com/Shadowcyng/goSocial/internal/store/cache"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
)

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTimes string
}

type mailConfig struct {
	exp       time.Duration
	apiKey    string
	fromEmail string
}

type basicConfig struct {
	user string
	pass string
}
type authConfig struct {
	basic basicConfig
	token tokenConfig
}
type tokenConfig struct {
	secret string
	exp    time.Duration
	issuer string
}

type redisConfig struct {
	addr    string
	pw      string
	db      int
	enabled bool
}
type config struct {
	addr        string
	db          dbConfig
	env         string
	version     string
	apiURL      string
	mail        mailConfig
	frontendURL string
	auth        authConfig
	redis       redisConfig
	rateLimiter ratelimiter.Config
}

type application struct {
	config        config
	store         store.Storage
	logger        *zap.SugaredLogger
	mailer        mailer.Client
	authenticator auth.Authenticator
	cacheStorage  cache.Storage
	rateLimiter   ratelimiter.Limiter
}

type Role struct {
	Admin     string
	User      string
	Moderator string
}

var Roles = Role{Admin: "admin", User: "user", Moderator: "moderator"}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()
	// cors handling
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"},
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	r.Use(app.rateLimiterMiddleware)
	// set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that request has timeout and furthur
	// processing should be stopped
	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/v1", func(r chi.Router) {

		// operational end points
		// r.With(app.BasicAuthMiddleware()).Get("/health", app.healthCheckHandler)
		r.Get("/health", app.healthCheckHandler)
		r.With(app.BasicAuthMiddleware()).Get("/debug/vars", expvar.Handler().ServeHTTP)

		docsURL := fmt.Sprintf("%s/swagger/doc.json", app.config.addr)
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))
		r.Route("/posts", func(r chi.Router) {
			r.Use(app.AuthTokenMiddleware)
			r.Post("/", app.createPostHandler)
			r.Route("/{postID}", func(r chi.Router) {
				r.Use(app.postContextMiddleware)
				r.Get("/", app.getPostHandler)
				r.Patch("/", app.checkPostOwnership(Roles.Moderator, app.updatePostHandler))
				r.Delete("/", app.checkPostOwnership(Roles.Admin, app.deletePostHandler))
				r.Post("/commnets", app.createCommentHandler)
			})
		})
		r.Route("/users", func(r chi.Router) {
			r.Put("/activate/{token}", app.activateUserHandler)
			r.Route("/{userID}", func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)
				r.Use(app.userContextMiddleware)
				r.Get("/", app.getUserHandler)
				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unfollowUserHandler)
			})
		})
		r.Group(func(r chi.Router) {
			r.Use(app.AuthTokenMiddleware)
			r.Get("/feed", app.getUserFeedHandler)
		})

		// Public routes
		r.Route("/authentication", func(r chi.Router) {
			r.Post("/user", app.registerUserHandler)
			r.Post("/token", app.createTokenHandler)
		})
	})
	return r
}

func (app *application) run(mux http.Handler) error {
	// Docs
	docs.SwaggerInfo.Version = Version
	docs.SwaggerInfo.Host = app.config.apiURL
	docs.SwaggerInfo.BasePath = "/v1"
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	shutdown := make(chan error)
	go func() {
		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		app.logger.Infow("Shutting down server", "signal", s.String())
		shutdown <- srv.Shutdown(ctx)
	}()

	app.logger.Infow("Server has start at", "addr", app.config.addr, "env", app.config.env)
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	err = <-shutdown
	if err != nil {
		return err
	}
	app.logger.Infow("Server has shutdown", "addr", app.config.addr, "env", app.config.env)
	return nil
}
