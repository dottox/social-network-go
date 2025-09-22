package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/dottox/social/internal/model"
	"github.com/lib/pq"
)

type PostStore struct {
	db *sql.DB
}

func (s *PostStore) Create(ctx context.Context, post *model.Post) error {
	// Create the query to insert the post
	query := `
		INSERT INTO posts (title, content, user_id, tags)
		VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at, version
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	// Send the query with the context and arguments
	err := s.db.QueryRowContext(
		ctx,
		query,
		post.Title,
		post.Content,
		post.UserId,
		pq.Array(post.Tags),
	).Scan( // Scan the post to insert the generated values
		&post.Id,
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.Version,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostStore) GetById(ctx context.Context, id uint32) (*model.Post, error) {
	// Create the query to get the post by the id
	query := `
		SELECT id, title, content, user_id, tags, created_at, updated_at, comments_count, version
		FROM posts
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	// Create a new blank post
	var post model.Post

	// Perform the query with the ctx and id
	// Scan all the data to the blank post
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&post.Id,
		&post.Title,
		&post.Content,
		&post.UserId,
		pq.Array(&post.Tags), // note: tags is a slice, so use pq.Array()
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.CommentsCount,
		&post.Version,
	)
	if err != nil {
		switch {
		// If no rows found, return a ErrResourceNotFound
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrResourceNotFound
		default:
			return nil, err
		}
	}

	// Return the post without erors
	return &post, nil
}

func (s *PostStore) Update(ctx context.Context, post *model.Post) error {

	// Create the query to get the post by the id
	updateQuery := `
		UPDATE posts
		SET title = $1, content = $2, updated_at = NOW(), version = version + 1
		WHERE id = $3 AND version = $4
		RETURNING title, content, updated_at, version
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	// Perform the query with the ctx and id
	err := s.db.QueryRowContext(
		ctx,
		updateQuery,
		post.Title,
		post.Content,
		post.Id,
		post.Version,
	).Scan(
		&post.Title,
		&post.Content,
		&post.UpdatedAt,
		&post.Version,
	)
	if err != nil {
		switch {
		// If no rows found, return a ErrResourceNotFound
		case errors.Is(err, sql.ErrNoRows):
			return ErrResourceNotFound
		default:
			return err
		}
	}

	// Return no errors
	return err
}

func (s *PostStore) DeleteById(ctx context.Context, id uint32) error {

	// Create the query to delete the post by the id
	query := `
		DELETE FROM posts
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	// Perform the query with the ctx and id
	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		switch {
		// If no rows found, return a ErrResourceNotFound
		case errors.Is(err, sql.ErrNoRows):
			return ErrResourceNotFound
		default:
			return err
		}
	}

	// Return no errors
	return nil
}

func (s *PostStore) GetUserFeed(ctx context.Context, userId uint32, fq PaginatedFeedQuery) ([]*model.Post, error) {

	fmt.Printf("Paginated feed query params: %+v\n", fq)

	query := `
		SELECT DISTINCT p.id, p.title, p.content, p.user_id, p.tags, p.created_at, p.updated_at, p.comments_count, p.version
		FROM posts p
		JOIN followers f ON p.user_id = f.user_id
		WHERE 
		    (f.follower_id = $1 OR p.user_id = $1) AND
		    (p.title ILIKE '%' || $4 || '%' OR p.content ILIKE '%' || $4 || '%') AND
		    (p.tags @> $5 OR $5 = '{}')
		ORDER BY p.created_at ` + fq.Sort + `
		LIMIT $2 OFFSET $3
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	// Create a new list of posts
	posts := []*model.Post{}

	// Perform the query with the ctx and id
	// Scan all the data to the blank post
	rows, err := s.db.QueryContext(
		ctx,
		query,
		userId,
		fq.Limit,
		fq.Offset,
		fq.Search,
		pq.Array(fq.Tags),
	)
	if err != nil {
		switch {
		// If no rows found, return a ErrResourceNotFound
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrResourceNotFound
		default:
			return nil, err
		}
	}

	defer rows.Close()

	for rows.Next() {
		post := &model.Post{}
		err := rows.Scan(
			&post.Id,
			&post.Title,
			&post.Content,
			&post.UserId,
			pq.Array(&post.Tags), // note: tags is a slice, so use pq.Array())
			&post.CreatedAt,
			&post.UpdatedAt,
			&post.CommentsCount,
			&post.Version,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func postExists(ctx context.Context, db *sql.DB, postId uint32) (bool, error) {
	var exists bool

	query := `
		SELECT EXISTS (
			SELECT 1 FROM posts WHERE id = $1
		)
	`

	err := db.QueryRowContext(ctx, query, postId).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
