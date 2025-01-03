package store

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

type Follower struct {
	UserID     int64  `json:"user_id"`
	FollowerId int64  `json:"follower_id"`
	CreatedAt  string `json:"created_at"`
}

type FollowerStore struct {
	db *sql.DB
}

func (s *FollowerStore) Follow(ctx context.Context, followerId int64, userId int64) error {
	query := `INSERT INTO followers (
	user_id, follower_id
	) VALUES ($1, $2)`

	ctx, cancelCtx := context.WithTimeout(ctx, QueryTimeout)
	defer cancelCtx()

	_, err := s.db.ExecContext(ctx, query, userId, followerId)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrorConflict
		}
	}
	return nil
}

func (s *FollowerStore) Unfollow(ctx context.Context, followerId int64, userId int64) error {
	query := `DELETE FROM followers where follower_id = $1 and user_id = $2`
	ctx, cancelCtx := context.WithTimeout(ctx, QueryTimeout)
	defer cancelCtx()

	_, err := s.db.ExecContext(ctx, query, followerId, userId)
	if err != nil {
		return err
	}
	return nil
}
