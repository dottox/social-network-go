package store

import (
	"context"
	"database/sql"
	"slices"
	"time"

	"github.com/dottox/social/internal/model"
)

func NewMockStore() *Storage {
	return &Storage{
		Users: &MockUserStore{},
		Followers: &MockFollowerStore{
			currentFollows: make(map[uint32][]uint32, 0),
		},
	}
}

// ----------------- USER STORE MOCK ---------------------
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

// ----------------- FOLLOWERS STORE MOCK ---------------------
type MockFollowerStore struct {
	currentFollows map[uint32][]uint32
}

func (m *MockFollowerStore) Follow(ctx context.Context, unfollower *model.FollowAction) error {
	sId := unfollower.SenderUserId
	tId := unfollower.TargetUserId

	if tId == 0 {
		return ErrResourceNotFound
	}

	if m.currentFollows[sId] == nil {
		m.currentFollows[sId] = make([]uint32, 0)
	}

	if slices.Contains(m.currentFollows[sId], tId) {
		return ErrResourceAlreadyExists
	}

	m.currentFollows[sId] = append(m.currentFollows[sId], tId)
	return nil
}

func (m *MockFollowerStore) Unfollow(ctx context.Context, follower *model.FollowAction) error {
	sId := follower.SenderUserId
	tId := follower.TargetUserId

	if tId == 0 {
		return ErrResourceNotFound
	}

	if m.currentFollows[sId] == nil {
		return ErrResourceNotFound
	}

	if !slices.Contains(m.currentFollows[sId], tId) {
		return ErrResourceNotFound
	}

	elemIdx := slices.Index(m.currentFollows[sId], tId)
	m.currentFollows[sId] = slices.Delete(m.currentFollows[sId], elemIdx, elemIdx)
	return nil
}
