package store

import (
	"context"
	"database/sql"
	"errors"
)

type RoleStore struct {
	db *sql.DB
}

type Role struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Level       int    `json:"level"`
	Description string `json:"description"`
}

func (s *RoleStore) GetByName(ctx context.Context, roleName string) (*Role, error) {
	query := `Select id, name, level, description FROM roles where name = $1`
	ctx, cancelCtx := context.WithTimeout(ctx, QueryTimeout)
	defer cancelCtx()
	var role Role
	err := s.db.QueryRowContext(ctx, query, roleName).Scan(
		&role.ID,
		&role.Name,
		&role.Level,
		&role.Description,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrorNotFound
		default:
			return nil, err
		}
	}
	return &role, nil
}
