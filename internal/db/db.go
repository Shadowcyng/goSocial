package db

import (
	"context"
	"database/sql"
	"time"
)

func New(addr string, maxOpenConns int, maxIdleConns int, maxIdleTime string) (*sql.DB, error) {
	db, err := sql.Open("postgres", addr)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	duration, er := time.ParseDuration(maxIdleTime)
	if er != nil {
		return nil, er
	}
	db.SetConnMaxIdleTime(duration)
	ctx, cancelCtx := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelCtx()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}
