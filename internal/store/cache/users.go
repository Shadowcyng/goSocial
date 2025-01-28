package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Shadowcyng/goSocial/internal/store"
	"github.com/go-redis/redis/v8"
)

type UserStore struct {
	rdb *redis.Client
}

const UserExpTime = time.Hour * 24 * 1 // 1days

func (s *UserStore) Get(ctx context.Context, userID int64) (*store.User, error) {
	cacheKey := fmt.Sprintf("user:%d", userID)
	user := &store.User{}
	data, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return user, err
	}
	if data != "" {
		if err := json.Unmarshal([]byte(data), user); err != nil {
			return user, err
		}
	}
	return user, nil
}
func (s *UserStore) Set(ctx context.Context, user *store.User) error {
	// TTL is 1 day
	if user == nil {
		return errors.New("user id is required")
	}
	cacheKey := fmt.Sprintf("user:%d", user.ID)
	json, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return s.rdb.SetEX(ctx, cacheKey, json, UserExpTime).Err()

}
func (s *UserStore) Delete(ctx context.Context, userID int64) {
	cacheKey := fmt.Sprintf("user-%d", userID)
	s.rdb.Del(ctx, cacheKey)
}
