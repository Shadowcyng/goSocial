package main

import (
	"fmt"
	"log"

	"github.com/Shadowcyng/goSocial/internal/db"
	"github.com/Shadowcyng/goSocial/internal/env"
	"github.com/Shadowcyng/goSocial/internal/store"
)

const Version = "0.0.1"

//	@title			GoSocial API
//	@version		1.0
//	@description	API for GoSocial, A social network
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@BasePath					/v1
//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							headder
//	@name						authorization
//	@description

func main() {
	cfg := config{
		addr:    env.GetString("ADDR", ":8080"),
		env:     env.GetString("ENV", "dev"),
		version: env.GetString("VERSION", Version),
		apiURL:  env.GetString("EXTERNAL_URL", "localhost:8080"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://satyam:satyam123@localhost/social?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTimes: env.GetString("DB_MAX_IDLE_TIMES", "15m"),
		},
	}

	db, err := db.New(cfg.db.addr, cfg.db.maxOpenConns, cfg.db.maxIdleConns, cfg.db.maxIdleTimes)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()
	fmt.Println("database connection pool established")
	store := store.NewStorage(db)

	app := &application{
		config: cfg,
		store:  store,
	}

	mux := app.mount()
	log.Fatal(app.run(mux))
}
