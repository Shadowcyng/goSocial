package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrorNotFound          = errors.New("resource not found")
	ErrorConflict          = errors.New("resource already exists")
	QueryTimeout           = time.Second * 1000
	ErrorDuplicateEmail    = errors.New("email already exists")
	ErrorDuplicateUsername = errors.New("username already exists")
)

type Storage struct {
	Posts interface {
		Create(context.Context, *Post) error
		GetById(context.Context, int64) (*Post, error)
		DeleteById(context.Context, int64) error
		UpdatePostById(context.Context, *Post) error
		GetUserFeed(context.Context, int64, PaginatedFeedQuery) ([]*PostWithMetadata, error)
	}
	Users interface {
		Create(context.Context, *sql.Tx, *User) error
		Delete(context.Context, int64) error
		GetById(context.Context, int64) (*User, error)
		GetByEmail(context.Context, string) (*User, error)
		GetByUsername(context.Context, string) (*User, error)
		CreateAndInvite(context.Context, *User, string, time.Duration) error
		Activate(context.Context, string) error
	}
	Comments interface {
		GetByPostID(context.Context, int64) ([]Comment, error)
		Create(context.Context, *Comment) error
	}
	Followers interface {
		Follow(context.Context, int64, int64) error
		Unfollow(context.Context, int64, int64) error
	}
	Role interface {
		GetByName(ctx context.Context, roleName string) (*Role, error)
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts:     &PostStore{db: db},
		Users:     &UserStore{db: db},
		Comments:  &CommentStore{db: db},
		Followers: &FollowerStore{db: db},
		Role:      &RoleStore{db: db},
	}
}

func withTx(db *sql.DB, ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
