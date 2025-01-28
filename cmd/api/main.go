package main

import (
	"log"
	"time"

	"github.com/Shadowcyng/goSocial/internal/auth"
	"github.com/Shadowcyng/goSocial/internal/db"
	"github.com/Shadowcyng/goSocial/internal/env"
	"github.com/Shadowcyng/goSocial/internal/mailer"
	"github.com/Shadowcyng/goSocial/internal/ratelimiter"
	"github.com/Shadowcyng/goSocial/internal/store"
	"github.com/Shadowcyng/goSocial/internal/store/cache"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
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

	// Config
	cfg := config{
		addr:        env.GetString("ADDR", ":8080"),
		env:         env.GetString("ENV", "dev"),
		version:     env.GetString("VERSION", Version),
		apiURL:      env.GetString("EXTERNAL_URL", "localhost:8080"),
		frontendURL: env.GetString("FRONTEND_URL", "localhost:4000"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://satyam:satyam123@localhost/social?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTimes: env.GetString("DB_MAX_IDLE_TIMES", "15m"),
		},
		mail: mailConfig{
			exp:       time.Hour * 24 * 3, // 3days
			apiKey:    env.GetString("MAIL_SERVICE_API_KEY", ""),
			fromEmail: env.GetString("FROM_EMAIL", ""),
		},
		auth: authConfig{
			basic: basicConfig{
				user: env.GetString("AUTH_BASIC_USER", "admin"),
				pass: env.GetString("AUTH_BASIC_PASS", "admin"),
			},
			token: tokenConfig{
				secret: env.GetString("JWT_SECRET", ""),
				exp:    (time.Hour * 24 * 3), // 3 days
				issuer: "GoSocial",
			},
		},
		redis: redisConfig{
			addr:    env.GetString("REDIS_ADDR", "localhost:6379"),
			pw:      env.GetString("REDIS_PW", "satyam123"),
			db:      env.GetInt("REDIS_DB", 0),
			enabled: env.GetBool("REDIS_ENABLED", false),
		},
		rateLimiter: ratelimiter.Config{
			RequestPerTimeFrame: env.GetInt("RATE_LIMITER_REQUEST_COUNT", 20),
			TimeFrame: time.Second*5,
			Enabled: env.GetBool("RATE_LIMITER_ENABLE", true),
		},
	}
	// Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	// Database
	db, err := db.New(
		cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTimes,
	)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()
	logger.Info("database connection pool established")

	// Cache
	var rdb *redis.Client
	if cfg.redis.enabled {
		rdb = cache.NewRedisClient(cfg.redis.addr, cfg.redis.pw, cfg.redis.db)
		logger.Info("redis connection  established")
	}

	// Store
	store := store.NewStorage(db)

	// cache
	cacheSotrage := cache.NewRedisStorage(rdb)
	mailer, err := mailer.NewMailerService(cfg.mail.apiKey, cfg.mail.fromEmail)
	if err != nil {
		log.Fatal(err)
	}

	// rate limiter
	rateLimiter := ratelimiter.NewFixedWindowLimiter(cfg.rateLimiter.RequestPerTimeFramem , cfg.rateLimiter.TimeFrame)

	jwtAuthenticator := auth.NewJWTAuthenticator(cfg.auth.token.secret, cfg.auth.token.issuer, cfg.auth.token.issuer)
	app := &application{
		config:        cfg,
		store:         store,
		logger:        logger,
		mailer:        mailer,
		authenticator: jwtAuthenticator,
		cacheStorage:  cacheSotrage,
		ratelimiter
	}

	mux := app.mount()
	logger.Fatal(app.run(mux))
}
