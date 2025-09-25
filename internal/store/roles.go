package store

import (
	"context"
	"database/sql"

	"github.com/dottox/social/internal/model"
)

type RoleStore struct {
	db *sql.DB
}

func (s *RoleStore) GetByName(ctx context.Context, name string) (*model.Role, error) {
	query := `
		SELECT id, name, description, level
		FROM roles
		WHERE name = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	role := &model.Role{}
	err := s.db.QueryRowContext(ctx, query, name).Scan(
		&role.Id,
		&role.Name,
		&role.Description,
		&role.Level,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrResourceNotFound
		default:
			return nil, err
		}
	}

	return role, nil
}
