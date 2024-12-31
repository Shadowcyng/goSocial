package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Shadowcyng/goSocial/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTimes string
}

type config struct {
	addr    string
	db      dbConfig
	env     string
	version string
}

type application struct {
	config config
	store  store.Storage
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)

	// set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that request has timeout and furthur
	// processing should be stopped
	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.healthCheckHandler)

		r.Route("/posts", func(r chi.Router) {
			r.Post("/", app.createPostHandler)
			r.Route("/{post_id}", func(r chi.Router) {
				r.Get("/", app.getPostHandler)
			})
		})
	})
	return r
}

func (app *application) run(mux http.Handler) error {
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}
	log.Printf("Server has start at %s", app.config.addr)
	return srv.ListenAndServe()
}
