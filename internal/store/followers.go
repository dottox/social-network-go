package store

import (
	"context"
	"database/sql"

	"github.com/dottox/social/internal/model"
	"github.com/lib/pq"
)

type FollowerStore struct {
	db *sql.DB
}

func (s *FollowerStore) Follow(ctx context.Context, unfollower *model.FollowAction) error {
	query := `
		INSERT INTO followers (user_id, follower_id)
		VALUES ($1, $2)
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(
		ctx,
		query,
		unfollower.TargetUserId,
		unfollower.SenderUserId,
	)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrResourceAlreadyExists
		}
	}

	return nil
}

func (s *FollowerStore) Unfollow(ctx context.Context, follower *model.FollowAction) error {
	query := `
		DELETE FROM followers
		WHERE user_id = $1 AND follower_id = $2 
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(
		ctx,
		query,
		follower.TargetUserId,
		follower.SenderUserId,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return ErrResourceNotFound
		default:
			return err
		}
	}

	return nil
}
