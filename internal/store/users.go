package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/dottox/social/internal/model"
)

type UserStore struct {
	db *sql.DB
}

func (s *UserStore) Create(ctx context.Context, tx *sql.Tx, user *model.User) error {
	query := `
		INSERT INTO users (username, email, password)
		VALUES ($1, $2, $3) RETURNING id, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := tx.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.Password.Hash,
	).Scan(
		&user.Id,
		&user.CreatedAt,
	)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrUsersDuplicateEmail
		case err.Error() == `pq: duplicate key value violates unique constraint "users_username_key"`:
			return ErrUsersDuplicateUsername
		default:
			return err
		}
	}

	return nil
}

func (s *UserStore) GetById(ctx context.Context, id uint32) (*model.User, error) {
	query := `
		SELECT id, username, email, password, created_at, is_active
		FROM users
		WHERE id = $1 AND is_active = true
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &model.User{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&user.Id,
		&user.Username,
		&user.Email,
		&user.Password.Hash,
		&user.CreatedAt,
		&user.IsActive,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrResourceNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, username, email, password, created_at
		FROM users
		WHERE email = $1 AND is_active = true
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &model.User{}
	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.Id,
		&user.Username,
		&user.Email,
		&user.Password.Hash,
		&user.CreatedAt,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrResourceNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}

func (s *UserStore) DeleteById(ctx context.Context, id uint32) error {
	query := `
		DELETE FROM users
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserStore) CreateAndInvite(ctx context.Context, user *model.User, token string, invitationExp time.Duration) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		if err := s.Create(ctx, tx, user); err != nil {
			return err
		}

		err := s.createUserInvitation(ctx, tx, token, invitationExp, user.Id)
		if err != nil {
			return err
		}

		return nil
	})
}

// Part of CreateAndInvite transaction
func (s *UserStore) createUserInvitation(ctx context.Context, tx *sql.Tx, token string, invitationExp time.Duration, userId uint32) error {
	query := `
		INSERT INTO user_invitations (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(
		ctx,
		query,
		userId,
		token,
		time.Now().Add(invitationExp),
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserStore) Activate(ctx context.Context, token string) error {
	// Within a transaction
	return withTx(s.db, ctx, func(tx *sql.Tx) error {

		// Get the user by the invitation token
		user, err := s.getUserByInvitationToken(ctx, tx, token)
		if err != nil {
			return err
		}

		// Update the user state to be active
		user.IsActive = true
		if err := s.updateUser(ctx, tx, user); err != nil {
			return err
		}

		// Delete the invitation token
		if err := s.deleteUserInvitation(ctx, tx, user.Id); err != nil {
			return err
		}

		return nil
	})
}

func (s *UserStore) getUserByInvitationToken(ctx context.Context, tx *sql.Tx, token string) (*model.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.password, u.created_at
		FROM users u
		INNER JOIN user_invitations ui ON ui.user_id = u.id
		WHERE ui.token = $1 AND ui.expires_at > $2
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &model.User{}
	err := tx.QueryRowContext(ctx, query, token, time.Now()).Scan(
		&user.Id,
		&user.Username,
		&user.Email,
		&user.Password.Hash,
		&user.CreatedAt,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrResourceNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}

func (s *UserStore) updateUser(ctx context.Context, tx *sql.Tx, user *model.User) error {
	query := `
		UPDATE users
		SET username = $1, email = $2, password = $3, is_active = $4
		WHERE id = $5
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.Password.Hash,
		user.IsActive,
		user.Id,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserStore) deleteUserInvitation(ctx context.Context, tx *sql.Tx, userId uint32) error {
	query := `
		DELETE FROM user_invitations
		WHERE user_id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, userId)
	if err != nil {
		return err
	}

	return nil
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
