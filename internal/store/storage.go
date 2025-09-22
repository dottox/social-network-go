package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/dottox/social/internal/model"
)

var (
	ErrResourceNotFound      = errors.New("resource not found")
	ErrResourceAlreadyExists = errors.New("resource already exists")
	QueryTimeoutDuration     = 5 * time.Second
)

type Storage struct {
	Posts interface {
		Create(context.Context, *model.Post) error
		GetById(context.Context, uint32) (*model.Post, error)
		Update(context.Context, *model.Post) error
		DeleteById(context.Context, uint32) error
		GetUserFeed(context.Context, uint32, PaginatedFeedQuery) ([]*model.Post, error)
	}
	Users interface {
		Create(context.Context, *model.User) error
		GetById(context.Context, uint32) (*model.User, error)
	}
	Comments interface {
		Create(context.Context, *model.Comment) error
		GetAllByPostId(context.Context, uint32) ([]*model.Comment, error)
	}
	Followers interface {
		Follow(context.Context, *model.FollowAction) error
		Unfollow(context.Context, *model.FollowAction) error
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts:     &PostStore{db},
		Users:     &UserStore{db},
		Comments:  &CommentStore{db},
		Followers: &FollowerStore{db},
	}
}
