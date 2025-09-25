package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/dottox/social/internal/model"
)

func NewMockStore() *Storage {
	return &Storage{
		Users: &MockUserStore{},
	}
}

type MockUserStore struct {
}

func (m *MockUserStore) Create(ctx context.Context, tx *sql.Tx, user *model.User) error {
	return nil
}

func (m *MockUserStore) GetById(ctx context.Context, id uint32) (*model.User, error) {
	if id == 0 {
		return nil, ErrResourceNotFound
	}
	return &model.User{
		Id: id,
	}, nil
}

func (m *MockUserStore) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	return &model.User{}, nil
}

func (m *MockUserStore) DeleteById(ctx context.Context, id uint32) error {
	return nil
}

func (m *MockUserStore) CreateAndInvite(ctx context.Context, user *model.User, baseUrl string, tokenTTL time.Duration) error {
	return nil
}

func (m *MockUserStore) Activate(ctx context.Context, token string) error {
	return nil
}
