package cache

import (
	"context"

	"github.com/Shadowcyng/goSocial/internal/store"
)

func NewMockCache() Storage {
	return Storage{
		Users: &MockUserStore{},
	}
}

type MockUserStore struct{}

func (m *MockUserStore) Get(ctx context.Context, userID int64) (*store.User, error) {
	return &store.User{}, nil
}

func (m *MockUserStore) Set(ctx context.Context, u *store.User) error {
	return nil
}

func (m *MockUserStore) Delete(ctx context.Context, userID int64) {
}
