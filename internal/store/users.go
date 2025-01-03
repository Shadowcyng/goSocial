package store

import (
	"context"
	"database/sql"
	"errors"
)

type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"-"`
	CreatedAt string `json:"created_at"`
}
type UserStore struct {
	db *sql.DB
}

func (s *UserStore) Create(ctx context.Context, user *User) error {
	query := `INSERT into users (username, password, email) 
	VALUES ($1, $2, $3) returning id, created_at
	`
	ctx, cancelCtx := context.WithTimeout(ctx, QueryTimeout)
	defer cancelCtx()
	err := s.db.QueryRowContext(ctx, query, user.Username, user.Password, user.Email).Scan(&user.ID, &user.CreatedAt)

	if err != nil {
		return err
	}

	return nil
}
func (s *UserStore) GetById(ctx context.Context, userId int64) (*User, error) {
	query := `Select id, username, email, password, created_at FROM users where id = $1 `

	ctx, cancelCtx := context.WithTimeout(ctx, QueryTimeout)
	defer cancelCtx()
	var user User
	err := s.db.QueryRowContext(ctx, query, userId).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrorNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}